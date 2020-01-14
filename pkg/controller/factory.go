package controller

import (
	"context"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/openshift/library-go/pkg/operator/events"
)

// SyncFunc is a function that contain main controller logic.
// The ctx.ctx passed is the main controller ctx, when cancelled it means the controller is being shut down.
// The ctx provides access to controller name, queue and event recorder.
type SyncFunc func(ctx context.Context, controllerContext Context) error

// Factory is generator that generate standard Kubernetes controllers.
// Factory is really generic and should be only used for simple controllers that does not require special stuff..
type Factory struct {
	sync         SyncFunc
	informers    []cache.SharedInformer
	cachesToSync []cache.InformerSynced
}

// NewFactory return new factory instance.
func NewFactory() *Factory {
	return &Factory{}
}

// Sync is used to set the controller synchronization function. This function is the core of the controller and is
// usually hold the main controller logic.
func (f *Factory) Sync(syncFn SyncFunc) *Factory {
	f.sync = syncFn
	return f
}

// Informers is used to register event handlers and get the caches synchronized functions.
// Pass informers you want to use to react to changes on resources. If informer event is observed, then the Sync() function
// is called.
func (f *Factory) Informers(informers ...cache.SharedInformer) *Factory {
	f.informers = informers
	return f
}

// Controller produce a runnable controller.
func (f *Factory) Controller(name string, eventRecorder events.Recorder) Controller {
	if f.sync == nil {
		panic("Sync function must be set")
	}
	c := &baseController{
		sync: f.sync,
		ctx: controllerContext{
			controllerName: name,
			eventRecorder:  eventRecorder.WithComponentSuffix(name),
			queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name),
		},
	}
	for i := range f.informers {
		f.informers[i].AddEventHandler(c.ctx.getEventHandler())
		c.cachesToSync = append(f.cachesToSync, f.informers[i].HasSynced)
	}
	return c
}

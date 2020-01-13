package controller

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/openshift/library-go/pkg/operator/events"
)

// Controller interface represents a runnable Kubernetes controller.
// Cancelling the context passed will cause the controller to shutdown.
// Number of workers determine how much parallel the job processing should be.
type Controller interface {
	Run(ctx context.Context, workers int)
}

// Context provide access to controller name, queue and event recorder.
type Context struct {
	Name          string
	Queue         workqueue.RateLimitingInterface
	EventRecorder events.Recorder
}

// getEventHandler provides default event handler that is added to an informers passed to controller factory.
func (c *Context) getEventHandler() cache.ResourceEventHandler {
	staticPodStateControllerWorkQueueKey := fmt.Sprintf("%sKey", c.Name)
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.Queue.Add(staticPodStateControllerWorkQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.Queue.Add(staticPodStateControllerWorkQueueKey) },
		DeleteFunc: func(obj interface{}) { c.Queue.Add(staticPodStateControllerWorkQueueKey) },
	}
}

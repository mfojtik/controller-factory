package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/openshift/library-go/pkg/operator/events"
)

// ctx provide access to controller name, queue and event recorder.
type controllerContext struct {
	queue          workqueue.RateLimitingInterface
	eventRecorder  events.Recorder
	controllerName string

	// queueObject holds the object we got from informer
	// There is no direct access to this object to prevent cache mutation.
	queueObject runtime.Object
}

var _ Context = controllerContext{}

func (c controllerContext) GetQueueObject() runtime.Object {
	if c.queueObject == nil {
		return nil
	}
	return c.queueObject.DeepCopyObject()
}

func (c controllerContext) Queue() workqueue.RateLimitingInterface {
	return c.queue
}

func (c controllerContext) Events() events.Recorder {
	return c.eventRecorder
}

func (c controllerContext) ControllerName() string {
	return c.controllerName
}

// GetObjectMeta return metadata of object we observed change to via informer.
// If the object is not set, it returns nil.
func (c controllerContext) GetObjectMeta() metav1.Object {
	if c.queueObject == nil {
		return nil
	}
	metaObj, err := meta.Accessor(c.queueObject)
	if err != nil {
		return nil
	}
	return metaObj
}

// withQueueObject makes a copy of original ctx and return new ctx that has sync object set.
func (c controllerContext) withQueueObject(obj runtime.Object) controllerContext {
	return controllerContext{
		controllerName: c.ControllerName(),
		eventRecorder:  c.Events(),
		queue:          c.Queue(),
		queueObject:    obj,
	}
}

// getEventHandler provides default event handler that is added to an informers passed to controller factory.
func (c *controllerContext) getEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			runtimeObj, ok := obj.(runtime.Object)
			if !ok {
				utilruntime.HandleError(fmt.Errorf("added object %+v is not runtime Object", obj))
				return
			}
			c.queue.Add(runtimeObj)
		},
		UpdateFunc: func(old, new interface{}) {
			runtimeObj, ok := new.(runtime.Object)
			if !ok {
				utilruntime.HandleError(fmt.Errorf("updated object %+v is not runtime Object", runtimeObj))
				return
			}
			c.queue.Add(runtimeObj)
		},
		DeleteFunc: func(obj interface{}) {
			runtimeObj, ok := obj.(runtime.Object)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if ok {
					c.queue.Add(tombstone.Obj.(runtime.Object))
					return
				}
				utilruntime.HandleError(fmt.Errorf("updated object %+v is not runtime Object", runtimeObj))
				return
			}
			c.queue.Add(runtimeObj)
		},
	}
}

package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"

	"github.com/openshift/library-go/pkg/operator/events"
)

// Controller interface represents a runnable Kubernetes controller.
// Cancelling the ctx passed will cause the controller to shutdown.
// Number of workers determine how much parallel the job processing should be.
type Controller interface {
	// Run runs the controller and blocks until the controller is finished.
	// Number of workers can be specified via workers parameter.
	// Note that having more than one worker usually means handing parallelization of Sync().
	Run(ctx context.Context, workers int)

	// ShutdownContext can be used to observe the finished shutdown of all controller workers and controller itself.
	// Example: <-controller.ShutdownContext().Done()
	ShutdownContext() context.Context
}

// Context interface represents a context given to the Sync() function where the main controller logic happen.
// Context exposes controller name and give user access to the queue (for manual requeue).
// Context also provides metadata about object that informers observed as changed.
type Context interface {
	// Queue gives access to controller queue. This can be used for manual requeue, although if a Sync() function return
	// an error, the object is automatically re-queued. Use with caution.
	Queue() workqueue.RateLimitingInterface

	// GetObjectMeta provides access to currently synced object metadata.
	GetObjectMeta() metav1.Object

	// GetQueueObject provides access to current synced object deep copy.
	// It is safe to mutate this object inside Sync().
	GetQueueObject() runtime.Object

	// Events provide access to event recorder.
	Events() events.Recorder

	// ControllerName gives name of the controller.
	ControllerName() string
}

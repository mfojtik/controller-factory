[![GoDoc](https://godoc.org/github.com/mfojtik/controller-factory/pkg/controller?status.svg)](https://godoc.org/github.com/mfojtik/controller-factory/pkg/controller)

# controller-factory

This is a prototype for a generic [Kubernetes](https://github.com/kubernetes/kubernetes) controller generator that wraps all the boiler plate and let you
focus straight on the main synchronization logic (the `Sync()` function). 

The boiler plate used by this generator will provide:

* Proper handling of controller shutdown (terminating all workers, signalling to Sync(), etc..)
* Automatic registration of event handlers to all informers
* Automatic wait for cache sync for every informer
* Single work queue mechanism and context for `Sync()` function that provide object metadata

The result is very simple Kubernetes controller that reacts to resource changes from passed informers or resync periodically if `.ResyncEvery()` is used.
In many cases this is enough, like writing a simple operator controller loop, but in some cases it is not, such as:

* If you need to handle Create, Update or Delete events differently - this is not for you
* If your controller require multiple queues and custom requeue mechanism - this is not for you

This this case, please use the [controller-runtime](https://github.com/kubernetes/controller-runtime) library that gives you low level access to all above.

An example Kubernetes controller:

```go
type SampleController struct {}

func NewController(secretsInformer v1.SecretInformer) Controller {
    // Multiple informers can be registered here, they all get proper event handlers.
    // Note that you can also use .ResyncEvery(duration) to call Sync() periodically, regardless of informers.
    factory := NewFactory().Informers(secretsInformer.Informer())
    controller := &SampleController{}

    // The "Sync()" function must be called prior to producing the Controller().
    return factory.Sync(controller.Sync).Controller("SampleController", events.NewInMemoryRecorder("sample-controller"))
}

func (f *SampleController) Sync(ctx context.Context, controllerContext Context) error {
    // ctx.Err() != nil means the controller is being terminated.
    // controllerContext provide ControllerName() = "SampleController", Queue() = so you can requeue faster, EventRecorder() to record events.
    // controllerContext also provides QueueObject() and GetObjectMeta() to get access to object that caused the Sync() to run.
    
    // This code will run when a secret is created, updated or deleted.

    // Returning error here means the controllerContext.QueueObject() will be re-queued.
    return nil
}

func Start(ctx context.Context) {
    // ... 
    controller := NewFakeController(kubeInformers.Core().V1().Secrets())

    // The controller will start shutdown when the context is cancelled.
    // Number of workers specify how much parallel the Sync() will be. Use with caution, one worker is usually enough.
    go controller.Run(ctx, 1)

    // Wait for the controller to finish shutdown.
    <-controller.ShutdownContext().Done()
}
``` 

This looks similar to any other controller mechanism, except you don't have to deal with workers, queues, event handler registration or graceful shutdown.

# controller-factory

This is a prototype for a generic [Kubernetes](https://github.com/kubernetes/kubernetes) controller generator that wraps all the boilerplate and let you
focus straight on the synchronization logic. The result is very simple controller that react to resource changes. In many cases this is enough, but in many
cases it is not, such as:

* if you need to handle Create, Update or Delete differently (different event handlers) - this is not for you
* if you need to pass the resource namespace/name via queue - this is not for you

An example Kubernetes controller produced by this factory is:

```go
type SampleController struct {}

func NewController(secretsInformer v1.SecretInformer) Controller {
	factory := NewFactory().Informers(secretsInformer.Informer())
	controller := &SampleController{}
	return factory.Sync(controller.Sync).Controller("SampleController", events.NewInMemoryRecorder("sample-controller"))
}

func (f *SampleController) Sync(ctx context.Context, controllerContext Context) error {
    // ctx.Err() != nil means the controller is being shut down
    // controllerContext provide Name = "SampleController", Queue = so you can requeue faster, EventRecorder
    
    // this code will run when a secret is create/updated or deleted.
	return nil
}

func Start() {
    // ...
	controller := NewFakeController(kubeInformers.Core().V1().Secrets())
	go controller.Run(ctx, 1)
}
``` 


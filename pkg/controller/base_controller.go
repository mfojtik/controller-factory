package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// baseController represents generic Kubernetes controller boiler-plate
type baseController struct {
	cachesToSync    []cache.InformerSynced
	sync            func(ctx context.Context, controllerContext Context) error
	ctx             controllerContext
	shutdownContext context.Context
}

var _ Controller = &baseController{}

func (c *baseController) Run(ctx context.Context, workers int) {
	shutdownContext, shutdownComplete := context.WithCancel(context.Background())
	c.shutdownContext = shutdownContext

	defer shutdownComplete()
	defer utilruntime.HandleCrash()
	defer c.ctx.Queue().ShutDown()
	defer klog.Infof("Shutting down %s ...", c.ctx.ControllerName())

	klog.Infof("Starting %s ...", c.ctx.ControllerName())
	if !cache.WaitForCacheSync(ctx.Done(), c.cachesToSync...) {
		return
	}
	klog.V(5).Infof("Caches synced for controller %s", c.ctx.ControllerName())

	var workerWaitGroup sync.WaitGroup

	for i := 1; i <= workers; i++ {
		klog.Infof("Starting #%d worker of %s controller ...", i, c.ctx.ControllerName())
		workerWaitGroup.Add(1)
		go wait.UntilWithContext(ctx, func(ctx context.Context) {
			defer workerWaitGroup.Done()
			defer klog.Infof("Shutting down worker of %s controller ...", c.ctx.ControllerName())
			c.runWorker(ctx)
		}, time.Second)
	}

	// wait for controller shutdown to be requested
	<-ctx.Done()

	// wait for all workers to finish their jobs
	workerWaitGroup.Wait()
}

func (c *baseController) ShutdownContext() context.Context {
	return c.shutdownContext
}

func (c *baseController) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *baseController) processNextWorkItem(ctx context.Context) bool {
	syncObject, quit := c.ctx.Queue().Get()
	if quit || ctx.Err() != nil {
		return false
	}

	runtimeObj, ok := syncObject.(runtime.Object)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("object is not runtime Object: %+v", syncObject))
		c.ctx.Queue().Forget(syncObject)
		return true
	}

	defer c.ctx.Queue().Done(runtimeObj)

	if err := c.sync(ctx, c.ctx.withQueueObject(runtimeObj)); err != nil {
		utilruntime.HandleError(fmt.Errorf("%v failed with : %v", syncObject, err))
		if !c.ctx.Queue().ShuttingDown() && ctx.Err() == nil {
			c.ctx.Queue().AddRateLimited(runtimeObj)
		}
	}
	c.ctx.Queue().Forget(runtimeObj)
	return true
}

package controller

import (
	"context"
	"fmt"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// baseController represents generic Kubernetes controller boiler-plate
type baseController struct {
	cachesToSync      []cache.InformerSynced
	sync              func(ctx context.Context, controllerContext Context) error
	controllerContext Context
}

func (b *baseController) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()
	defer b.controllerContext.Queue.ShutDown()
	defer klog.Infof("Shutting down %s ...", b.controllerContext.Name)

	klog.Infof("Starting %s ...", b.controllerContext.Name)
	if !cache.WaitForCacheSync(ctx.Done(), b.cachesToSync...) {
		return
	}

	// doesn't matter what workers say, only start one.
	for i := 1; i <= workers; i++ {
		klog.Infof("Starting #%d worker of %s controller ...", i, b.controllerContext.Name)
		go wait.UntilWithContext(ctx, b.runWorker, time.Second)
	}

	<-ctx.Done()
}

func (b *baseController) runWorker(ctx context.Context) {
	for b.processNextWorkItem(ctx) {
	}
}

func (b *baseController) processNextWorkItem(ctx context.Context) bool {
	dsKey, quit := b.controllerContext.Queue.Get()
	if quit || ctx.Err() != nil {
		return false
	}

	defer b.controllerContext.Queue.Done(dsKey)

	if err := b.sync(ctx, b.controllerContext); err != nil {
		utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
		if !b.controllerContext.Queue.ShuttingDown() && ctx.Err() == nil {
			b.controllerContext.Queue.AddRateLimited(dsKey)
		}
	}
	b.controllerContext.Queue.Forget(dsKey)
	return true
}

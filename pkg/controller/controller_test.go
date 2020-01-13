package controller

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	v12 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/openshift/library-go/pkg/operator/events"
)

func makeFakeSecret() *v1.Secret {
	return &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test",
		},
		Data: map[string][]byte{
			"test": {},
		},
	}
}

type FakeController struct {
	synced chan struct{}
	t      *testing.T
}

func NewFakeController(t *testing.T, synced chan struct{}, secretsInformer v12.SecretInformer) Controller {
	factory := NewFactory().Informers(secretsInformer.Informer())
	controller := &FakeController{synced: synced, t: t}
	return factory.Sync(controller.Sync).Controller("FakeController", events.NewInMemoryRecorder("fake-controller"))
}

func (f *FakeController) Sync(ctx context.Context, controllerContext Context) error {
	defer close(f.synced)
	if ctx.Err() != nil {
		f.t.Logf("context %v", ctx.Err())
		return ctx.Err()
	}
	f.t.Logf("controller %s sync called", controllerContext.Name)
	return nil
}

func TestEmbeddedController(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()

	kubeInformers := informers.NewSharedInformerFactoryWithOptions(kubeClient, 1*time.Minute, informers.WithNamespace("test"))
	ctx, cancel := context.WithCancel(context.TODO())

	go kubeInformers.Start(ctx.Done())

	controllerSynced := make(chan struct{})
	controller := NewFakeController(t, controllerSynced, kubeInformers.Core().V1().Secrets())
	go controller.Run(ctx, 1)

	if _, err := kubeClient.CoreV1().Secrets("test").Create(makeFakeSecret()); err != nil {
		t.Fatalf("failed to create fake secret: %v", err)
	}

	select {
	case <-controllerSynced:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("test timeout")
	}
}

func TestSimpleController(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()

	kubeInformers := informers.NewSharedInformerFactoryWithOptions(kubeClient, 1*time.Minute, informers.WithNamespace("test"))
	ctx, cancel := context.WithCancel(context.TODO())

	go kubeInformers.Start(ctx.Done())
	factory := NewFactory().Informers(kubeInformers.Core().V1().Secrets().Informer())

	controllerSynced := make(chan struct{})
	controller := factory.Sync(func(ctx context.Context, controllerContext Context) error {
		defer close(controllerSynced)
		t.Logf("controller %s sync called", controllerContext.Name)
		return nil
	}).Controller("FakeController", events.NewInMemoryRecorder("fake-controller"))

	go controller.Run(ctx, 1)

	if _, err := kubeClient.CoreV1().Secrets("test").Create(makeFakeSecret()); err != nil {
		t.Fatalf("failed to create fake secret: %v", err)
	}

	select {
	case <-controllerSynced:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("test timeout")
	}
}

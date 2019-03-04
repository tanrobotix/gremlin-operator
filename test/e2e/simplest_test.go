package e2e

import (
	goctx "context"
	"fmt"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Kubedex/gremlin-operator/pkg/apis/gremlin/v1alpha1"
)

func SimplestGremlint *testing.T) {
	t.Parallel()
	ctx := prepare(t)
	defer ctx.Cleanup()

	if err := simplest(t, framework.Global, ctx); err != nil {
		t.Fatal(err)
	}
}

func simplest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	// create gremlin custom resource
	exampleGremlin := &v1alpha1.Gremlin{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gremlin",
			APIVersion: "gremlin.kubedex.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-gremlin",
			Namespace: namespace,
		},
		Spec: v1alpha1.GremlinSpec{},
	}
	err = f.Client.Create(goctx.TODO(), exampleGremlin, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		return err
	}

	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "my-gremlin", 1, retryInterval, timeout)
}

package stub

import (
	"log"

	fakeservingv1beta1 "github.com/kserve/kserve/pkg/client/clientset/versioned/fake"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

func NewFakeClient() dynamic.Interface {
	scheme := runtime.NewScheme()
	if err := fakeservingv1beta1.AddToScheme(scheme); err != nil {
		log.Fatal(err)
	}
	fakeservingv1beta1.NewSimpleClientset()
	return fake.NewSimpleDynamicClient(scheme)
}

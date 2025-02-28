package rhoai_normalizer

import (
	"context"
	"fmt"
	serverapiv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/kubeflowmodelregistry"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/common"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/kfmr"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/location"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
	"testing"
)

func TestReconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = serverapiv1beta1.AddToScheme(scheme)
	kts1 := kfmr.CreateGetServerWithInference(t)
	defer kts1.Close()
	kts2 := kfmr.CreateGetServer(t)
	defer kts2.Close()
	brts := location.CreateBridgeLocationServer(t)
	defer brts.Close()
	callback := sync.Map{}
	bsts := storage.CreateBridgeStorageRESTClient(t, &callback)
	defer bsts.Close()

	r := &RHOAINormalizerReconcile{
		scheme:        scheme,
		eventRecorder: nil,
		k8sToken:      "",
		myNS:          "",
		routeClient:   nil,
		storage:       storage.SetupBridgeStorageRESTClient(bsts),
	}

	for _, tc := range []struct {
		name          string
		is            *serverapiv1beta1.InferenceService
		route         *routev1.Route
		kfmrSvr       *httptest.Server
		expectedValue string
	}{
		{
			name: "kserve inference service without kubeflow route",
			is: &serverapiv1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec:       serverapiv1beta1.InferenceServiceSpec{},
				Status:     serverapiv1beta1.InferenceServiceStatus{},
			},
			expectedValue: "description: KServe instance foo:bar",
		},
		{
			name: "kserve inference service with kubeflow route but not kubeflow inference service",
			route: &routev1.Route{
				Spec: routev1.RouteSpec{
					Host: "http://foo.com",
				},
				Status: routev1.RouteStatus{Ingress: []routev1.RouteIngress{{}}},
			},
			is: &serverapiv1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{Namespace: "faa", Name: "bor"},
				Spec:       serverapiv1beta1.InferenceServiceSpec{},
				Status:     serverapiv1beta1.InferenceServiceStatus{},
			},
			kfmrSvr:       kts2,
			expectedValue: "description: KServe instance faa:bor",
		},
		{
			name: "kserve inference service with kubeflow route and kubeflow inference service",
			route: &routev1.Route{
				Spec: routev1.RouteSpec{
					Host: "http://foo.com",
				},
				Status: routev1.RouteStatus{Ingress: []routev1.RouteIngress{{}}},
			},
			is: &serverapiv1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "mnist-v1", Namespace: "ggmtest"},
				Spec:       serverapiv1beta1.InferenceServiceSpec{},
				Status:     serverapiv1beta1.InferenceServiceStatus{},
			},
			kfmrSvr:       kts1,
			expectedValue: "url: https://huggingface.co/tarilabs/mnist/resolve/v20231206163028/mnist.onnx",
		},
	} {
		ctx := context.TODO()
		objs := []client.Object{tc.is}
		if tc.kfmrSvr != nil {
			cfg := &config.Config{}
			kfmr.SetupKubeflowTestRESTClient(tc.kfmrSvr, cfg)
			r.kfmr = kubeflowmodelregistry.SetupKubeflowRESTClient(cfg)
		}
		r.client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
		r.kfmrRoute = tc.route
		r.Reconcile(ctx, reconcile.Request{types.NamespacedName{Namespace: tc.is.Namespace, Name: tc.is.Name}})
		found := false
		callback.Range(func(key, value any) bool {
			found = true
			t.Logf(fmt.Sprintf("found key %s for test %s", key, tc.name))
			postStr, ok := value.(string)
			common.AssertEqual(t, ok, true)
			common.AssertContains(t, postStr, []string{tc.expectedValue})

			return true
		})
		common.AssertEqual(t, found, true)
	}
}

func TestStart(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = serverapiv1beta1.AddToScheme(scheme)
	kts1 := kfmr.CreateGetServerWithInference(t)
	defer kts1.Close()
	kts2 := kfmr.CreateGetServer(t)
	defer kts2.Close()
	brts := location.CreateBridgeLocationServer(t)
	defer brts.Close()
	callback := sync.Map{}
	bsts := storage.CreateBridgeStorageRESTClient(t, &callback)
	defer bsts.Close()

	r := &RHOAINormalizerReconcile{
		scheme:        scheme,
		eventRecorder: nil,
		k8sToken:      "",
		myNS:          "",
		routeClient:   nil,
		kfmrRoute: &routev1.Route{
			Spec: routev1.RouteSpec{
				Host: "http://foo.com",
			},
			Status: routev1.RouteStatus{Ingress: []routev1.RouteIngress{{}}},
		},
		storage: storage.SetupBridgeStorageRESTClient(bsts),
	}

	for _, tc := range []struct {
		name          string
		is            *serverapiv1beta1.InferenceService
		kfmrSvr       *httptest.Server
		expectedKey   string
		expectedValue string
	}{
		{
			name: "not deployed, only registered model, model version, model artifact",
			is: &serverapiv1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{Namespace: "faa", Name: "bor"},
				Spec:       serverapiv1beta1.InferenceServiceSpec{},
				Status:     serverapiv1beta1.InferenceServiceStatus{},
			},
			kfmrSvr:       kts2,
			expectedKey:   "/model-1/v1/catalog-info.yaml",
			expectedValue: "description: dummy model 1",
		},
		{
			name: "deployed, with inference_service and serving_environments added",
			is: &serverapiv1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "mnist-v1", Namespace: "ggmtest"},
				Spec:       serverapiv1beta1.InferenceServiceSpec{},
				Status:     serverapiv1beta1.InferenceServiceStatus{},
			},
			kfmrSvr:       kts1,
			expectedKey:   "/mnist/v1/catalog-info.yaml",
			expectedValue: "url: https://huggingface.co/tarilabs/mnist/resolve/v20231206163028/mnist.onnx",
		},
	} {
		ctx := context.TODO()
		objs := []client.Object{tc.is}
		if tc.kfmrSvr != nil {
			cfg := &config.Config{}
			kfmr.SetupKubeflowTestRESTClient(tc.kfmrSvr, cfg)
			r.kfmr = kubeflowmodelregistry.SetupKubeflowRESTClient(cfg)
		}
		r.client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()

		r.innerStart(ctx)

		found := false
		callback.Range(func(key, value any) bool {
			found = true
			t.Logf(fmt.Sprintf("found key %s for test %s", key, tc.name))
			postStr, ok := value.(string)
			common.AssertEqual(t, ok, true)
			common.AssertContains(t, postStr, []string{tc.expectedValue})

			return true
		})
		common.AssertEqual(t, found, true)
	}

}

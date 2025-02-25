package rhoai_normalizer

import (
	"context"
	"fmt"
	serverapiv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/kubeflowmodelregistry"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/server/gin-gonic-http/server"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
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
	bts := stub.CreateServer(t)
	defer bts.Close()
	kts1 := stub.CreateGetServerWithInference(t)
	defer kts1.Close()
	kts2 := stub.CreateGetServer(t)
	defer kts2.Close()
	brts := stub.CreateBridgeLocationServer(t)
	defer brts.Close()

	r := &RHOAINormalizerReconcile{
		scheme:           scheme,
		eventRecorder:    nil,
		k8sToken:         "",
		locationRouteURL: "",
		myNS:             "",
		routeClient:      nil,
		bkstg:            backstage.SetupBackstageTestRESTClient(bts),
		brdgImport:       stub.SetupBridgeLocationRESTClient(brts),
	}

	for _, tc := range []struct {
		name          string
		is            *serverapiv1beta1.InferenceService
		route         *routev1.Route
		kfmrSvr       *httptest.Server
		expectedKey   string
		expectedValue string
	}{
		{
			name: "kserve inference service without kubeflow route",
			is: &serverapiv1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec:       serverapiv1beta1.InferenceServiceSpec{},
				Status:     serverapiv1beta1.InferenceServiceStatus{},
			},
			expectedKey:   "/foo/bar/catalog-info.yaml",
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
			expectedKey:   "/faa/bor/catalog-info.yaml",
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
			expectedKey:   "/mnist/v1/catalog-info.yaml",
			expectedValue: "url: https://huggingface.co/tarilabs/mnist/resolve/v20231206163028/mnist.onnx",
		},
	} {
		ctx := context.TODO()
		objs := []client.Object{tc.is}
		r.pushedLocations = sync.Map{}
		if tc.kfmrSvr != nil {
			cfg := &config.Config{}
			stub.SetupKubeflowTestRESTClient(tc.kfmrSvr, cfg)
			r.kfmr = kubeflowmodelregistry.SetupKubeflowRESTClient(cfg)
		}
		r.client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
		r.kfmrRoute = tc.route
		r.Reconcile(ctx, reconcile.Request{types.NamespacedName{Namespace: tc.is.Namespace, Name: tc.is.Name}})
		found := false
		r.pushedLocations.Range(func(key, value any) bool {
			t.Log(fmt.Sprintf("found key %s for test %s", key, tc.name))
			if key == tc.expectedKey {
				found = true
			}
			if found {
				postBody, ok := value.(*server.PostBody)
				stub.AssertEqual(t, ok, true)
				pb := string(postBody.Body)
				stub.AssertContains(t, pb, []string{tc.expectedValue})
			}

			return true
		})
		stub.AssertEqual(t, found, true)
	}
}

func TestStart(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = serverapiv1beta1.AddToScheme(scheme)
	bts := stub.CreateServer(t)
	defer bts.Close()
	kts1 := stub.CreateGetServerWithInference(t)
	defer kts1.Close()
	kts2 := stub.CreateGetServer(t)
	defer kts2.Close()
	brts := stub.CreateBridgeLocationServer(t)
	defer brts.Close()

	r := &RHOAINormalizerReconcile{
		scheme:           scheme,
		eventRecorder:    nil,
		k8sToken:         "",
		locationRouteURL: "",
		myNS:             "",
		routeClient:      nil,
		kfmrRoute: &routev1.Route{
			Spec: routev1.RouteSpec{
				Host: "http://foo.com",
			},
			Status: routev1.RouteStatus{Ingress: []routev1.RouteIngress{{}}},
		},
		bkstg:      backstage.SetupBackstageTestRESTClient(bts),
		brdgImport: stub.SetupBridgeLocationRESTClient(brts),
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
		r.pushedLocations = sync.Map{}
		if tc.kfmrSvr != nil {
			cfg := &config.Config{}
			stub.SetupKubeflowTestRESTClient(tc.kfmrSvr, cfg)
			r.kfmr = kubeflowmodelregistry.SetupKubeflowRESTClient(cfg)
		}
		r.client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()

		r.innerStart(ctx)

		found := false
		r.pushedLocations.Range(func(key, value any) bool {
			t.Log(fmt.Sprintf("found key %s for test %s", key, tc.name))
			if key == tc.expectedKey {
				found = true
			}
			if found {
				postBody, ok := value.(*server.PostBody)
				stub.AssertEqual(t, ok, true)
				pb := string(postBody.Body)
				stub.AssertContains(t, pb, []string{tc.expectedValue})
			}

			return true
		})
		stub.AssertEqual(t, found, true)
	}

}

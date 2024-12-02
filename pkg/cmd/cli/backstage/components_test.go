package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"testing"
)

func TestListComponents(t *testing.T) {
	ts := CreateServer(t)
	defer ts.Close()

	// Get with no args calls List
	str, err := SetupBackstageTestRESTClient(ts).GetComponent()
	stub.AssertError(t, err)
	stub.AssertLineCompare(t, str, components, 0)
}

func TestGetComponents(t *testing.T) {
	ts := CreateServer(t)
	defer ts.Close()

	nsName := "default:ollama-service-component"
	str, err := SetupBackstageTestRESTClient(ts).GetComponent(nsName)

	stub.AssertError(t, err)
	stub.AssertContains(t, str, []string{nsName})
}

func TestGetComponentsError(t *testing.T) {
	ts := CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetComponent(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetComponentsWithTags(t *testing.T) {
	ts := CreateServer(t)
	defer ts.Close()

	bs := SetupBackstageTestRESTClient(ts)
	bs.Tags = true

	for _, tc := range []struct {
		args   []string
		str    string
		subset bool
	}{
		{
			args: []string{"genai", "meta"},
		},
		{
			args: []string{"gateway", "authenticated", "developer-model-service", "llm", "vllm", "ibm-granite", "genai"},
			str:  componentsFromTagsNoSubset,
		},
		{
			args:   []string{"genai"},
			subset: true,
			str:    componentsFromTags,
		},
	} {
		bs.Subset = tc.subset
		str, err := bs.GetComponent(tc.args...)
		stub.AssertError(t, err)
		stub.AssertLineCompare(t, str, tc.str, 0)
	}
}

const (
	componentsJson = `{"items":[{"metadata":{"namespace":"default","annotations":{"backstage.io/managed-by-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml","backstage.io/managed-by-origin-location":"url:https://github.com/johnmcollier/model-catalog-reference/blob/main/developer-model-service/catalog-info.yaml","backstage.io/view-url":"https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml","backstage.io/edit-url":"https://github.com/johnmcollier/model-catalog-reference/edit/main/developer-model-service/catalog-info.yaml","backstage.io/source-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/","backstage.io/techdocs-ref":"dir:./"},"name":"developer-model-service","description":"A vLLM and 3scale-based model service providing models for developer tools. A single model (IBM Granite Code 8b) is deployed on it through Red Hat OpenShift AI, and accessed over a secured API.","links":[{"url":"https://model-service.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"Access","type":"website","icon":"WebAssett"},{"url":"https://ibm-granite-8b-code-instruct-vllm.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"API URL","type":"website","icon":"WebAsset"}],"tags":["genai","ibm-granite","vllm","llm","developer-model-service","authenticated","gateway"],"uid":"c8bdd75e-22e3-479a-bba9-e6793056d838","etag":"3cefd4371db72b80c9d6b8d034542833c3094bea"},"apiVersion":"backstage.io/v1alpha1","kind":"Component","spec":{"type":"model-server","owner":"user:exampleuser","lifecycle":"production","providesApis":["model-service-api"],"dependsOn":["resource:ibm-granite-8b-code-instruct","api:model-service-api"],"profile":{"displayName":"Developer Model Service"}},"relations":[{"type":"dependsOn","targetRef":"api:default/model-service-api"},{"type":"dependsOn","targetRef":"resource:default/ibm-granite-8b-code-instruct"},{"type":"ownedBy","targetRef":"user:default/exampleuser"},{"type":"providesApi","targetRef":"api:default/model-service-api"}]},{"metadata":{"namespace":"default","annotations":{"backstage.io/managed-by-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/managed-by-origin-location":"url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml","backstage.io/view-url":"https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/edit-url":"https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml","backstage.io/source-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/","backstage.io/techdocs-ref":"dir:./"},"name":"ollama-model-service","description":"Ollama-based model service running Red Hat OpenShift providing a variety of LLMs. The models are available over a simple OpenShift route, providing an easy way to quickly test out new models.","links":[{"url":"https://model-service.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"Access","type":"website","icon":"WebAssett"},{"url":"https://ollama-model-service-apicast-production.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"API URL","type":"website","icon":"WebAsset"}],"tags":["genai","gemma2","llama3","mistral","phi","granite-code","ollama","llm","ollama-model-service"],"uid":"e50d0a02-4fef-4919-ae7e-28f278a7a8f0","etag":"b64ed630da5cb825b07af3e888cf3ff2cb077cd0"},"apiVersion":"backstage.io/v1alpha1","kind":"Component","spec":{"type":"model-server","owner":"user:exampleuser","lifecycle":"production","providesApis":["ollama-service-api"],"dependsOn":["resource:granite-20b-code-instruct","resource:meta-llama-32-1b","resource:gemma2-2b","resource:phi-mini-instruct","api:ollama-service-api"],"profile":{"displayName":"Ollama Model Service"}},"relations":[{"type":"dependsOn","targetRef":"api:default/ollama-service-api"},{"type":"dependsOn","targetRef":"resource:default/gemma2-2b"},{"type":"dependsOn","targetRef":"resource:default/granite-20b-code-instruct"},{"type":"dependsOn","targetRef":"resource:default/meta-llama-32-1b"},{"type":"dependsOn","targetRef":"resource:default/phi-mini-instruct"},{"type":"ownedBy","targetRef":"user:default/exampleuser"},{"type":"providesApi","targetRef":"api:default/ollama-service-api"}]}],"totalItems":2,"pageInfo":{}}`

	components = `[
    {
        "metadata": {
            "uid": "c8bdd75e-22e3-479a-bba9-e6793056d838",
            "etag": "3cefd4371db72b80c9d6b8d034542833c3094bea",
            "name": "developer-model-service",
            "namespace": "default",
            "description": "A vLLM and 3scale-based model service providing models for developer tools. A single model (IBM Granite Code 8b) is deployed on it through Red Hat OpenShift AI, and accessed over a secured API.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/developer-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/developer-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/",
                "backstage.io/techdocs-ref": "dir:./",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml"
            },
            "tags": [
                "genai",
                "ibm-granite",
                "vllm",
                "llm",
                "developer-model-service",
                "authenticated",
                "gateway"
            ],
            "links": [
                {
                    "url": "https://model-service.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "Access",
                    "icon": "WebAssett",
                    "type": "website"
                },
                {
                    "url": "https://ibm-granite-8b-code-instruct-vllm.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependsOn",
                "targetRef": "api:default/model-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/ibm-granite-8b-code-instruct",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "ownedBy",
                "targetRef": "user:default/exampleuser",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "providesApi",
                "targetRef": "api:default/model-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Component",
        "spec": {
            "type": "model-server",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "model-service-api"
            ],
            "dependsOn": [
                "resource:ibm-granite-8b-code-instruct",
                "api:model-service-api"
            ],
            "profile": {
                "displayName": "Developer Model Service"
            }
        }
    },
    {
        "metadata": {
            "uid": "e50d0a02-4fef-4919-ae7e-28f278a7a8f0",
            "etag": "b64ed630da5cb825b07af3e888cf3ff2cb077cd0",
            "name": "ollama-model-service",
            "namespace": "default",
            "description": "Ollama-based model service running Red Hat OpenShift providing a variety of LLMs. The models are available over a simple OpenShift route, providing an easy way to quickly test out new models.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/",
                "backstage.io/techdocs-ref": "dir:./",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml"
            },
            "tags": [
                "genai",
                "gemma2",
                "llama3",
                "mistral",
                "phi",
                "granite-code",
                "ollama",
                "llm",
                "ollama-model-service"
            ],
            "links": [
                {
                    "url": "https://model-service.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "Access",
                    "icon": "WebAssett",
                    "type": "website"
                },
                {
                    "url": "https://ollama-model-service-apicast-production.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependsOn",
                "targetRef": "api:default/ollama-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/gemma2-2b",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/granite-20b-code-instruct",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/meta-llama-32-1b",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/phi-mini-instruct",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "ownedBy",
                "targetRef": "user:default/exampleuser",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "providesApi",
                "targetRef": "api:default/ollama-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Component",
        "spec": {
            "type": "model-server",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "ollama-service-api"
            ],
            "dependsOn": [
                "resource:granite-20b-code-instruct",
                "resource:meta-llama-32-1b",
                "resource:gemma2-2b",
                "resource:phi-mini-instruct",
                "api:ollama-service-api"
            ],
            "profile": {
                "displayName": "Ollama Model Service"
            }
        }
    }
]
`

	componentsFromTags = `[
    {
        "metadata": {
            "uid": "c8bdd75e-22e3-479a-bba9-e6793056d838",
            "etag": "3cefd4371db72b80c9d6b8d034542833c3094bea",
            "name": "developer-model-service",
            "namespace": "default",
            "description": "A vLLM and 3scale-based model service providing models for developer tools. A single model (IBM Granite Code 8b) is deployed on it through Red Hat OpenShift AI, and accessed over a secured API.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/developer-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/developer-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/",
                "backstage.io/techdocs-ref": "dir:./",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml"
            },
            "tags": [
                "genai",
                "ibm-granite",
                "vllm",
                "llm",
                "developer-model-service",
                "authenticated",
                "gateway"
            ],
            "links": [
                {
                    "url": "https://model-service.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "Access",
                    "icon": "WebAssett",
                    "type": "website"
                },
                {
                    "url": "https://ibm-granite-8b-code-instruct-vllm.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependsOn",
                "targetRef": "api:default/model-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/ibm-granite-8b-code-instruct",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "ownedBy",
                "targetRef": "user:default/exampleuser",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "providesApi",
                "targetRef": "api:default/model-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Component",
        "spec": {
            "type": "model-server",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "model-service-api"
            ],
            "dependsOn": [
                "resource:ibm-granite-8b-code-instruct",
                "api:model-service-api"
            ],
            "profile": {
                "displayName": "Developer Model Service"
            }
        }
    },
    {
        "metadata": {
            "uid": "e50d0a02-4fef-4919-ae7e-28f278a7a8f0",
            "etag": "b64ed630da5cb825b07af3e888cf3ff2cb077cd0",
            "name": "ollama-model-service",
            "namespace": "default",
            "description": "Ollama-based model service running Red Hat OpenShift providing a variety of LLMs. The models are available over a simple OpenShift route, providing an easy way to quickly test out new models.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/",
                "backstage.io/techdocs-ref": "dir:./",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml"
            },
            "tags": [
                "genai",
                "gemma2",
                "llama3",
                "mistral",
                "phi",
                "granite-code",
                "ollama",
                "llm",
                "ollama-model-service"
            ],
            "links": [
                {
                    "url": "https://model-service.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "Access",
                    "icon": "WebAssett",
                    "type": "website"
                },
                {
                    "url": "https://ollama-model-service-apicast-production.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependsOn",
                "targetRef": "api:default/ollama-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/gemma2-2b",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/granite-20b-code-instruct",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/meta-llama-32-1b",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/phi-mini-instruct",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "ownedBy",
                "targetRef": "user:default/exampleuser",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "providesApi",
                "targetRef": "api:default/ollama-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Component",
        "spec": {
            "type": "model-server",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "ollama-service-api"
            ],
            "dependsOn": [
                "resource:granite-20b-code-instruct",
                "resource:meta-llama-32-1b",
                "resource:gemma2-2b",
                "resource:phi-mini-instruct",
                "api:ollama-service-api"
            ],
            "profile": {
                "displayName": "Ollama Model Service"
            }
        }
    }
]`

	componentsFromTagsNoSubset = `[
    {
        "metadata": {
            "uid": "c8bdd75e-22e3-479a-bba9-e6793056d838",
            "etag": "3cefd4371db72b80c9d6b8d034542833c3094bea",
            "name": "developer-model-service",
            "namespace": "default",
            "description": "A vLLM and 3scale-based model service providing models for developer tools. A single model (IBM Granite Code 8b) is deployed on it through Red Hat OpenShift AI, and accessed over a secured API.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/developer-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/developer-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/",
                "backstage.io/techdocs-ref": "dir:./",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml"
            },
            "tags": [
                "authenticated",
                "developer-model-service",
                "gateway",
                "genai",
                "ibm-granite",
                "llm",
                "vllm"
            ],
            "links": [
                {
                    "url": "https://model-service.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "Access",
                    "icon": "WebAssett",
                    "type": "website"
                },
                {
                    "url": "https://ibm-granite-8b-code-instruct-vllm.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependsOn",
                "targetRef": "api:default/model-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "dependsOn",
                "targetRef": "resource:default/ibm-granite-8b-code-instruct",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "ownedBy",
                "targetRef": "user:default/exampleuser",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            },
            {
                "type": "providesApi",
                "targetRef": "api:default/model-service-api",
                "target": {
                    "name": "",
                    "kind": "",
                    "namespace": ""
                }
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Component",
        "spec": {
            "type": "model-server",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "model-service-api"
            ],
            "dependsOn": [
                "resource:ibm-granite-8b-code-instruct",
                "api:model-service-api"
            ],
            "profile": {
                "displayName": "Developer Model Service"
            }
        }
    }
]`
)

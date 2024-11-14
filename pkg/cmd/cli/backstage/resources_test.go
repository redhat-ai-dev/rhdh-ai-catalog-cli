package backstage

import (
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/test/stub"
	"testing"
)

func TestListResources(t *testing.T) {
	ts := CreateServer(t)
	defer ts.Close()

	// Get with no args calls List
	str, err := SetupBackstageTestRESTClient(ts).GetResource()
	stub.AssertError(t, err)
	stub.AssertLineCompare(t, str, resources, 0)
}

func TestGetResources(t *testing.T) {
	ts := CreateServer(t)
	defer ts.Close()

	nsName := "default:phi-mini-instruct"
	str, err := SetupBackstageTestRESTClient(ts).GetResource(nsName)

	stub.AssertError(t, err)
	stub.AssertContains(t, str, []string{nsName})
}

func TestGetResourceError(t *testing.T) {
	ts := CreateServer(t)
	defer ts.Close()

	nsName := "404:404"
	_, err := SetupBackstageTestRESTClient(ts).GetResource(nsName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetResourceWithTags(t *testing.T) {
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
			args: []string{"task-text-generation", "multilingual", "meta", "llm", "llama", "genai", "conversational", "1b"},
			str:  resourcesFromTagsNoSubset,
		},
		{
			args:   []string{"genai", "meta"},
			subset: true,
			str:    resourcesFromTags,
		},
	} {
		bs.Subset = tc.subset
		str, err := bs.GetResource(tc.args...)
		stub.AssertError(t, err)
		stub.AssertLineCompare(t, str, tc.str, 0)
	}
}

const (
	resourcesJson = `{"items":[{"metadata":{"namespace":"default","annotations":{"backstage.io/managed-by-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/managed-by-origin-location":"url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml","backstage.io/view-url":"https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/edit-url":"https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml","backstage.io/source-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/","backstage.io/techdocs-ref":"dir:phi-mini-instruct"},"name":"phi-mini-instruct","description":"Phi-3.5-mini is a lightweight, state-of-the-art open model built upon datasets used for Phi-3 - synthetic data and filtered publicly available websites - with a focus on very high-quality, reasoning dense data. The model belongs to the Phi-3 model family and supports 128K token context length. The model underwent a rigorous enhancement process, incorporating both supervised fine-tuning, proximal policy optimization, and direct preference optimization to ensure precise instruction adherence and robust safety measures.","links":[{"url":"https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"API URL","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/microsoft/Phi-3.5-mini-instruct","title":"Huggingface","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/microsoft/Phi-3.5-mini-instruct/tree/main","title":"Download Model Artifacts","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/microsoft/Phi-3.5-mini-instruct/resolve/main/LICENSE","title":"MIT License","type":"website","icon":"WebAsset"}],"tags":["phi","microsoft","llm","task-strong-reasoning","task-text-generation"],"uid":"18ebee9f-69bb-4ca0-8225-7aa6fa8f1bbf","etag":"3ebfb86df72652320f6dc2db180f8e6e56b59c47"},"apiVersion":"backstage.io/v1alpha1","kind":"Resource","spec":{"type":"ai-model","owner":"user:exampleuser","lifecycle":"production","providesApis":["ollama-service-api"],"dependencyOf":["component:ollama-model-service"],"profile":{"displayName":"Microsoft Phi-3.5 Mini Instruct"}},"relations":[{"type":"dependencyOf","targetRef":"component:default/ollama-model-service"},{"type":"ownedBy","targetRef":"user:default/exampleuser"}]},{"metadata":{"namespace":"default","annotations":{"backstage.io/managed-by-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/managed-by-origin-location":"url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml","backstage.io/view-url":"https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/edit-url":"https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml","backstage.io/source-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/","backstage.io/techdocs-ref":"dir:granite-20b-code-instruct"},"name":"granite-20b-code-instruct","description":"IBM Granite is a decoder-only code model for code generative tasks (e.g. fixing bugs, explaining code, documenting code. Trained with code written in 116 programming languages.","links":[{"url":"https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"API URL","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/ibm-granite/granite-20b-code-instruct-8k","title":"Huggingface","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/ibm-granite/granite-20b-code-instruct-8k/tree/main","title":"Download Model Artifacts","type":"website","icon":"WebAsset"},{"url":"https://www.apache.org/licenses/LICENSE-2.0","title":"Apache-2.0 License","type":"website","icon":"WebAsset"}],"tags":["genai","ibm","llm","granite","conversational","task-text-generation","20b"],"uid":"8aaf3892-11f7-4153-9fe3-18ea228fe4f2","etag":"de5f7cacb6eaa1690a053a5e0ab9ba4d5d05227a"},"apiVersion":"backstage.io/v1alpha1","kind":"Resource","spec":{"type":"ai-model","owner":"user:exampleuser","lifecycle":"production","providesApis":["ollama-service-api"],"dependencyOf":["component:ollama-model-service"],"profile":{"displayName":"IBM Granite Code Model 20B"}},"relations":[{"type":"dependencyOf","targetRef":"component:default/ollama-model-service"},{"type":"ownedBy","targetRef":"user:default/exampleuser"}]},{"metadata":{"namespace":"default","annotations":{"backstage.io/managed-by-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml","backstage.io/managed-by-origin-location":"url:https://github.com/johnmcollier/model-catalog-reference/blob/main/developer-model-service/catalog-info.yaml","backstage.io/view-url":"https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml","backstage.io/edit-url":"https://github.com/johnmcollier/model-catalog-reference/edit/main/developer-model-service/catalog-info.yaml","backstage.io/source-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/","backstage.io/techdocs-ref":"dir:ibm-granite-8b-code-instruct"},"name":"ibm-granite-8b-code-instruct","description":"IBM Granite is a decoder-only code model for code generative tasks (e.g. fixing bugs, explaining code, documenting code. Trained with code written in 116 programming languages.","links":[{"url":"https://model-service.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"Access","type":"website","icon":"WebAssett"},{"url":"https://ibm-granite-8b-code-instruct-vllm.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"API URL","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/ibm-granite/granite-8b-code-instruct","title":"Model Repository","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/ibm-granite/granite-8b-code-instruct-4k/tree/main","title":"Download Model Artifacts","type":"website","icon":"WebAsset"},{"url":"https://www.apache.org/licenses/LICENSE-2.0","title":"Apache-2.0 License","type":"website","icon":"WebAsset"}],"tags":["genai","ibm","llm","granite","conversational","task-text-generation"],"uid":"c2765b38-97e4-403e-a22f-f12088539856","etag":"7473a2c6058aebe0230260aa91784d6de8b9f914"},"apiVersion":"backstage.io/v1alpha1","kind":"Resource","spec":{"type":"ai-model","owner":"user:exampleuser","lifecycle":"production","providesApis":["model-service-api"],"dependencyOf":["component:developer-model-service"],"profile":{"displayName":"IBM Granite Code Model"}},"relations":[{"type":"dependencyOf","targetRef":"component:default/developer-model-service"},{"type":"ownedBy","targetRef":"user:default/exampleuser"}]},{"metadata":{"namespace":"default","annotations":{"backstage.io/managed-by-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/managed-by-origin-location":"url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml","backstage.io/view-url":"https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/edit-url":"https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml","backstage.io/source-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/","backstage.io/techdocs-ref":"dir:gemma2-2b"},"name":"gemma2-2b","description":"Gemma is a family of lightweight, state-of-the-art open models from Google, built from the same research and technology used to create the Gemini models. They are text-to-text, decoder-only large language models, available in English, with open weights for both pre-trained variants and instruction-tuned variants. Gemma models are well-suited for a variety of text generation tasks, including question answering, summarization, and reasoning. Their relatively small size makes it possible to deploy them in environments with limited resources such as a laptop, desktop or your own cloud infrastructure, democratizing access to state of the art AI models and helping foster innovation for everyone.","links":[{"url":"https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"API URL","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/google/gemma-2-2b","title":"Huggingface","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/google/gemma-2-2b/tree/main","title":"Download Model Artifacts","type":"website","icon":"WebAsset"},{"url":"https://ai.google.dev/gemma/terms","title":"Gemma License","type":"website","icon":"WebAsset"}],"tags":["genai","google","llm","gemma2","text-to-text","task-text-generation","2b"],"uid":"ca9804f9-0bd0-4fb1-af00-29984919452a","etag":"6a1021a0b1aa90996fb350b43dc0e787785ea135"},"apiVersion":"backstage.io/v1alpha1","kind":"Resource","spec":{"type":"ai-model","owner":"user:exampleuser","lifecycle":"production","providesApis":["ollama-service-api"],"dependencyOf":["component:ollama-model-service"],"profile":{"displayName":"Google Gemma2 2B"}},"relations":[{"type":"dependencyOf","targetRef":"component:default/ollama-model-service"},{"type":"ownedBy","targetRef":"user:default/exampleuser"}]},{"metadata":{"namespace":"default","annotations":{"backstage.io/managed-by-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/managed-by-origin-location":"url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml","backstage.io/view-url":"https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml","backstage.io/edit-url":"https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml","backstage.io/source-location":"url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/","backstage.io/techdocs-ref":"dir:meta-llama-32-1b"},"name":"meta-llama-32-1b","description":"The Meta Llama 3.2 collection of multilingual large language models (LLMs) is a collection of pretrained and instruction-tuned generative models.","links":[{"url":"https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com","title":"API URL","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/meta-llama/Llama-3.2-1B","title":"Huggingface","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/meta-llama/Llama-3.2-1B/tree/main","title":"Download Model Artifacts","type":"website","icon":"WebAsset"},{"url":"https://huggingface.co/meta-llama/Llama-3.2-1B/blob/main/LICENSE.txt","title":"Llama-3.2 License","type":"website","icon":"WebAsset"}],"tags":["genai","meta","llm","llama","conversational","multilingual","task-text-generation","1b"],"uid":"cb4d6cb2-8e80-430b-9dbd-7ca6df639b58","etag":"65d92a73982463dc8c04453f6949b0693f2f7817"},"apiVersion":"backstage.io/v1alpha1","kind":"Resource","spec":{"type":"ai-model","owner":"user:exampleuser","lifecycle":"production","providesApis":["ollama-service-api"],"dependencyOf":["component:ollama-model-service"],"profile":{"displayName":"Meta Llama 3.2 1B"}},"relations":[{"type":"dependencyOf","targetRef":"component:default/ollama-model-service"},{"type":"ownedBy","targetRef":"user:default/exampleuser"}]}],"totalItems":5,"pageInfo":{}}`
	resources     = `[
    {
        "metadata": {
            "uid": "18ebee9f-69bb-4ca0-8225-7aa6fa8f1bbf",
            "etag": "3ebfb86df72652320f6dc2db180f8e6e56b59c47",
            "name": "phi-mini-instruct",
            "namespace": "default",
            "description": "Phi-3.5-mini is a lightweight, state-of-the-art open model built upon datasets used for Phi-3 - synthetic data and filtered publicly available websites - with a focus on very high-quality, reasoning dense data. The model belongs to the Phi-3 model family and supports 128K token context length. The model underwent a rigorous enhancement process, incorporating both supervised fine-tuning, proximal policy optimization, and direct preference optimization to ensure precise instruction adherence and robust safety measures.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/",
                "backstage.io/techdocs-ref": "dir:phi-mini-instruct",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml"
            },
            "tags": [
                "phi",
                "microsoft",
                "llm",
                "task-strong-reasoning",
                "task-text-generation"
            ],
            "links": [
                {
                    "url": "https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/microsoft/Phi-3.5-mini-instruct",
                    "title": "Huggingface",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/microsoft/Phi-3.5-mini-instruct/tree/main",
                    "title": "Download Model Artifacts",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/microsoft/Phi-3.5-mini-instruct/resolve/main/LICENSE",
                    "title": "MIT License",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependencyOf",
                "targetRef": "component:default/ollama-model-service",
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
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Resource",
        "spec": {
            "type": "ai-model",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "ollama-service-api"
            ],
            "dependencyOf": [
                "component:ollama-model-service"
            ],
            "profile": {
                "displayName": "Microsoft Phi-3.5 Mini Instruct"
            }
        }
    },
    {
        "metadata": {
            "uid": "8aaf3892-11f7-4153-9fe3-18ea228fe4f2",
            "etag": "de5f7cacb6eaa1690a053a5e0ab9ba4d5d05227a",
            "name": "granite-20b-code-instruct",
            "namespace": "default",
            "description": "IBM Granite is a decoder-only code model for code generative tasks (e.g. fixing bugs, explaining code, documenting code. Trained with code written in 116 programming languages.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/",
                "backstage.io/techdocs-ref": "dir:granite-20b-code-instruct",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml"
            },
            "tags": [
                "genai",
                "ibm",
                "llm",
                "granite",
                "conversational",
                "task-text-generation",
                "20b"
            ],
            "links": [
                {
                    "url": "https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/ibm-granite/granite-20b-code-instruct-8k",
                    "title": "Huggingface",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/ibm-granite/granite-20b-code-instruct-8k/tree/main",
                    "title": "Download Model Artifacts",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://www.apache.org/licenses/LICENSE-2.0",
                    "title": "Apache-2.0 License",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependencyOf",
                "targetRef": "component:default/ollama-model-service",
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
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Resource",
        "spec": {
            "type": "ai-model",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "ollama-service-api"
            ],
            "dependencyOf": [
                "component:ollama-model-service"
            ],
            "profile": {
                "displayName": "IBM Granite Code Model 20B"
            }
        }
    },
    {
        "metadata": {
            "uid": "c2765b38-97e4-403e-a22f-f12088539856",
            "etag": "7473a2c6058aebe0230260aa91784d6de8b9f914",
            "name": "ibm-granite-8b-code-instruct",
            "namespace": "default",
            "description": "IBM Granite is a decoder-only code model for code generative tasks (e.g. fixing bugs, explaining code, documenting code. Trained with code written in 116 programming languages.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/developer-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/developer-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/",
                "backstage.io/techdocs-ref": "dir:ibm-granite-8b-code-instruct",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/developer-model-service/catalog-info.yaml"
            },
            "tags": [
                "genai",
                "ibm",
                "llm",
                "granite",
                "conversational",
                "task-text-generation"
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
                },
                {
                    "url": "https://huggingface.co/ibm-granite/granite-8b-code-instruct",
                    "title": "Model Repository",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/ibm-granite/granite-8b-code-instruct-4k/tree/main",
                    "title": "Download Model Artifacts",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://www.apache.org/licenses/LICENSE-2.0",
                    "title": "Apache-2.0 License",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependencyOf",
                "targetRef": "component:default/developer-model-service",
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
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Resource",
        "spec": {
            "type": "ai-model",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "model-service-api"
            ],
            "dependencyOf": [
                "component:developer-model-service"
            ],
            "profile": {
                "displayName": "IBM Granite Code Model"
            }
        }
    },
    {
        "metadata": {
            "uid": "ca9804f9-0bd0-4fb1-af00-29984919452a",
            "etag": "6a1021a0b1aa90996fb350b43dc0e787785ea135",
            "name": "gemma2-2b",
            "namespace": "default",
            "description": "Gemma is a family of lightweight, state-of-the-art open models from Google, built from the same research and technology used to create the Gemini models. They are text-to-text, decoder-only large language models, available in English, with open weights for both pre-trained variants and instruction-tuned variants. Gemma models are well-suited for a variety of text generation tasks, including question answering, summarization, and reasoning. Their relatively small size makes it possible to deploy them in environments with limited resources such as a laptop, desktop or your own cloud infrastructure, democratizing access to state of the art AI models and helping foster innovation for everyone.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/",
                "backstage.io/techdocs-ref": "dir:gemma2-2b",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml"
            },
            "tags": [
                "genai",
                "google",
                "llm",
                "gemma2",
                "text-to-text",
                "task-text-generation",
                "2b"
            ],
            "links": [
                {
                    "url": "https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/google/gemma-2-2b",
                    "title": "Huggingface",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/google/gemma-2-2b/tree/main",
                    "title": "Download Model Artifacts",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://ai.google.dev/gemma/terms",
                    "title": "Gemma License",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependencyOf",
                "targetRef": "component:default/ollama-model-service",
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
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Resource",
        "spec": {
            "type": "ai-model",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "ollama-service-api"
            ],
            "dependencyOf": [
                "component:ollama-model-service"
            ],
            "profile": {
                "displayName": "Google Gemma2 2B"
            }
        }
    },
    {
        "metadata": {
            "uid": "cb4d6cb2-8e80-430b-9dbd-7ca6df639b58",
            "etag": "65d92a73982463dc8c04453f6949b0693f2f7817",
            "name": "meta-llama-32-1b",
            "namespace": "default",
            "description": "The Meta Llama 3.2 collection of multilingual large language models (LLMs) is a collection of pretrained and instruction-tuned generative models.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/",
                "backstage.io/techdocs-ref": "dir:meta-llama-32-1b",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml"
            },
            "tags": [
                "genai",
                "meta",
                "llm",
                "llama",
                "conversational",
                "multilingual",
                "task-text-generation",
                "1b"
            ],
            "links": [
                {
                    "url": "https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B",
                    "title": "Huggingface",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B/tree/main",
                    "title": "Download Model Artifacts",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B/blob/main/LICENSE.txt",
                    "title": "Llama-3.2 License",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependencyOf",
                "targetRef": "component:default/ollama-model-service",
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
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Resource",
        "spec": {
            "type": "ai-model",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "ollama-service-api"
            ],
            "dependencyOf": [
                "component:ollama-model-service"
            ],
            "profile": {
                "displayName": "Meta Llama 3.2 1B"
            }
        }
    }
]`

	resourcesFromTags = `[
    {
        "metadata": {
            "uid": "cb4d6cb2-8e80-430b-9dbd-7ca6df639b58",
            "etag": "65d92a73982463dc8c04453f6949b0693f2f7817",
            "name": "meta-llama-32-1b",
            "namespace": "default",
            "description": "The Meta Llama 3.2 collection of multilingual large language models (LLMs) is a collection of pretrained and instruction-tuned generative models.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/",
                "backstage.io/techdocs-ref": "dir:meta-llama-32-1b",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml"
            },
            "tags": [
                "1b",
                "conversational",
                "genai",
                "llama",
                "llm",
                "meta",
                "multilingual",
                "task-text-generation"
            ],
            "links": [
                {
                    "url": "https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B",
                    "title": "Huggingface",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B/tree/main",
                    "title": "Download Model Artifacts",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B/blob/main/LICENSE.txt",
                    "title": "Llama-3.2 License",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependencyOf",
                "targetRef": "component:default/ollama-model-service",
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
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Resource",
        "spec": {
            "type": "ai-model",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "ollama-service-api"
            ],
            "dependencyOf": [
                "component:ollama-model-service"
            ],
            "profile": {
                "displayName": "Meta Llama 3.2 1B"
            }
        }
    }
]`

	resourcesFromTagsNoSubset = `[
    {
        "metadata": {
            "uid": "cb4d6cb2-8e80-430b-9dbd-7ca6df639b58",
            "etag": "65d92a73982463dc8c04453f6949b0693f2f7817",
            "name": "meta-llama-32-1b",
            "namespace": "default",
            "description": "The Meta Llama 3.2 collection of multilingual large language models (LLMs) is a collection of pretrained and instruction-tuned generative models.",
            "annotations": {
                "backstage.io/edit-url": "https://github.com/johnmcollier/model-catalog-reference/edit/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/managed-by-origin-location": "url:https://github.com/johnmcollier/model-catalog-reference/blob/main/ollama-model-service/catalog-info.yaml",
                "backstage.io/source-location": "url:https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/",
                "backstage.io/techdocs-ref": "dir:meta-llama-32-1b",
                "backstage.io/view-url": "https://github.com/johnmcollier/model-catalog-reference/tree/main/ollama-model-service/catalog-info.yaml"
            },
            "tags": [
                "1b",
                "conversational",
                "genai",
                "llama",
                "llm",
                "meta",
                "multilingual",
                "task-text-generation"
            ],
            "links": [
                {
                    "url": "https://ollama-route-ollama.apps.rosa.redhat-ai-dev.m6no.p3.openshiftapps.com",
                    "title": "API URL",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B",
                    "title": "Huggingface",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B/tree/main",
                    "title": "Download Model Artifacts",
                    "icon": "WebAsset",
                    "type": "website"
                },
                {
                    "url": "https://huggingface.co/meta-llama/Llama-3.2-1B/blob/main/LICENSE.txt",
                    "title": "Llama-3.2 License",
                    "icon": "WebAsset",
                    "type": "website"
                }
            ]
        },
        "relations": [
            {
                "type": "dependencyOf",
                "targetRef": "component:default/ollama-model-service",
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
            }
        ],
        "apiVersion": "backstage.io/v1alpha1",
        "kind": "Resource",
        "spec": {
            "type": "ai-model",
            "lifecycle": "production",
            "owner": "user:exampleuser",
            "providesApis": [
                "ollama-service-api"
            ],
            "dependencyOf": [
                "component:ollama-model-service"
            ],
            "profile": {
                "displayName": "Meta Llama 3.2 1B"
            }
        }
    }
]`
)

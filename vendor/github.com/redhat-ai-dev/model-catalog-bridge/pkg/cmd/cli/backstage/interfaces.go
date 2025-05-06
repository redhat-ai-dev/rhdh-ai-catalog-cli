package backstage

// model catalog json schema populators

import (
	"io"
	"strings"

	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/util"
	"github.com/redhat-ai-dev/model-catalog-bridge/schema/types/golang"
	"k8s.io/klog/v2"
)

type ModelCatalogPopulator interface {
	GetModels() []golang.Model
	GetModelServer() *golang.ModelServer
}

type CommonModelSchemaPopulator interface {
	GetName() string
	GetOwner() string
	GetLifecyle() string
	GetDescription() string
	GetTags() []string
}

type ModelServerPopulator interface {
	CommonModelSchemaPopulator

	GetAPI() *golang.API
	GetAuthentication() *bool
	GetHomepageURL() *string
	GetUsage() *string
}

type ModelPopulator interface {
	CommonModelSchemaPopulator

	GetArtifactLocationURL() *string
	GetEthics() *string
	GetHowToUseURL() *string
	GetSupport() *string
	GetTraining() *string
	GetUsage() *string
	GetTechDocs() *string
}

type ModelServerAPIPopulator interface {
	GetSpec() string
	GetTags() []string
	GetType() golang.Type
	GetURL() string
}

func PrintModelCatalogPopulator(svrPop ModelCatalogPopulator, writer io.Writer) error {
	modelCatalog := &golang.ModelCatalog{
		Models: svrPop.GetModels(),
	}
	ms := svrPop.GetModelServer()
	// only add the model server if it has an inference endpoint URL
	if ms != nil && ms.API != nil && len(ms.API.URL) > 0 {
		modelCatalog.ModelServer = ms
	}
	err := util.PrintJSON(modelCatalog, writer)
	if err != nil {
		klog.Errorf("ERROR: converting ModelCatalog to yaml and printing: %s, %#v", err.Error(), modelCatalog)
		return err
	}
	return nil
}

// catalog-info.yaml populators

type CommonPopulator interface {
	GetOwner() string
	GetLifecycle() string
	GetName() string
	GetDescription() string
	GetLinks() []EntityLink
	GetTags() []string
	GetProvidedAPIs() []string
	GetTechdocRef() string
	GetDisplayName() string
}

type ComponentPopulator interface {
	CommonPopulator
	GetDependsOn() []string
}

type ResourcePopulator interface {
	CommonPopulator
	GetDependencyOf() []string
}

type APIPopulator interface {
	CommonPopulator
	GetDefinition() string
	GetDependencyOf() []string
}

func PrintComponent(pop ComponentPopulator, writer io.Writer) error {
	component := &ComponentEntityV1alpha1{
		Kind:       "Component",
		ApiVersion: VERSION,
		Entity:     buildEntity("Component", pop),
	}
	component.Entity.Metadata.Annotations = map[string]string{TECHDOC_REFS: pop.GetTechdocRef()}
	component.Metadata = component.Entity.Metadata
	component.Spec = &ComponentEntityV1alpha1Spec{
		Type:         COMPONENT_TYPE,
		Lifecycle:    pop.GetLifecycle(),
		Owner:        "user:" + pop.GetOwner(),
		ProvidesApis: pop.GetProvidedAPIs(),
		DependsOn:    pop.GetDependsOn(),
		Profile:      Profile{DisplayName: pop.GetDisplayName()},
	}
	err := util.PrintYaml(component, true, writer)
	if err != nil {
		klog.Errorf("ERROR: converting component to yaml and printing: %s, %#v", err.Error(), component)
		return err
	}
	return nil
}

func PrintResource(pop ResourcePopulator, writer io.Writer) error {
	resource := &ResourceEntityV1alpha1{
		Kind:       "Resource",
		ApiVersion: VERSION,
		Entity:     buildEntity("Resource", pop),
	}
	resource.Entity.Metadata.Annotations = map[string]string{TECHDOC_REFS: pop.GetTechdocRef()}
	resource.Metadata = resource.Entity.Metadata
	resource.Spec = &ResourceEntityV1alpha1Spec{
		Type:         RESOURCE_TYPE,
		Owner:        "user:" + pop.GetOwner(),
		Lifecycle:    pop.GetLifecycle(),
		ProvidesApis: pop.GetProvidedAPIs(),
		DependencyOf: pop.GetDependencyOf(),
		Profile:      Profile{DisplayName: pop.GetDisplayName()},
	}
	err := util.PrintYaml(resource, true, writer)
	if err != nil {
		klog.Errorf("ERROR: converting resource to yaml and printing: %s, %#v", err.Error(), resource)
		return err
	}
	return nil
}

func PrintAPI(pop APIPopulator, writer io.Writer) error {
	api := &ApiEntityV1alpha1{
		Kind:       "API",
		ApiVersion: VERSION,
		Entity:     buildEntity("API", pop),
	}
	api.Entity.Metadata.Annotations = map[string]string{TECHDOC_REFS: pop.GetTechdocRef()}
	api.Metadata = api.Entity.Metadata
	api.Spec = &ApiEntityV1alpha1Spec{
		Type:         "",
		Lifecycle:    pop.GetLifecycle(),
		Owner:        "user:" + pop.GetOwner(),
		Definition:   pop.GetDefinition(),
		DependencyOf: pop.GetDependencyOf(),
		Profile:      Profile{DisplayName: pop.GetDisplayName()},
	}
	switch {
	case strings.Contains(api.Spec.Definition, OPENAPI_API_TYPE):
		api.Spec.Type = OPENAPI_API_TYPE
	case strings.Contains(api.Spec.Definition, ASYNCAPI_API_TYPE):
		api.Spec.Type = ASYNCAPI_API_TYPE
	case strings.Contains(api.Spec.Definition, GRAPHQL_API_TYPE):
		api.Spec.Type = GRAPHQL_API_TYPE
	case strings.Contains(api.Spec.Definition, TRPC_API_TYPE):
		api.Spec.Type = TRPC_API_TYPE
	case strings.Contains(api.Spec.Definition, "proto"):
		api.Spec.Type = GRPC_API_TYPE
	default:
		api.Spec.Type = UNKNOWN_API_TYPE
	}

	err := util.PrintYaml(api, false, writer)
	if err != nil {
		klog.Errorf("ERROR: converting api to yaml and printing: %s, %#v", err.Error(), api)
		return err
	}
	return nil
}

func buildEntity(kind string, pop CommonPopulator) Entity {
	entity := Entity{
		Kind:       kind,
		ApiVersion: VERSION,
		Metadata: EntityMeta{
			Name:        pop.GetName(),
			Description: pop.GetDescription(),
			Tags:        pop.GetTags(),
			Links:       pop.GetLinks(),
		},
	}
	return entity
}

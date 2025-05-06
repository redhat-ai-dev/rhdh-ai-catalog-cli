package types

type NormalizerFormat string

const (
	CatalogInfoYamlFormat NormalizerFormat = "CatalogInfoYamlFormat"
	JsonArrayForamt       NormalizerFormat = "JsonArrayFormat"

	LocationUrlEnvVar        = "BRIDGE_URL"
	ModelRegistryRouteEnvVar = "MR_ROUTE"
	BackstageUrlEnvVar       = "BKSTG_URL"
	FormatEnvVar             = "NORMALIZER_FORMAT"
	PollingIntEnvVar         = "POLLING_INTERVAL"

	RHDHTokenEnvVar          = "RHDH_TOKEN"
	ModelRegistryTokenEnvVar = "KFMR_TOKEN"

	OwnerEnvVar     = "DEFAULT_OWNER"
	LifecycleEnvVar = "DEFAULT_LIFECYCLE"
)

// These custom property keys are what RHOAI Model Catalog define for metadata in their UID, which gets propagated to RHOAI Model Registry
// ModelVersion
const (
	RHOAIModelCatalogLicenseKey         = "License"
	RHOAIModelCatalogProviderKey        = "Provider"
	RHOAIModelCatalogRegisteredFromKey  = "Registered from"
	RHOAIModelCatalogSourceModelKey     = "Source model"
	RHOAIModelCatalogSourceModelVersion = "Source model version"
)

// These key values are internal RHOAI Model Registry keys which also get set during RHOAI model catalog to model registry propagation
// where there are some duplicates with the Model Catalog keys abvoe
const (
	// apparent duplicate of RHOAIModelCatalogSourceModelKey
	RHOAIModelRegistryRegisteredFromCatalogModelName      = "_registeredFromCatalogModelName"
	RHOAIModelRegistryRegisteredFromCatalogRepositoryName = "_registeredFromCatalogRepositoryName"
	// apparent duplicate of RHOAIModelCatalogProviderKey
	RHOAIModelRegistryRegisteredFromCatalogSourceName = "_registeredFromCatalogSourceName"
	// apparent duplicate of RHOAIModelCatalogSourceModelVersion
	RHOAIModelRegistryRegisteredFromCatalogTag = "_registeredFromCatalogTag"
	// they post the last modified type as a k/v ... value not useful without key, so combine perhaps
	RHOAIModelRegistryLastModified = "_lastModified"
)

// name of kubeflow inference_service after the '/<random ID>' may match name of model from the '/v1/models' query that
// we want for the resource, when filling out templates ... need to turn dots into empty chars

// These are the keys that we will expose to users of RHOAI Model Registry for setting data the normalizer will seed
// into the JSON array
const (
	EthicsKey      = "Ethics"
	HowToUseKey    = "How to use"
	SupportKey     = "Support"
	TrainingKey    = "Training"
	UsageKey       = "Usage"
	HomepageURLKey = "Homepage URL"
	APISpecKey     = "API Spec"
	APITypeKey     = "API Type"
	Owner          = "Owner"
	Lifecycle      = "Lifecycle"
	TechDocsKey    = "TechDocs"
	LicenseKey     = "License"
)

// These const represent the curated techdocs repos we provide for certain models in the RHOAI model catalog
const (
	Granite318bLabName     = "granite-31-8b-lab"
	Granite318bLabTechDocs = "https://github.com/redhat-ai-dev/granite-3.1-8b-lab-docs/tree/main"
)

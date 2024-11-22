package backstage

// KindAPI defines name for API kind.
const KindAPI = "API"

// So there are upstream projects which took the backstage schema definitions, such as
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/kinds/API.v1alpha1.schema.json
// and attempted to auto generate Golang based structs for the API entity.
// However, for reasons we want to try and find, the upstream backstage schema files have not been kept up to date
// with the latest format we see when using the UI. The additions we need for the AI Model Catalog format we've devised
// have been made.  And until we can sort out the upstream schema update policy, we'll have to track changes we want
// to pull and make those manually for now.

type ApiEntityV1alpha1 struct {
	Entity

	// ApiVersion is always "backstage.io/v1alpha1".
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`

	// Kind is always "API".
	Kind string `json:"kind" yaml:"kind"`

	// Spec is the specification data describing the API itself.
	Spec *ApiEntityV1alpha1Spec `json:"spec" yaml:"spec"`
}

// ApiEntityV1alpha1Spec describes the specification data describing the API itself.
type ApiEntityV1alpha1Spec struct {
	// Type of the API definition.
	Type string `json:"type" yaml:"type"`

	// Lifecycle state of the API.
	Lifecycle string `json:"lifecycle" yaml:"lifecycle"`

	// Owner is entity reference to the owner of the API.
	Owner string `json:"owner" yaml:"owner"`

	// Definition of the API, based on the format defined by the type.
	Definition string `json:"definition" yaml:"definition"`

	//FIX from schema
	DependencyOf []string `json:"dependencyOf,omitempty" yaml:"dependencyOf,omitempty"`

	// System is entity reference to the system that the API belongs to.
	System string `json:"system,omitempty" yaml:"system,omitempty"`

	//FIX from schema
	Profile Profile `json:"profile" yaml:"profile"`
}

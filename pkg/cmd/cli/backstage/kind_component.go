package backstage

// KindComponent defines name for component kind.
const KindComponent = "Component"

// So there are upstream projects which took the backstage schema definitions, such as
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/kinds/Component.v1alpha1.schema.json
// and attempted to auto generate Golang based structs for the API entity.
// However, for reasons we want to try and find, the upstream backstage schema files have not been kept up to date
// with the latest format we see when using the UI. The additions we need for the AI Model Catalog format we've devised
// have been made.  And until we can sort out the upstream schema update policy, we'll have to track changes we want
// to pull and make those manually for now.

type ComponentEntityV1alpha1 struct {
	Entity

	// ApiVersion is always "backstage.io/v1alpha1".
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`

	// Kind is always "Component".
	Kind string `json:"kind" yaml:"kind"`

	// Spec is the specification data describing the component itself.
	Spec *ComponentEntityV1alpha1Spec `json:"spec"  yaml:"spec"`
}

// ComponentEntityV1alpha1Spec describes the specification data describing the component itself.
type ComponentEntityV1alpha1Spec struct {
	// Type of component.
	Type string `json:"type" yaml:"type"`

	// Lifecycle state of the component.
	Lifecycle string `json:"lifecycle" yaml:"lifecycle"`

	// Owner is an entity reference to the owner of the component.
	Owner string `json:"owner" yaml:"owner"`

	// SubcomponentOf is an entity reference to another component of which the component is a part.
	SubcomponentOf string `json:"subcomponentOf,omitempty" yaml:"subcomponentOf,omitempty"`

	// ProvidesApis is an array of entity references to the APIs that are provided by the component.
	ProvidesApis []string `json:"providesApis,omitempty" yaml:"providesApis,omitempty"`

	// ConsumesApis is an array of entity references to the APIs that are consumed by the component.
	ConsumesApis []string `json:"consumesApis,omitempty" yaml:"onsumesApis,omitempty"`

	// DependsOn is an array of entity references to the components and resources that the component depends on.
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`

	// System is an array of references to other entities that the component depends on to function.
	System string `json:"system,omitempty" yaml:"system,omitempty"`

	//FIX from schema
	Profile Profile `json:"profile" yaml:"profile"`
}

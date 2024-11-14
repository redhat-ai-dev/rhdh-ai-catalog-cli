package backstage

// KindResource defines name for resource kind.
const KindResource = "Resource"

// ResourceEntityV1alpha1 describes the infrastructure a system needs to operate, like BigTable databases, Pub/Sub topics, S3 buckets
// or CDNs. Modelling them together with components and systems allows to visualize resource footprint, and create tooling around them.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/kinds/Resource.v1alpha1.schema.json
type ResourceEntityV1alpha1 struct {
	Entity

	// ApiVersion is always "backstage.io/v1alpha1".
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`

	// Kind is always "Resource".
	Kind string `json:"kind" yaml:"kind"`

	// Spec is the specification data describing the resource itself.
	Spec *ResourceEntityV1alpha1Spec `json:"spec" yaml:"spec"`
}

// ResourceEntityV1alpha1Spec describes the specification data describing the resource itself.
type ResourceEntityV1alpha1Spec struct {
	// Type of resource.
	Type string `json:"type" yaml:"type"`

	//GGM FIX Lifecycle state of the component.
	Lifecycle string `json:"lifecycle" yaml:"lifecycle"`

	// Owner is an entity reference to the owner of the resource.
	Owner string `json:"owner" yaml:"owner"`

	//FIX from schema ProvidesApis is an array of entity references to the APIs that are provided by the component.
	ProvidesApis []string `json:"providesApis,omitempty" yaml:"providesApis,omitempty"`

	//FIX from schema
	DependencyOf []string `json:"dependencyOf,omitempty" yaml:"dependencyOf,omitempty"`

	// System is an entity reference to the system that the resource belongs to.
	System string `json:"system,omitempty" yaml:"system,omitempty"`

	//FIX from schema
	Profile Profile `json:"profile" yaml:"profile"`
}

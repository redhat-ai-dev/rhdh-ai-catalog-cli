package backstage

// Entity represents the parts of the format that's common to all versions/kinds of entity.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/Entity.schema.json
type Entity struct {
	// ApiVersion is the version of specification format for this particular entity that this is written against.
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`

	// Kind is the high level entity type being described.
	Kind string `json:"kind" yaml:"kind"`

	// Metadata is metadata related to the entity. Should always be "System".
	Metadata EntityMeta `json:"metadata" yaml:"metadata"`

	// Spec is the specification data describing the entity itself.
	Spec map[string]interface{} `json:"spec,omitempty" yaml:"spec,omitempty"`

	// Relations that this entity has with other entities.
	Relations []EntityRelation `json:"relations,omitempty" yaml:"relations,omitempty"`

	// The current status of the entity, as claimed by various sources.
	Status *EntityStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// EntityMeta represents metadata fields common to all versions/kinds of entity.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/EntityMeta.schema.json
type EntityMeta struct {
	// UID A globally unique ID for the entity. This field can not be set by the user at creation time, and the server will reject
	// an attempt to do so. The field will be populated in read operations.
	UID string `json:"uid,omitempty" yaml:"uid,omitempty"`

	// Etag is an opaque string that changes for each update operation to any part of the entity, including metadata. This field
	// can not be set by the user at creation time, and the server will reject an attempt to do so. The field will be populated in read
	// operations.The field can (optionally) be specified when performing update or delete operations, and the server will then reject
	// the operation if it does not match the current stored value.
	Etag string `json:"etag,omitempty" yaml:"etag,omitempty"`

	// Name of the entity. Must be unique within the catalog at any given point in time, for any given namespace + kind pair.
	Name string `json:"name" yaml:"name"`

	// Namespace that the entity belongs to.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Title is a display name of the entity, to be presented in user interfaces instead of the name property, when available.
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Description is a short (typically relatively few words, on one line) description of the entity.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Labels are key/value pairs of identifying information attached to the entity.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// Annotations are key/value pairs of non-identifying auxiliary information attached to the entity.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`

	// Tags is a list of single-valued strings, to for example classify catalog entities in various ways.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`

	// Links is a list of external hyperlinks related to the entity. Links can provide additional contextual
	// information that may be located outside of Backstage itself. For example, an admin dashboard or external CMS page.
	Links []EntityLink `json:"links,omitempty" yaml:"links,omitempty"`
}

// EntityLink represents a link to external information that is related to the entity.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/EntityMeta.schema.json
type EntityLink struct {
	// URL in a standard uri format.
	URL string `json:"url" yaml:"url"`

	// Title is a user-friendly display name for the link.
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Icon is a key representing a visual icon to be displayed in the UI.
	Icon string `json:"icon,omitempty" yaml:"icon,omitempty"`

	// Type is an optional value to categorize links into specific groups.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// EntityRelation is a directed relation from one entity to another.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/shared/common.schema.json
type EntityRelation struct {
	// Type of the relation.
	Type string `json:"type" yaml:"type"`

	// TargetRef is the entity ref of the target of this relation.
	TargetRef string `json:"targetRef" yaml:"targetRef"`

	// Target is the entity of the target of this relation.
	Target EntityRelationTarget `json:"target" yaml:"target"`
}

// EntityRelationTarget describes the target of an entity relation.
type EntityRelationTarget struct {
	// Name of the target entity. Must be unique within the catalog at any given point in time, for any given namespace + kind pair.
	Name string `json:"name" yaml:"name"`

	// Kind is the high level target entity type being described.
	Kind string `json:"kind" yaml:"kind"`

	// Namespace that the target entity belongs to.
	Namespace string `json:"namespace" yaml:"namespace"`
}

// EntityStatus informs current status of the entity, as claimed by various sources.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/shared/common.schema.json
type EntityStatus struct {
	// A specific status item on a well known format.
	Items []EntityStatusItem `json:"items,omitempty" yaml:"items,omitempty"`
}

// EntityStatusItem contains a specific status item on a well known format.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/shared/common.schema.json
type EntityStatusItem struct {
	// The item type
	Type string `json:"type" yaml:"type"`

	// The status level / severity of the status item.
	// Either ["info", "warning", "error"]
	Level string `json:"level" yaml:"level"`

	// A brief message describing the status, intended for human consumption.
	Message string `json:"message" yaml:"message"`

	// An optional serialized error object related to the status.
	Error *EntityStatusItemError `json:"error" yaml:"error"`
}

// EntityStatusItemError has aA serialized error object.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/shared/common.schema.json
type EntityStatusItemError struct {
	// The type name of the error"
	Name string `json:"name" yaml:"name"`

	// The message of the error
	Message string `json:"message" yaml:"message"`

	// An error code associated with the error
	Code *string `json:"code" yaml:"code"`

	// An error stack trace
	Stack *string `json:"stack" yaml:"stack"`
}

// fields currently not in the schema files, but we see them used by the console
type Profile struct {
	DisplayName string `json:"displayName" yaml:"displayName"`
}

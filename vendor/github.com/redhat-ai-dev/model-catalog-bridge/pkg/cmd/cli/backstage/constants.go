package backstage

const (
	COMPONENT_TYPE = "model-server"
	RESOURCE_TYPE  = "ai-model"
	// we don't have a specific type in our ai model catalog defiition at this time, so any of the valid API formats which
	// backstage supports are possible:  openapi, asyncapi, graphql, grpc
	OPENAPI_API_TYPE   = "openapi"
	ASYNCAPI_API_TYPE  = "asyncapi"
	GRAPHQL_API_TYPE   = "graphql"
	GRPC_API_TYPE      = "grpc"
	TRPC_API_TYPE      = "trpc"
	UNKNOWN_API_TYPE   = "unknown"
	LINK_API_URL       = "API URL"
	LINK_TYPE_WEBSITE  = "website"
	LINK_ICON_WEBASSET = "WebAsset"
	TECHDOC_REFS       = "backstage.io/techdocs-ref"
	VERSION            = "backstage.io/v1alpha1"
)

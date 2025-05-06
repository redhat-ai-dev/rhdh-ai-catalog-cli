package rest

import "fmt"

const (
	BASE_URI      = "/api/catalog"
	LOCATION_URI  = "/locations"
	ENTITIES_URI  = "/entities"
	COMPONENT_URI = "/entities/by-name/component/%s/%s"
	RESOURCE_URI  = "/entities/by-name/resource/%s/%s"
	API_URI       = "/entities/by-name/api/%s/%s"
	QUERY_URI     = "/entities/by-query"
	DEFAULT_NS    = "default"
)

type BackstageImport interface {
	ImportLocation(url string) (map[string]any, error)
	DeleteLocation(id string) (string, error)
	GetLocation(id string) (map[string]any, error)
}

func ParseImportLocationMap(retJSON map[string]any) (id string, target string, ok bool) {
	var location interface{}
	location, ok = retJSON["location"]
	if ok {
		var locationMap map[string]interface{}
		locationMap, ok = location.(map[string]interface{})
		if ok {
			id = fmt.Sprintf("%s", locationMap["id"])
			target = fmt.Sprintf("%s", locationMap["target"])
			return id, target, ok
		}
	}
	var idi interface{}
	idi, ok = retJSON["id"]
	if ok {
		var targeti interface{}
		targeti, ok = retJSON["target"]
		if ok {
			id = fmt.Sprintf("%s", idi)
			target = fmt.Sprintf("%s", targeti)
			return id, target, ok
		}
	}

	return id, target, ok
}

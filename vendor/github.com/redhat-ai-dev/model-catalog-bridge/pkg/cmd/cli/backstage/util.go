package backstage

import (
	"fmt"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
	"net/url"
	"sort"
	"strings"
)

func buildKeys(args ...string) map[string][]string {
	keys := map[string][]string{}
	for _, arg := range args {
		array := strings.Split(arg, ":")
		if len(array) == 1 {
			arr := keys[rest.DEFAULT_NS]
			arr = append(arr, arg)
			keys[rest.DEFAULT_NS] = arr
			continue
		}
		arr := keys[array[0]]
		arr = append(arr, array[1])
		keys[array[0]] = arr
	}
	return keys
}

func (b *BackstageRESTClientWrapper) pullSavedArgsFromQueryParams(qparms *url.Values) []string {
	var argsArr []string
	if b.Tags {
		argsStr := qparms.Get("metadata.tags")
		argsArr = strings.Split(argsStr, " ")
		qparms.Del("metadata.tags")
	}
	return argsArr
}

func tagsIncluded(args, tags []string) bool {
	if len(tags) < len(args) {
		return false
	}
	// we don't require exact order with the set of tags specified so we sort the two arrays to facilitate the compare
	sort.Strings(args)
	sort.Strings(tags)
	for _, arg := range args {
		found := false
		for _, tag := range tags {
			if arg == tag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func tagsMatch(args, tags []string) bool {
	if len(args) != len(tags) {
		return false
	}
	// we don't require exact order with the set of tags specified so we sort the two arrays to facilitate the compare
	sort.Strings(args)
	sort.Strings(tags)
	for i, tag := range tags {
		if args[i] != tag {
			return false
		}
	}
	return true
}

func updateQParams(kind, specType string, args []string) *url.Values {
	// example 'filter' value from swagger doc:  'kind=component,metadata.annotations.backstage.io/orphan=true'
	// NOTE: while our AI model catalog proscribes specific types for Component and Resource, for API, there are a
	// set of predefined types related to the format of the API (openapi, grpc, etc.) so we have no means currently
	// of using spec.type to distinguish our AI model catalog API items like we do for Component and Resource.
	filterValue := fmt.Sprintf("kind=%s,spec.type=%s", kind, specType)
	if strings.EqualFold(kind, "api") {
		filterValue = fmt.Sprintf("kind=%s", kind)
	}
	qparams := &url.Values{
		"filter": []string{filterValue},
	}
	//TODO could not determine single query parameter format that resulted in returning
	// a list of entities filtered by both 'kind', 'spec.type', and `metadata.tags` array directly matched a provided list of args;
	// for now, we fetch all types, and filter based on tags afterwards
	return qparams
}

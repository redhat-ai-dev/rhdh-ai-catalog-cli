package backstage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
)

type listResources struct {
	Items      []ResourceEntityV1alpha1 `json:"items" yaml:"items"`
	TotalItems int                      `json:"totalItems" yaml:"totalItems"`
	PageInfo   interface{}              `json:"pageInfo" yaml:"pageInfo"`
}

func (b *BackstageRESTClientWrapper) ListResources(args ...string) (string, error) {
	qparms := updateQParams("resource", RESOURCE_TYPE, args)
	str, err := b.getWithKindParamFromBackstage(b.RootURL+rest.QUERY_URI, qparms)
	if err != nil {
		return "", err
	}

	buf := []byte(str)

	lr := &listResources{}
	err = json.Unmarshal(buf, lr)
	if err != nil {
		return str, err
	}

	//TODO remove this post query filter logic if an exact query parameter check for the 'metadata.tags' array is determined
	if b.Tags {
		filteredResources := []ResourceEntityV1alpha1{}
		for _, resource := range lr.Items {
			switch {
			case !b.Subset && tagsMatch(args, resource.Metadata.Tags):
				filteredResources = append(filteredResources, resource)
			case b.Subset && tagsIncluded(args, resource.Metadata.Tags):
				filteredResources = append(filteredResources, resource)
			}
		}
		lr.Items = filteredResources

	}

	buf, err = json.MarshalIndent(lr.Items, "", "    ")
	return string(buf), err
}

func (b *BackstageRESTClientWrapper) GetResource(args ...string) (string, error) {
	if len(args) == 0 || b.Tags {
		return b.ListResources(args...)
	}

	keys := buildKeys(args...)
	buffer := &bytes.Buffer{}
	for namespace, names := range keys {
		for _, name := range names {
			str, err := b.getFromBackstage(b.RootURL + fmt.Sprintf(rest.RESOURCE_URI, namespace, name))
			if err != nil {
				return buffer.String(), err
			}
			buf := []byte(str)
			err = json.Indent(buffer, buf, "", "    ")
			buffer.WriteString("\n")
		}
	}

	return buffer.String(), nil
}

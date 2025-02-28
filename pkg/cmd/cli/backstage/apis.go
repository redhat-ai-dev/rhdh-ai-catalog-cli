package backstage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
)

type listAPIs struct {
	Items      []ApiEntityV1alpha1 `json:"items" yaml:"items"`
	TotalItems int                 `json:"totalItems" yaml:"totalItems"`
	PageInfo   interface{}         `json:"pageInfo" yaml:"pageInfo"`
}

func (b *BackstageRESTClientWrapper) ListAPIs(args ...string) (string, error) {
	qparms := updateQParams("api", OPENAPI_API_TYPE, args)

	str, err := b.getWithKindParamFromBackstage(b.RootURL+rest.QUERY_URI, qparms)
	if err != nil {
		return "", err
	}

	buf := []byte(str)

	la := &listAPIs{}
	err = json.Unmarshal(buf, la)
	if err != nil {
		return str, err
	}

	//TODO remove this post query filter logic if an exact query parameter check for the 'metadata.tags' array is determined
	if b.Tags && !b.Subset {
		filteredAPIs := []ApiEntityV1alpha1{}
		for _, api := range la.Items {
			switch {
			case !b.Subset && tagsMatch(args, api.Metadata.Tags):
				filteredAPIs = append(filteredAPIs, api)
			case b.Subset && tagsIncluded(args, api.Metadata.Tags):
				filteredAPIs = append(filteredAPIs, api)
			}
		}
		la.Items = filteredAPIs
	}

	buf, err = json.MarshalIndent(la.Items, "", "    ")
	return string(buf), err
}

func (b *BackstageRESTClientWrapper) GetAPI(args ...string) (string, error) {
	if len(args) == 0 || b.Tags {
		return b.ListAPIs(args...)
	}

	keys := buildKeys(args...)
	buffer := &bytes.Buffer{}
	for namespace, names := range keys {
		for _, name := range names {
			str, err := b.getFromBackstage(b.RootURL + fmt.Sprintf(rest.API_URI, namespace, name))
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

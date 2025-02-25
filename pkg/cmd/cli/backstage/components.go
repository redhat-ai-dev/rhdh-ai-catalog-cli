package backstage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
)

type listComponents struct {
	Items      []ComponentEntityV1alpha1 `json:"items" yaml:"items"`
	TotalItems int                       `json:"totalItems" yaml:"totalItems"`
	PageInfo   interface{}               `json:"pageInfo" yaml:"pageInfo"`
}

func (b *BackstageRESTClientWrapper) ListComponents(args ...string) (string, error) {
	qparms := updateQParams("component", COMPONENT_TYPE, args)

	str, err := b.getWithKindParamFromBackstage(b.RootURL+rest.QUERY_URI, qparms)
	if err != nil {
		return str, err
	}

	buf := []byte(str)

	lc := &listComponents{}
	err = json.Unmarshal(buf, lc)
	if err != nil {
		return str, err
	}

	//TODO remove this post query filter logic if an exact query parameter check for the 'metadata.tags' array is determined
	if b.Tags && !b.Subset {
		filteredComponents := []ComponentEntityV1alpha1{}
		for _, component := range lc.Items {
			switch {
			case !b.Subset && tagsMatch(args, component.Metadata.Tags):
				filteredComponents = append(filteredComponents, component)
			case b.Subset && tagsIncluded(args, component.Metadata.Tags):
				filteredComponents = append(filteredComponents, component)
			}
		}
		lc.Items = filteredComponents
	}

	buf, err = json.MarshalIndent(lc.Items, "", "    ")
	return string(buf), err
}

func (b *BackstageRESTClientWrapper) GetComponent(args ...string) (string, error) {
	if len(args) == 0 || b.Tags {
		return b.ListComponents(args...)
	}

	keys := buildKeys(args...)
	buffer := &bytes.Buffer{}
	for namespace, names := range keys {
		for _, name := range names {
			str, err := b.getFromBackstage(b.RootURL + fmt.Sprintf(rest.COMPONENT_URI, namespace, name))
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

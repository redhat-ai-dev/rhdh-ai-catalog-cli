package backstage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
	"strings"
)

func (b *BackstageRESTClientWrapper) ListLocations() (string, error) {
	str, err := b.getFromBackstage(b.RootURL + rest.LOCATION_URI)
	if err != nil {
		return "", err
	}
	buf := []byte(str)
	buffer := &bytes.Buffer{}
	err = json.Indent(buffer, buf, "", "    ")
	return buffer.String(), err
}

func (b *BackstageRESTClientWrapper) GetLocation(args ...string) (string, error) {
	if len(args) == 0 {
		return b.ListLocations()
	}
	buffer := &bytes.Buffer{}
	for _, id := range args {
		str, err := b.getFromBackstage(b.RootURL + rest.LOCATION_URI + "/" + id)
		if err != nil {
			return buffer.String(), err
		}
		buf := []byte(str)
		err = json.Indent(buffer, buf, "", "    ")
		buffer.WriteString("\n")
	}
	return buffer.String(), nil
}

func (b *BackstageRESTClientWrapper) ImportLocation(url string) (map[string]any, error) {
	if strings.Contains(url, "github") {
		return b.postToBackstage(b.RootURL+rest.LOCATION_URI, map[string]interface{}{"target": url, "type": "url"})
	}
	return b.postToBackstage(b.RootURL+rest.LOCATION_URI, map[string]interface{}{"target": url, "type": "rhdh-rhoai-bridge"})
}

func (b *BackstageRESTClientWrapper) DeleteLocation(id string) (string, error) {
	return b.deleteFromBackstage(b.RootURL + rest.LOCATION_URI + "/" + id)
}

func (b *BackstageRESTClientWrapper) PrintImportLocation(retJSON map[string]any) (string, error) {
	var location interface{}
	var id interface{}
	var target interface{}
	var ok bool
	location, ok = retJSON["location"]
	if ok {
		locationMap, o1 := location.(map[string]interface{})
		if o1 {
			id = locationMap["id"]
			target = locationMap["target"]
		}
		return fmt.Sprintf("Backstage location %s from %s created", id, target), nil
	}
	id, ok = retJSON["id"]
	if ok {
		target, ok = retJSON["target"]
		if ok {
			return fmt.Sprintf("Backstage location %s from %s created", id, target), nil
		}
		return fmt.Sprintf("Backstage location %s created", id), nil
	}
	return fmt.Sprintf("%#v", retJSON), nil

}

func (b *BackstageRESTClientWrapper) ParseImportLocationMap(retJSON map[string]any) (id string, target string, ok bool) {
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

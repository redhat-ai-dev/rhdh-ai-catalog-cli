package backstage

import (
	"bytes"
	"encoding/json"
)

func (b *BackstageRESTClientWrapper) ListLocations() (string, error) {
	str, err := b.getFromBackstage(b.RootURL + LOCATION_URI)
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
		str, err := b.getFromBackstage(b.RootURL + LOCATION_URI + "/" + id)
		if err != nil {
			return buffer.String(), err
		}
		buf := []byte(str)
		err = json.Indent(buffer, buf, "", "    ")
		buffer.WriteString("\n")
	}
	return buffer.String(), nil
}

func (b *BackstageRESTClientWrapper) ImportLocation(url string) (string, error) {
	return b.postToBackstage(b.RootURL+LOCATION_URI, map[string]interface{}{"target": url, "type": "url"})
}

func (b *BackstageRESTClientWrapper) DeleteLocation(id string) (string, error) {
	return b.deleteFromBackstage(b.RootURL + LOCATION_URI + "/" + id)
}

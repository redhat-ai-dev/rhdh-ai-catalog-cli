package backstage

import (
	"bytes"
	"encoding/json"
)

func (b *BackstageRESTClientWrapper) ListEntities() (string, error) {
	str, err := b.getFromBackstage(b.RootURL + ENTITIES_URI)
	if err != nil {
		return "", err
	}

	buf := []byte(str)
	buffer := &bytes.Buffer{}
	err = json.Indent(buffer, buf, "", "    ")
	return buffer.String(), err
}

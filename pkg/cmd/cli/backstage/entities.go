package backstage

import (
	"bytes"
	"encoding/json"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
)

func (b *BackstageRESTClientWrapper) ListEntities() (string, error) {
	str, err := b.getFromBackstage(b.RootURL + rest.ENTITIES_URI)
	if err != nil {
		return "", err
	}

	buf := []byte(str)
	buffer := &bytes.Buffer{}
	err = json.Indent(buffer, buf, "", "    ")
	return buffer.String(), err
}

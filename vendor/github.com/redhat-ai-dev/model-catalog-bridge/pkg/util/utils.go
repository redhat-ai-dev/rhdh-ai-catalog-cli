package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	serverapiv1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"

	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/rest"
	"github.com/redhat-ai-dev/model-catalog-bridge/pkg/types"

	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	NameInvalidCharRegexp = `[^a-zA-Z0-9\-_]`

	NameNoDuplicateSpecialCharRegexp = `[-_.]{2,}`
)

func PrintYaml(obj interface{}, addDivider bool, w io.Writer) error {
	writer := printers.GetNewTabWriter(w)
	output, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = writer.Write(output)
	if addDivider {
		fmt.Fprintln(w, "---")
	}
	return err
}

func PrintJSON(obj interface{}, w io.Writer) error {
	output, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(output)
	return err
}

var (
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})
)

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func ProcessOutput(str string, err error) {
	klog.Infoln(str)
	klog.Flush()
	if err != nil {
		klog.Errorf("%s", err.Error())
		klog.Flush()
	}
}

func BuildImportKeyAndURI(seg1, seg2 string, format types.NormalizerFormat) (string, string) {
	// no spaces in keys
	seg1 = strings.ReplaceAll(seg1, " ", "")
	seg2 = strings.ReplaceAll(seg2, " ", "")
	fn := "catalog-info.yaml"
	if format == types.JsonArrayForamt {
		fn = "model-catalog.json"
	}
	return fmt.Sprintf("%s_%s", seg1, seg2), fmt.Sprintf("/%s/%s/%s", seg1, seg2, fn)
}

func SanitizeModelVersion(mv string) string {
	replacer := strings.NewReplacer(" ", "-")
	mv = strings.ToLower(mv)
	mv = replacer.Replace(mv)
	return strings.ToLower(SanitizeName(mv))
}
func SanitizeName(name string) string {
	sanitizedName := name

	// Replace any invalid characters with an empty character
	validChars := regexp.MustCompile(NameInvalidCharRegexp)
	sanitizedName = validChars.ReplaceAllString(sanitizedName, "")

	// Remove duplicated special characters
	noDupeChars := regexp.MustCompile(NameNoDuplicateSpecialCharRegexp)
	sanitizedName = noDupeChars.ReplaceAllString(sanitizedName, "")

	// Trim to no more than 63 characters
	if len(sanitizedName) > 63 {
		sanitizedName = sanitizedName[:63]
	}

	// Finally, ensure only alphanumeric characters at beginning and end of the name
	sanitizedName = strings.Trim(sanitizedName, "-_.")
	return sanitizedName

}

func KServeInferenceServiceMapping(rId, mId string, is *serverapiv1beta1.InferenceService) bool {
	if is.Labels == nil {
		return false
	}

	rmVal, rok := is.Labels[rest.INF_SVC_RM_ID_LABEL]
	if !rok {
		return false
	}

	if strings.TrimSpace(rId) != strings.TrimSpace(rmVal) {
		return false
	}

    mvVal, mok := is.Labels[rest.INF_SVC_MV_ID_LABEL]
    if !mok {
         return false
    }

    if strings.TrimSpace(mId) != strings.TrimSpace(mvVal) {
         return false
    }

    return true
}

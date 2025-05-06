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

func KServeInferenceServiceMapping(rName, mName, isName string) bool {
	// we have to special case the names a bit before sanitzing when mapping to the inference service name to match
	// what kubeflow / kserve does; our sanitize already converts dots to empty chars, but we also need to a) convert
	// spaces to hyphens, and b) make everything lower case
	replacer := strings.NewReplacer(" ", "-")
	rName = strings.ToLower(rName)
	rName = replacer.Replace(rName)
	rName = SanitizeName(rName)
	mName = strings.ToLower(mName)
	mName = replacer.Replace(mName)
	mName = SanitizeName(mName)
	key := fmt.Sprintf("%s-%s", rName, mName)
	return key == isName
}

package util

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"sigs.k8s.io/yaml"
	"syscall"
)

const ApplicationName = "bac"

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

var (
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})
)

func BuildYaml(obj interface{}, buf []byte, addDivider bool) ([]byte, error) {
	b := bytes.NewBuffer(buf)
	writer := printers.GetNewTabWriter(b)
	output, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}
	_, err = writer.Write(output)
	if addDivider {
		fmt.Fprintln(b, "---")
	}
	buf = append(buf, b.Bytes()...)
	return buf, nil
}

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

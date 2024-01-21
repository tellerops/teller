package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// ApiLog is a separate logrus logger for the API. The destination for
// its messages is controlled by the environment variable
// CONJURAPI_LOG. CONJRAPI_LOG can be "stdout", "stderr", or the path
// to a file. If it's a path, the file's contents will be overwritten
// with new messages. If the environment variable is not set, logging
// is disabled.
var ApiLog = logrus.New()

func init() {
	dest, ok := os.LookupEnv("CONJURAPI_LOG")
	if !ok {
		return
	}

	var (
		out io.Writer
		err error
	)
	switch dest {
	case "stdout":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	default:
		out, err = os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logrus.Fatalf("Failed to open %s: %v", dest, err.Error())
		}
		logrus.Infof("Logging to %s", dest)
	}

	ApiLog.Out = out
	ApiLog.Level = logrus.DebugLevel
}

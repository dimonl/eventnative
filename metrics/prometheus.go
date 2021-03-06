package metrics

import (
	"github.com/ksensehq/eventnative/logging"
	"strings"
)

var Enabled = false

func Init(enabled bool) {
	Enabled = enabled
	if Enabled {
		logging.Info("Initializing Prometheus metrics..")
		initEvents()
	} else {
		logging.Warnf("Metrics isn't enabled")
	}
}

func extractLabels(destinationName string) (projectId, destinationId string) {
	splitted := strings.Split(destinationName, ".")
	if len(splitted) > 1 {
		return splitted[0], splitted[1]
	}

	return "-", destinationName
}

package agent

import (
	"encoding/json"

	"github.com/go-logr/logr"
	clusterinfov1alpha1 "cluster-info-addon/api/v1alpha1"
)

func reportClusterInfo(log logr.Logger, clusterInfo clusterinfov1alpha1.ClusterInfoData) {
	payload, err := json.MarshalIndent(clusterInfo, "", "  ")
	if err != nil {
		log.Error(err, "unable to marshal collected cluster info for verbose output")
		return
	}

	log.Info("collected cluster info", "report", string(payload))
}

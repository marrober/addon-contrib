package agent

import (
	"context"
	"fmt"
	"sort"

	configv1 "github.com/openshift/api/config/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterinfov1alpha1 "cluster-info-addon/api/v1alpha1"
)

func collectClusterInfo(ctx context.Context, spokeClient client.Client, clusterName string) (clusterinfov1alpha1.ClusterInfoData, error) {
	clusterVersion, err := collectClusterVersion(ctx, spokeClient)
	if err != nil {
		return clusterinfov1alpha1.ClusterInfoData{}, err
	}

	clusterOperators, err := collectClusterOperators(ctx, spokeClient)
	if err != nil {
		return clusterinfov1alpha1.ClusterInfoData{}, err
	}

	installedOperators, err := collectInstalledOperators(ctx, spokeClient)
	if err != nil {
		return clusterinfov1alpha1.ClusterInfoData{}, err
	}

	return clusterinfov1alpha1.ClusterInfoData{
		ClusterName:        clusterName,
		ClusterVersion:     clusterVersion,
		ClusterOperators:   clusterOperators,
		InstalledOperators: installedOperators,
	}, nil
}

func collectClusterVersion(ctx context.Context, spokeClient client.Client) (clusterinfov1alpha1.ClusterVersionInfo, error) {
	clusterVersion := &configv1.ClusterVersion{}
	if err := spokeClient.Get(ctx, types.NamespacedName{Name: "version"}, clusterVersion); err != nil {
		return clusterinfov1alpha1.ClusterVersionInfo{}, fmt.Errorf("failed to get cluster version: %w", err)
	}

	info := clusterinfov1alpha1.ClusterVersionInfo{
		Version: clusterVersion.Status.Desired.Version,
		Status:  summarizeClusterVersionStatus(clusterVersion.Status.Conditions),
		Message: conditionMessage(clusterVersion.Status.Conditions, configv1.OperatorDegraded, configv1.OperatorProgressing, configv1.OperatorAvailable),
	}

	return info, nil
}

func collectClusterOperators(ctx context.Context, spokeClient client.Client) ([]clusterinfov1alpha1.ClusterOperatorInfo, error) {
	operatorList := &configv1.ClusterOperatorList{}
	if err := spokeClient.List(ctx, operatorList); err != nil {
		return nil, fmt.Errorf("failed to list cluster operators: %w", err)
	}

	operators := make([]clusterinfov1alpha1.ClusterOperatorInfo, 0, len(operatorList.Items))
	for _, operator := range operatorList.Items {
		operators = append(operators, clusterinfov1alpha1.ClusterOperatorInfo{
			Name:        operator.Name,
			Version:     clusterOperatorVersion(operator.Status.Versions),
			Available:   string(conditionValue(operator.Status.Conditions, configv1.OperatorAvailable)),
			Progressing: string(conditionValue(operator.Status.Conditions, configv1.OperatorProgressing)),
			Degraded:    string(conditionValue(operator.Status.Conditions, configv1.OperatorDegraded)),
			Status:      summarizeClusterOperatorStatus(operator.Status.Conditions),
			Message:     conditionMessage(operator.Status.Conditions, configv1.OperatorDegraded, configv1.OperatorProgressing, configv1.OperatorAvailable),
		})
	}

	sort.Slice(operators, func(i, j int) bool {
		return operators[i].Name < operators[j].Name
	})

	return operators, nil
}

func collectInstalledOperators(ctx context.Context, spokeClient client.Client) ([]clusterinfov1alpha1.InstalledOperatorInfo, error) {
	csvList := &operatorsv1alpha1.ClusterServiceVersionList{}
	if err := spokeClient.List(ctx, csvList); err != nil {
		return nil, fmt.Errorf("failed to list clusterserviceversions: %w", err)
	}

	operators := make([]clusterinfov1alpha1.InstalledOperatorInfo, 0, len(csvList.Items))
	for _, csv := range csvList.Items {
		operators = append(operators, clusterinfov1alpha1.InstalledOperatorInfo{
			Namespace: csv.Namespace,
			Name:      csv.Name,
			Version:   csv.Spec.Version,
			Phase:     string(csv.Status.Phase),
			Status:    summarizeCSVStatus(csv.Status.Phase, csv.Status.Reason),
			Message:   csv.Status.Message,
		})
	}

	sort.Slice(operators, func(i, j int) bool {
		if operators[i].Namespace == operators[j].Namespace {
			return operators[i].Name < operators[j].Name
		}
		return operators[i].Namespace < operators[j].Namespace
	})

	return operators, nil
}

func clusterOperatorVersion(versions []configv1.OperandVersion) string {
	for _, version := range versions {
		if version.Name == "operator" {
			return version.Version
		}
	}
	if len(versions) > 0 {
		return versions[0].Version
	}
	return ""
}

func conditionValue(conditions []configv1.ClusterOperatorStatusCondition, conditionType configv1.ClusterStatusConditionType) configv1.ConditionStatus {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}
	return configv1.ConditionUnknown
}

func conditionMessage(conditions []configv1.ClusterOperatorStatusCondition, types ...configv1.ClusterStatusConditionType) string {
	for _, conditionType := range types {
		for _, condition := range conditions {
			if condition.Type != conditionType {
				continue
			}
			if condition.Status == configv1.ConditionTrue || condition.Status == configv1.ConditionFalse {
				if condition.Message != "" {
					return condition.Message
				}
				if condition.Reason != "" {
					return condition.Reason
				}
			}
		}
	}
	return ""
}

func summarizeClusterOperatorStatus(conditions []configv1.ClusterOperatorStatusCondition) string {
	if conditionValue(conditions, configv1.OperatorDegraded) == configv1.ConditionTrue {
		return "Degraded"
	}
	if conditionValue(conditions, configv1.OperatorAvailable) == configv1.ConditionFalse {
		return "Unavailable"
	}
	if conditionValue(conditions, configv1.OperatorProgressing) == configv1.ConditionTrue {
		return "Progressing"
	}
	if conditionValue(conditions, configv1.OperatorAvailable) == configv1.ConditionTrue {
		return "Available"
	}
	return "Unknown"
}

func summarizeClusterVersionStatus(conditions []configv1.ClusterOperatorStatusCondition) string {
	if conditionValue(conditions, configv1.OperatorDegraded) == configv1.ConditionTrue {
		return "Degraded"
	}
	if conditionValue(conditions, configv1.OperatorProgressing) == configv1.ConditionTrue {
		return "Progressing"
	}
	if conditionValue(conditions, configv1.OperatorAvailable) == configv1.ConditionTrue {
		return "Available"
	}
	if conditionValue(conditions, configv1.OperatorAvailable) == configv1.ConditionFalse {
		return "Unavailable"
	}
	return "Unknown"
}

func summarizeCSVStatus(phase operatorsv1alpha1.ClusterServiceVersionPhase, reason string) string {
	switch phase {
	case operatorsv1alpha1.CSVPhaseSucceeded:
		return "Succeeded"
	case operatorsv1alpha1.CSVPhaseFailed:
		if reason != "" {
			return fmt.Sprintf("Failed (%s)", reason)
		}
		return "Failed"
	case operatorsv1alpha1.CSVPhasePending:
		return "Pending"
	case operatorsv1alpha1.CSVPhaseInstalling:
		return "Installing"
	case operatorsv1alpha1.CSVPhaseReplacing:
		return "Replacing"
	case operatorsv1alpha1.CSVPhaseDeleting:
		return "Deleting"
	default:
		if phase == "" {
			return "Unknown"
		}
		return string(phase)
	}
}

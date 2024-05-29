package upgraders

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	upgradev1alpha1 "github.com/openshift/managed-upgrade-operator/api/v1alpha1"
	"github.com/openshift/managed-upgrade-operator/pkg/metrics"
)

// PreUpgradeHealthCheck performs cluster healthy check
func (c *clusterUpgrader) PreUpgradeHealthCheck(ctx context.Context, logger logr.Logger) (bool, error) {
	upgradeCommenced, err := c.cvClient.HasUpgradeCommenced(c.upgradeConfig)
	if err != nil {
		return false, err
	}
	if upgradeCommenced {
		logger.Info(fmt.Sprintf("Skipping upgrade step %s", upgradev1alpha1.UpgradePreHealthCheck))
		return true, nil
	}

	ok, err := CriticalAlerts(c.metrics, c.config, c.upgradeConfig, logger)
	if err != nil || !ok {
		return false, err
	}

	ok, err = ClusterOperators(c.metrics, c.cvClient, c.upgradeConfig, logger)
	if err != nil || !ok {
		return false, err
	}

	if c.upgradeConfig.Spec.CapacityReservation {
		ok, err := c.scaler.CanScale(c.client, logger)
		if err != nil || !ok {
			c.metrics.UpdateMetricHealthcheckFailed(c.upgradeConfig.Name, metrics.DefaultWorkerMachinepoolNotFound)
			return false, err
		}
	}

	c.metrics.UpdateMetricHealthcheckSucceeded(c.upgradeConfig.Name)
	return true, nil
}

// PostUpgradeHealthCheck performs cluster healthy check
func (c *clusterUpgrader) PostUpgradeHealthCheck(ctx context.Context, logger logr.Logger) (bool, error) {
	ok, err := CriticalAlerts(c.metrics, c.config, c.upgradeConfig, logger)
	if err != nil || !ok {
		return false, err
	}

	ok, err = ClusterOperators(c.metrics, c.cvClient, c.upgradeConfig, logger)
	if err != nil || !ok {
		return false, err
	}
	c.metrics.UpdateMetricHealthcheckSucceeded(c.upgradeConfig.Name)
	return true, nil
}

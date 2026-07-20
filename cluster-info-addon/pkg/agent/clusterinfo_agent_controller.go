package agent

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterinfov1alpha1 "cluster-info-addon/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const syncInterval = 60 * time.Second

type ClusterInfoController struct {
	spokeClient client.Client
	hubClient   client.Client
	log         logr.Logger
	clusterName string
	verbose     bool
}

func (c *ClusterInfoController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterinfov1alpha1.ClusterInfo{}).
		Complete(c)
}

func (c *ClusterInfoController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	c.log.Info(fmt.Sprintf("reconciling... %s", req))
	defer c.log.Info(fmt.Sprintf("done reconcile %s", req))

	clusterInfoCR := clusterinfov1alpha1.ClusterInfo{}
	err := c.spokeClient.Get(ctx, req.NamespacedName, &clusterInfoCR)
	switch {
	case errors.IsNotFound(err):
		return ctrl.Result{}, nil
	case err != nil:
		c.log.Error(err, "unable to get ClusterInfo")
		return ctrl.Result{}, err
	}

	clusterInfoData, err := collectClusterInfo(ctx, c.spokeClient, c.clusterName)
	if err != nil {
		c.log.Error(err, "unable to collect cluster version and operator info from spoke cluster")
		return ctrl.Result{RequeueAfter: syncInterval}, err
	}

	if c.verbose {
		reportClusterInfo(c.log, clusterInfoData)
	}

	updatedStatus := clusterInfoCR.Status.DeepCopy()
	updatedStatus.ClusterInfo = clusterInfoData
	updatedStatus.LastSync = metav1.Now()

	if !reflect.DeepEqual(clusterInfoCR.Status, *updatedStatus) {
		clusterInfoCR.Status = *updatedStatus
		if err = c.spokeClient.Status().Update(ctx, &clusterInfoCR); err != nil {
			c.log.Error(err, "unable to update spoke ClusterInfo status with cluster info")
			return ctrl.Result{RequeueAfter: syncInterval}, err
		}
	}

	hubClusterInfo := clusterinfov1alpha1.ClusterInfo{}
	err = c.hubClient.Get(ctx, types.NamespacedName{Namespace: c.clusterName, Name: clusterInfoCR.Name}, &hubClusterInfo)
	switch {
	case errors.IsNotFound(err):
		hubClusterInfo.Name = clusterInfoCR.Name
		hubClusterInfo.Namespace = c.clusterName
		if err = c.hubClient.Create(ctx, &hubClusterInfo); err != nil {
			c.log.Error(err, "unable to create hub ClusterInfo")
			return ctrl.Result{RequeueAfter: syncInterval}, err
		}
		hubClusterInfo.Status = clusterInfoCR.Status
		err = c.hubClient.Status().Update(ctx, &hubClusterInfo)
		if err != nil {
			c.log.Error(err, "unable to update hub ClusterInfo status after create")
			return ctrl.Result{RequeueAfter: syncInterval}, err
		}
	case err != nil:
		c.log.Error(err, "unable to get hub ClusterInfo")
		return ctrl.Result{RequeueAfter: syncInterval}, err
	default:
		hubClusterInfo.Status = clusterInfoCR.Status
		err = c.hubClient.Status().Update(ctx, &hubClusterInfo)
		if err != nil {
			c.log.Error(err, "unable to update hub ClusterInfo")
			return ctrl.Result{RequeueAfter: syncInterval}, err
		}
	}

	return ctrl.Result{RequeueAfter: syncInterval}, nil
}

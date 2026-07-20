# cluster-info-addon

 cluster-info-addon is an example addon that showcase the capability of syncing a CR from the managed cluster to the hub cluster. The agent periodically collects the OpenShift cluster version, cluster operator versions/status, and OLM-installed operator versions/status (equivalent to `oc get csv -A`) from each managed cluster and reports the results to the hub.

## Install the cluster-info-addon to the Hub cluster

Switch context to Hub cluster.

```
make deploy
```

You can check the addon manager status by:
```
$ kubectl -n open-cluster-management get deploy cluster-info-addon-manager
NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
cluster-info-addon-manager   1/1     1            1           2m17s

kubectl -n cluster1 get managedclusteraddon cluster-info-addon # Replace 'cluster1' with the managed cluster name
NAME               AVAILABLE   DEGRADED   PROGRESSING
cluster-info-addon   True                   
```

## Verify the cluster-info-addon agent is installed on the Managed cluster and create a ClusterInfo CR

Switch context to Managed cluster.

```
$ kubectl -n open-cluster-management-agent-addon get deploy cluster-info-addon-agent
NAME                     READY   UP-TO-DATE   AVAILABLE   AGE
cluster-info-addon-agent   1/1     1            1           4m23s
```

```
make deploy-clusterinfo-cr-sample
```

## Verify the ClusterInfo CR is created on the Hub cluster

Switch context to Hub cluster.

```
$ kubectl -n cluster1 get clusterinfos # Replace 'cluster1' with the managed cluster name
NAME          AGE
clusterinfo   5m35s
```

## Verify cluster version and operator information is synced to the Hub cluster

After the ClusterInfo CR is created on the managed cluster, the agent collects cluster version and operator information every 60 seconds and syncs it to the hub.

Switch context to Hub cluster.

```
$ kubectl -n cluster1 get clusterinfos clusterinfo -o yaml
```

The status will include a single `clusterInfo` JSON object containing the cluster version, cluster operators, and installed OLM operators:

```
status:
  clusterInfo:
    clusterName: cluster1
    clusterVersion:
      version: 4.16.12
      status: Available
    clusterOperators:
    - name: authentication
      version: 4.16.12
      available: "True"
      progressing: "False"
      degraded: "False"
      status: Available
    installedOperators:
    - namespace: openshift-operators
      name: advanced-cluster-management.v2.12.0
      version: 2.12.0
      phase: Succeeded
      status: Succeeded
  lastSync: "2026-07-07T10:00:00Z"
```

## Update the ClusterInfo status on the Managed cluster (optional)

Using a tool such as [kubectl-edit-status](https://github.com/ulucinar/kubectl-edit-status), 
modify the ClusterInfo CR status to have the following:

```
status:
  spokeURL: hello
```

## Verify the  ClusterInfo CR status is updated on the Hub cluster

```
$ kubectl -n cluster1 get clusterinfos clusterinfo -o yaml
apiVersion: clusterinfo.open-cluster-management.io/v1alpha1
kind: ClusterInfo
metadata:
...
  name: clusterinfo
  namespace: cluster1
...
status:
  spokeURL: hello
```

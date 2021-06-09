## About this experiment

This experiment deploys the zfs-localpv provisioner in kube-system namespace which includes zfs-controller statefulset (with default value of replica count 1) and csi-node agent deamonset. Apart from this, zpool creation on nodes and generic use-case storage-classes and snapshot class for dynamic provisioning of the volumes based on values provided via env's in run_e2e_test.yml file gets created in this experiment.

## Supported platforms:

K8s: 1.18+

OS: Ubuntu, CentOS

ZFS: 0.7, 0.8

## Entry-Criteria

- K8s cluster should be in healthy state including all the desired worker nodes in ready state.
- External disk should be attached to the nodes for zpool creation on top of it.
- If we dont use this experiment to deploy zfs-localpv provisioner, we can directly apply the zfs-operator file via command as mentioned below and make sure you have zpool created on desired nodes to provision volumes.
```kubectl apply -f https://openebs.github.io/charts/zfs-operator.yaml```

## Exit-Criteria

- zfs-localpv components should be deployed successfully and all the pods including zfs-controller and csi node-agent daemonset are in running state.

## Steps performed

- zpool creation on nodes
  - if `ZPOOL_CREATION` env value is set to `true` zpool is created on the nodes.
  - selection of nodes on which zpool will be created, is taken via the values of     `ZPOOL_NODE_NAME` env. if it is blank then zpool will be created on all worker nodes.
  - selected nodes will be labeled (if all nodes are used then labeling will be skipped as it is unnecessary) so that a privileged daemoset pods can schedule on those nodes and can create zpool on respected nodes by executing zpool create command via daemonset pods.
  - Delete the daemonset and remove label from nodes after zpool creation.
- Download the operator file for zfs-localpv driver from `ZFS_BRANCH`.
- Update the zfs-operator namespace if it is specified other than default value `openebs` in `ZFS_OPERATOR_NAMESPACE` env.
- Update the zfs-driver image tag. (if specified other than ci tag)
- Apply the operator yaml and wait for zfs-controller and csi-node agent pods to come up in Running state.
- Create general use case storage_classes for dynamic volume provisioning.
- Create one volumesnapshot class for capturing zfs volume snapshot.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of deploying zfs-localpv provisioner, clone openebs/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then first apply rbac and crds for e2e-framework.
```
kubectl apply -f zfs-localpv/e2e-tests/hack/rbac.yaml
kubectl apply -f zfs-localpv/e2e-tests/hack/crds.yaml
```
then update the needed test specific values in run_e2e_test.yml file and create the kubernetes job.
```
kubectl create -f run_e2e_test.yml
```
All the env variables description is provided with the comments in the same file.

After creating kubernetes job, when the job’s pod is instantiated, we can see the logs of that pod which is executing the test-case.

```
kubectl get pods -n e2e
kubectl logs -f <zfs-localpv-provisioner-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er zfs-localpv-provisioner -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er zfs-localpv-provisioner -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```
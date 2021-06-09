## About this experiment

This experiment creates the volume snapshot of zfs-localpv which can be used further for creating a clone. Snapshot will be created in the same namespace where application pvc is created. One thing need to be noted that this experiment scale down the application before taking the snapshot, it is done this way to create the application consistent volume snapshot. After creating the snapshot application will be scaled up again.

## Supported platforms:

K8s : 1.18+

OS : Ubuntu, CentOS

ZFS : 0.7, 0.8

## Entry-Criteria

- K8s cluster should be in healthy state including all desired nodes in ready state.
- zfs-controller and node-agent daemonset pods should be in running state.
- Application should be deployed succesfully consuming the zfs-localpv storage.
- Volume snapshot class of zfs csi driver should be present to create the snapshot.

## Steps performed

This experiment consist of provisioning and deprovisioing of volume snapshot but performs one task at a time based on ACTION env value < provision or deprovision >.

Provision: 

- Check the application pod status, should be in running state.
- If DATA_PERSISTENCT check is enabled then dump some data into application pod mount point.
- Check if volume snapshot class is present.
- Scale down the application and wait till pod terminates successfully.
- Create the volume snapshot in the application namespace itself.
- Check the created snapshot resource and make sure readyToUse field is true.
- Scale up the application again.

Deprovision: 

- Delete the volume snapshot from the application namespace.
- Verify that volume snapshot content is no longer present.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of zfspv snapshot, clone openebs/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then first apply rbac and crds for e2e-framework.
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
kubectl logs -f <zfspv-snapshot-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er zfspv-snapshot -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er zfspv-snapshot -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```
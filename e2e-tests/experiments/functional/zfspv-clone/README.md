## About this experiment

This experiment creates the clone from the volume snapshot created before and use that cloned volume for some application. This experiment verifies that clone volume should be provisioned on the same node where original volume was provisioned. Apart from this, it verifies that clone volume should have the same data for which snaphsot was taken and this data should be easily accessible from some new application when this clone volume is mounted on it.

## Supported platforms:

K8s : 1.18+

OS : Ubuntu, CentOS

ZFS : 0.7, 0.8

Please note that for kubernetes version less than 1.17, `VolumeSnapshotDataSource` feature gate needs to be enabled at kubelet and kube-apiserver.

## Entry-Criteria

- K8s cluster should be in healthy state including all desired nodes in ready state.
- zfs-controller and node-agent daemonset pods should be in running state.
- volume snapshot should be present and should be in ready to use state.
- volume snapshot should be in the same namespace where application and pvc are present.
- size for the clone-pvc should be equal to the original pvc.

## Steps performed

This experiment consist of provisioning and deprovisioing of zfspv-clone but performs one task at a time based on ACTION env value < provision or deprovision >.

Provision:

- Check if the volume snapshot is present.
- Create the clone by applying the pvc yaml with snapshot name in the datasource.
- Verify that pvc gets bound.
- Deploy new application and verifies that clone volume gets successully mounted on application.
- Verify the data consistency that it should contain the same data as of volume snapshot.

Deprovision:

- Delete the application which is using the cloned volume.
- Verify that clone pvc is deleted successfully.
- Verify that zvolume is deleted successfully.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of zfs-localpv clone creation, first clone openens/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then apply rbac and crds for e2e-framework.
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
kubectl logs -f <zfspv-clone-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er zfspv-clone -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er zfspv-clone -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```
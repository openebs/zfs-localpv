## About this experiment

This experiment verifies that zvolume properties are same as set via the stoarge-class.

## Supported platforms:

K8s : 1.18+

OS : Ubuntu, CentOS

ZFS : 0.7, 0.8

## Entry-Criteria

- K8s cluster should be in healthy state including all desired nodes in ready state.
- zfs-controller and node-agent daemonset pods should be in running state.

## Steps performed

- Get the zvolume name and the storage class name by which volume was provisioned.
- After that following properties are verified to be same from zvol properties as well as from storage class.
  1. File-system type
  2. Compression
  3. Dedup
  4. Recordsize / volblocksize

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of zv properties verify, clone openens/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then first apply rbac and crds for e2e-framework.
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
kubectl logs -f <zv-properties-verify-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er zv-properties-verify -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er zv-properties-verify -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```
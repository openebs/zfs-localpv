## About the experiment

- This functional test verifies the zfs-localpv shared mount volume support via multiple pods. Applications who wants to share the volume can use the storage-class as below.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
parameters:
  shared: "yes"
  fstype: "zfs"
  poolname: "< zpool_name >"
provisioner: zfs.csi.openebs.io
```
Note: For running this experiment above storage-class should be present. This storage will be created as a part of zfs-localpv provisioner experiment. If zfs-localpv components are not deployed using e2e-test script located at `openebs/zfs-localpv/e2e-tests/experiment/zfs-localpv-provisioiner` please make sure you create the storage class from above mentioned yaml.

## Supported platforms:

K8s : 1.18+

OS : Ubuntu, CentOS

ZFS : 0.7, 0.8

## Entry-Criteria

- K8s cluster should be in healthy state including all the nodes in ready state.
- zfs-controller and node-agent daemonset pods should be in running state.
- storage class with `shared: yes` enabled should be present.

## Steps performed in this experiment:

1. First deploy the busybox application using `shared: yes` enabled storage-class
2. Then we dump some dummy data into the application pod mount point.
3. Scale the busybox deployment replicas so that multiple pods (here replicas = 2) can share the volume.
4. After that data consistency is verified from the scaled application pod in the way that data is accessible from both the pods and after restarting the application pod data consistency should be maintained.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of zfspv shared mount, first clone openens/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then apply rbac and crds for e2e-framework.
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
kubectl logs -f <zfspv-shared-mount-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er zfspv-shared-mount -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er zfspv-shared-mount -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```
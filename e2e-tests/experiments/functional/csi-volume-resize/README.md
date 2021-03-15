## About this experiment

This experiment verifies the csi volume resize feature of zfs-localpv. For resizing the volume we just need to update the pvc yaml with desired size and apply it. We can directly edit the pvc by ```kubectl edit pvc <pvc_name> -n <namespace>``` command and update the spec.resources.requests.storage field with desired volume size. One thing need to be noted that volume resize can only be done from lower pvc size to higher pvc size. We can not resize the volume from higher pvc size to lower one, in-short volume shrink is not possible. zfs driver supports online volume expansion, so that for using the resized volume, application pod restart is not required. For resize, storage-class which will provision the pvc should have `allowVolumeExpansion: true` field.

for e.g.
```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: zfspv-sc
allowVolumeExpansion: true
parameters:
  poolname: "zfs-test-pool"
provisioner: zfs.csi.openebs.io
```

## Supported platforms:

K8s : 1.18+

OS : Ubuntu, CentOS

ZFS : 0.7, 0.8

## Entry-criteria

- K8s cluster should be in healthy state including desired worker nodes in ready state.
- zfs-controller and csi node-agent daemonset pods should be in running state.
- storage class with `allowVolumeExpansion: true` enable should be present.
- Application should be deployed succesfully consuming the zfs-localpv storage.

## Exit-criteria

- Volume should be resized successfully and application should be accessible seamlessly.
- Application should be able to use the new resize volume space.

## Steps performed

- Check pvc status, it should be Bound. Get storage class name and capacity size of this pvc.
- Update the pvc size with desired volume size, which should not be lesser than previous volume size because volume shrink is not supported.
- Check the updated size in pvc spec.
- Since it is online volume expansion, we don't need to restart the application pod but here we restart intentionally to validate that resized space is available after restart of application pod.
- To use the resized space this test will dump some dummy data at application mount point. This dummy data size will be previous volume size + 1 Gi. So make sure we have this much enough space.
- At last this test will delete the dummy data files to free the space.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of csi volume resize, clone openens/zfs-localpv[https://github.com/openebs/zfs-localpv] repo and then first apply rbac and crds for e2e-framework.
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
kubectl logs -f <csi-volume-resize-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er csi-volume-resize -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er csi-volume-resize -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```
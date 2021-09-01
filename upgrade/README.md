From zfs-driver 0.6 version, the ZFS-LocalPV related CRs are now grouped together in its own group called `zfs.openebs.io`. So if we are using the driver of version less than 0.6 and want to upgrade to 0.6 release or later then we have to follow these steps. If we are already using the ZFS-LocalPV version greater or equal to 0.6 then we just have to apply the yaml from the release branch to upgrade.

So if my current version is 0.6 and want to upgrade to 0.7, then we can just do this to upgrade :-

```
$ kubectl apply -f https://raw.githubusercontent.com/openebs/zfs-localpv/v0.7.x/deploy/zfs-operator.yaml
```

And if my current version is 0.2 and want to upgrade to 0.5, then we can just do this to upgrade :-

```
$ kubectl apply -f https://raw.githubusercontent.com/openebs/zfs-localpv/v0.5.x/deploy/zfs-operator.yaml
```

And if my current version is 0.4 and want to upgrade to 0.7, that means we want to upgrade to version greater than 0.6, then we have to follow all the steps mentioned here :-

*Prerequisite*

Please do not provision/deprovision any volumes during the upgrade, if we can not control it, then we can scale down the openebs-zfs-controller stateful set to zero replica which will pause all the provisioning/deprovisioning request. And once upgrade is done, the upgraded Driver will continue the provisioning/deprovisioning process.

```
$ kubectl edit sts openebs-zfs-controller -n kube-system

```
And set replicas to zero :

```
spec:
  podManagementPolicy: OrderedReady
    *replicas: 0*
      revisionHistoryLimit: 10
```

After this, the controller pod openebs-zfs-controller-x in kube-system namespace will be terminated. Now all the volume provisioning requets will be halted. And it will be recreated as a part of step 3, which will upgrade the image to latest release and also set replicas to the 1(default) which will recreat the controller pod in kube-system and volume provisioning will resume on the upgraded system.

steps to upgrade:-

1. *Apply the new CRD*

```
$ kubectl apply -f upgrade/crd.yaml
customresourcedefinition.apiextensions.k8s.io/zfsvolumes.zfs.openebs.io created
customresourcedefinition.apiextensions.k8s.io/zfssnapshots.zfs.openebs.io created
```

2. *run upgrade.sh*

```
$ sh upgrade/upgrade.sh
zfsvolume.zfs.openebs.io/pvc-086a8608-9057-42df-b684-ee4ae8d35f71 created
zfsvolume.zfs.openebs.io/pvc-5286d646-c93d-4413-9707-fd95ebaae8c0 created
zfsvolume.zfs.openebs.io/pvc-74abefb8-8423-4b13-a607-7184ef088fb5 created
zfsvolume.zfs.openebs.io/pvc-82368c44-eee8-47ee-85a6-633a8023faa8 created
zfssnapshot.zfs.openebs.io/snapshot-dc61a056-f495-482b-8e6e-e7ddc4c13f47 created
zfssnapshot.zfs.openebs.io/snapshot-f9db91ea-529e-4dac-b2b8-ead045c612da created
```
Please note that if you have modified the OPENEBS_NAMESPACE env in the driver's deployment to other namespace. Then you have to pass the namespace as an argument to the upgrade.sh script `sh upgrade/upgrash.sh [namespace]`.


3. *upgrade the driver*

We can now upgrade the driver to the desired release. For example, to upgrade to v0.6, we can apply the below yaml, which will upgrade the driver to 0.6 release.

```
$ kubectl apply -f https://raw.githubusercontent.com/openebs/zfs-localpv/v0.6.x/deploy/zfs-operator.yaml
```

For future releases if you want to upgrade from v0.4 or v0.5 to the newer version replace `v0.6.x` to the desired version. Check everything is good after upgrading the zfs-driver. Then run the cleanup script to remove old CRDs


4. *run cleanup.sh*

```
$ sh upgrade/cleanup.sh
Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply
zfsvolume.openebs.io/pvc-086a8608-9057-42df-b684-ee4ae8d35f71 configured
Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply
zfsvolume.openebs.io/pvc-5286d646-c93d-4413-9707-fd95ebaae8c0 configured
Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply
zfsvolume.openebs.io/pvc-74abefb8-8423-4b13-a607-7184ef088fb5 configured
Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply
  svolume.openebs.io/pvc-82368c44-eee8-47ee-85a6-633a8023faa8 configured
zfsvolume.openebs.io "pvc-086a8608-9057-42df-b684-ee4ae8d35f71" deleted
zfsvolume.openebs.io "pvc-5286d646-c93d-4413-9707-fd95ebaae8c0" deleted
zfsvolume.openebs.io "pvc-74abefb8-8423-4b13-a607-7184ef088fb5" deleted
zfsvolume.openebs.io "pvc-82368c44-eee8-47ee-85a6-633a8023faa8" deleted
customresourcedefinition.apiextensions.k8s.io "zfsvolumes.openebs.io" deleted
Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply
zfssnapshot.openebs.io/snapshot-dc61a056-f495-482b-8e6e-e7ddc4c13f47 configured
Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply
zfssnapshot.openebs.io/snapshot-f9db91ea-529e-4dac-b2b8-ead045c612da configured
zfssnapshot.openebs.io "snapshot-dc61a056-f495-482b-8e6e-e7ddc4c13f47" deleted
zfssnapshot.openebs.io "snapshot-f9db91ea-529e-4dac-b2b8-ead045c612da" deleted
customresourcedefinition.apiextensions.k8s.io "zfssnapshots.openebs.io" deleted
```

Please note that if you have modified the OPENEBS_NAMESPACE env in the driver's deployment to other namespace. Then you have to pass the namespace as an argument to the cleanup.sh script `sh upgrade/cleanup.sh [namespace]`.

5. *restart kube-controller [optional]*

kube-controller-manager might be using stale volumeattachment resources, it might get flooded with the error logs. Restarting kube-controller will fix it.

### *Note*

While upgrading zfs-driver from v1.9.1 to later version by applying zfs-operator file, we may get this error.
```
The CSIDriver "zfs.csi.openebs.io" is invalid: spec.storageCapacity: Invalid value: true: field is immutable
```
It occurs due to newly added field `storageCapacity: true` in csi driver spec. In that case, first delete the csi-driver by running this command:
```
$ kubectl delete csidriver zfs.csi.openebs.io 
```
Now we can again apply the operator yaml file.
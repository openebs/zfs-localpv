From zfs-driver:v0.6 version ZFS-LocalPV related CRs are now grouped together in its own group called `zfs.openebs.io`. Here steps are mentioned for how to upgrade for refactoring the CRDs.

steps to upgrade:-

1. Apply the new CRD

```
$ kubectl apply -f upgrade/crd.yaml
customresourcedefinition.apiextensions.k8s.io/zfsvolumes.zfs.openebs.io created
customresourcedefinition.apiextensions.k8s.io/zfssnapshots.zfs.openebs.io created
```

2. run upgrade.sh

```
$ sh upgrade/upgrade.sh
zfsvolume.zfs.openebs.io/pvc-086a8608-9057-42df-b684-ee4ae8d35f71 created
zfsvolume.zfs.openebs.io/pvc-5286d646-c93d-4413-9707-fd95ebaae8c0 created
zfsvolume.zfs.openebs.io/pvc-74abefb8-8423-4b13-a607-7184ef088fb5 created
zfsvolume.zfs.openebs.io/pvc-82368c44-eee8-47ee-85a6-633a8023faa8 created
zfssnapshot.zfs.openebs.io/snapshot-dc61a056-f495-482b-8e6e-e7ddc4c13f47 created
zfssnapshot.zfs.openebs.io/snapshot-f9db91ea-529e-4dac-b2b8-ead045c612da created
```
`
3. upgrade the driver to v0.6

```
$ kubectl apply -f https://github.com/openebs/zfs-localpv/blob/v0.6.x/deploy/zfs-operator.yaml
```

For future releases if you want to upgrade from v0.4 or v0.5 to the newer version replace `v0.6.x` to the desired version. Check everything is good after upgrading the zfs-driver. Then run the cleanup script to remove old CRDs


4. run cleanup.sh

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
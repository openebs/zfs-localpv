# OpenEBS ZFS LocalPV Helm Repository

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

[Helm3](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```bash
$ helm repo add openebs-zfslocalpv https://openebs.github.io/zfs-localpv
```

You can then run `helm search repo openebs-zfslocalpv` to see the charts.

#### Update OpenEBS ZFS LocalPV Repo

Once OpenEBS ZFS Localpv repository has been successfully fetched into the local system, it has to be updated to get the latest version. The ZFS LocalPV charts repo can be updated using the following command.

```bash
helm repo update
```

#### Install using Helm 3

- Assign openebs namespace to the current context:
```bash
kubectl config set-context <current_context_name> --namespace=openebs
```

- If namespace is not created, run the following command
```bash
helm install <your-relase-name> openebs-zfslocalpv/zfs-localpv --create-namespace
```
- Else, if namespace is already created, run the following command
```bash
helm install <your-relase-name> openebs-zfslocalpv/zfs-localpv
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
# Helm
$ helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
# Helm
$ helm upgrade [RELEASE_NAME] [CHART] --install
```


## Configuration

The following table lists the configurable parameters of the OpenEBS ZFS LocalPV Provisioner chart and their default values.

| Parameter| Description| Default|
| -| -| -|
| `imagePullSecrets`| Provides image pull secrect| `""`|
| `zfsPlugin.image.registry`| Registry for openebs-zfs-plugin image| `""`|
| `zfsPlugin.image.repository`| Image repository for openebs-zfs-plugin| `openebs/zfs-driver`|
| `zfsPlugin.image.pullPolicy`| Image pull policy for openebs-zfs-plugin| `IfNotPresent`|
| `zfsPlugin.image.tag`| Image tag for openebs-zfs-plugin| `1.1.0`|
| `zfsNode.driverRegistrar.image.registry`| Registry for csi-node-driver-registrar image| `quay.io/`|
| `zfsNode.driverRegistrar.image.repository`| Image repository for csi-node-driver-registrar| `k8scsi/csi-node-driver-registrar`|
| `zfsNode.driverRegistrar.image.pullPolicy`| Image pull policy for csi-node-driver-registrar| `IfNotPresent`|
| `zfsNode.driverRegistrar.image.tag`| Image tag for csi-node-driver-registrar| `v1.2.0`|
| `zfsNode.updateStrategy.type`| Update strategy for zfsnode daemonset | `RollingUpdate` |
| `zfsNode.kubeletDir`| Kubelet mount point for zfsnode daemonset| `"/var/lib/kubelet/"` |
| `zfsNode.annotations` | Annotations for zfsnode daemonset metadata| `""`|
| `zfsNode.podAnnotations`| Annotations for zfsnode daemonset's pods metadata | `""`|
| `zfsNode.resources`| Resource and request and limit for zfsnode daemonset containers | `""`|
| `zfsNode.labels`| Labels for zfsnode daemonset metadata | `""`|
| `zfsNode.podLabels`| Appends labels to the zfsnode daemonset pods| `""`|
| `zfsNode.nodeSelector`| Nodeselector for zfsnode daemonset pods| `""`|
| `zfsNode.tolerations` | zfsnode daemonset's pod toleration values | `""`|
| `zfsNode.securityContext` | Seurity context for zfsnode daemonset container | `""`|
| `zfsController.resizer.image.registry`| Registry for csi-resizer image| `quay.io/`|
| `zfsController.resizer.image.repository`| Image repository for csi-resizer| `k8scsi/csi-resizer`|
| `zfsController.resizer.image.pullPolicy`| Image pull policy for csi-resizer| `IfNotPresent`|
| `zfsController.resizer.image.tag`| Image tag for csi-resizer| `v0.4.0`|
| `zfsController.snapshotter.image.registry`| Registry for csi-snapshotter image| `quay.io/`|
| `zfsController.snapshotter.image.repository`| Image repository for csi-snapshotter| `k8scsi/csi-snapshotter`|
| `zfsController.snapshotter.image.pullPolicy`| Image pull policy for csi-snapshotter| `IfNotPresent`|
| `zfsController.snapshotter.image.tag`| Image tag for csi-snapshotter| `v2.0.1`|
| `zfsController.snapshotController.image.registry`| Registry for snapshot-controller image| `quay.io/`|
| `zfsController.snapshotController.image.repository`| Image repository for snapshot-controller| `k8scsi/snapshot-controller`|
| `zfsController.snapshotController.image.pullPolicy`| Image pull policy for snapshot-controller| `IfNotPresent`|
| `zfsController.snapshotController.image.tag`| Image tag for snapshot-controller| `v2.0.1`|
| `zfsController.provisioner.image.registry`| Registry for csi-provisioner image| `quay.io/`|
| `zfsController.provisioner.image.repository`| Image repository for csi-provisioner| `k8scsi/csi-provisioner`|
| `zfsController.provisioner.image.pullPolicy`| Image pull policy for csi-provisioner| `IfNotPresent`|
| `zfsController.provisioner.image.tag`| Image tag for csi-provisioner| `v1.6.0`|
| `zfsController.updateStrategy.type`| Update strategy for zfs localpv controller statefulset | `RollingUpdate` |
| `zfsController.annotations` | Annotations for zfs localpv controller statefulset metadata| `""`|
| `zfsController.podAnnotations`| Annotations for zfs localpv controller statefulset's pods metadata | `""`|
| `zfsController.resources`| Resource and request and limit for zfs localpv controller statefulset containers | `""`|
| `zfsController.labels`| Labels for zfs localpv controller statefulset metadata | `""`|
| `zfsController.podLabels`| Appends labels to the zfs localpv controller statefulset pods| `""`|
| `zfsController.nodeSelector`| Nodeselector for zfs localpv controller statefulset pods| `""`|
| `zfsController.tolerations` | zfs localpv controller statefulset's pod toleration values | `""`|
| `zfsController.securityContext` | Seurity context for zfs localpv controller statefulset container | `""`|
| `serviceAccount.zfsNode.create` | Create a service account for zfsnode or not| `true`|
| `serviceAccount.zfsNode.name` | Name for the zfsnode service account| `openebs-zfs-node-sa`|
| `serviceAccount.zfsController.create` | Create a service account for zfs localpv controller or not| `true`|
| `serviceAccount.zfsController.name` | Name for the zfs localpv controller service account| `openebs-zfs-controller-sa`|
| `analytics.enabled` | Enable or Disable google analytics for the controller| `true`|

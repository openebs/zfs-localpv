
#  OpenEBS Local PV ZFS Provisioner

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Chart Lint and Test](https://github.com/openebs/zfs-localpv/workflows/Chart%20Lint%20and%20Test/badge.svg)
![Release Charts](https://github.com/openebs/zfs-localpv/workflows/Release%20Charts/badge.svg?branch=develop)

A Helm chart for openebs localpv zfs provisioner. This chart bootstraps OpenEBS ZFS LocalPV provisioner deployment on a [Kubernetes](http://kubernetes.io) cluster using the  [Helm](https://helm.sh) package manager.


**Homepage:** <http://www.openebs.io/>

## Get Repo Info

```console
helm repo add openebs-zfslocalpv https://openebs.github.io/zfs-localpv
helm repo update
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Chart

Please visit the [link](https://openebs.github.io/zfs-localpv/) for install instructions via helm3.

```console
# Helm
$ helm install [RELEASE_NAME] openebs-zfslocalpv/zfs-localpv
```

**Note:** If moving from the operator to helm
- Make sure the namespace provided in the helm install command is same as `OPENEBS_NAMESPACE` (by default it is `openebs`) env in the controller deployment.
- Before installing, clean up the stale deployment and daemonset from `openebs` namespace using the below commands
```sh
kubectl delete deploy openebs-zfs-controller -n openebs
kubectl delete ds openebs-zfs-node -n openebs
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

The following table lists the configurable parameters of the OpenEBS ZFS Localpv chart and their default values.

## Values

| Key                                                                 | Type   | Default Value               | Description                                                                                           |
|----------------------------------------------------------------------|--------|-----------------------------|-------------------------------------------------------------------------------------------------------|
| `global.imageRegistry`                                                | string | `""`                        | Global override for container image registry                                                           |
| `global.imagePullPolicy`                                              | string | `""`                        | Global override for  image pull policy.                                                            |
| `global.imagePullSecrets`                                             | list   | `[]`                        | Global image pull secrets (merged with local imagePullSecrets and not overridden).                                                                       |
| `global.analytics.enabled`                                            | bool   | `null`                      | Global override for analytics.                                                                         |
| `global.kubeletDir`                                                   | string |  `""`                       | Global override for kubelet directory                                                                  |
| `analytics.enabled`                                                  | bool   | `true`                      | Enables or disables analytics reporting for the chart.                                                 |
| `analytics.installerType`                                             | string | `"zfs-localpv-helm"`        | Specifies the installer type for analytics.                                                           |
| `backupGC.enabled`                                                    | bool   | `false`                     | Enables or disables garbage collection for backups.                                                   |
| `crds.csi.volumeSnapshots.enabled`                                    | bool   | `true`                      | Enables or disables the installation of Volume Snapshot CRDs.                                          |
| `crds.zfsLocalPv.enabled`                                             | bool   | `true`                      | Enables or disables the installation of ZFS Local Persistent Volume CRDs.                             |
| `enableHelmMetaLabels`                                                | bool   | `true`                      | Adds Helm-specific metadata labels to the components.                                                 |
| `feature.storageCapacity`                                             | bool   | `true`                      | Enables or disables storage capacity tracking feature.                                                |
| `imagePullSecrets`                                                    | list   | `[]`                        | List of secrets to use when pulling images from private registries.                                    |
| `loggingLabels."openebs.io/logging"`                                  | string | `"true"`                    | Enables or disables logging for OpenEBS components.                                                   |
| `logLevel`                                                           | int/null | `null`                    | Common klog verbosity (`--v`) for ZFS controller/node and all CSI sidecars. Unset preserves existing defaults; individual levels, including `0`, take precedence. |
| `rbac.pspEnabled`                                                     | bool   | `false`                     | Enables or disables the creation of PodSecurityPolicy resources.                                       |
| `role`                                                                | string | `"openebs-zfs"`             | Specifies the role for the OpenEBS ZFS component.                                                     |
| `serviceAccount.zfsController.create`                                 | bool   | `true`                      | Specifies whether a service account should be created for the ZFS controller.                          |
| `serviceAccount.zfsController.name`                                   | string | `"openebs-zfs-controller-sa"` | The name of the service account to use for the ZFS controller.                                         |
| `serviceAccount.zfsNode.create`                                       | bool   | `true`                      | Specifies whether a service account should be created for the ZFS node.                                |
| `serviceAccount.zfsNode.name`                                         | string | `"openebs-zfs-node-sa"`     | The name of the service account to use for the ZFS node.                                               |
| `zfs.bin`                                                             | string | `"zfs"`                     | Path to the ZFS binary.                                                                                |
| `zfsController.additionalVolumes`                                     | map    | `{}`                        | Extra volumes for the ZFS controller pod, keyed by volume name; each value is the volume source (e.g. `hostPath`) and is rendered through `tpl`. |
| `zfsController.annotations`                                           | map    | `{}`                        | Annotations to add to the ZFS controller Deployment.                                                    |
| `zfsController.componentName`                                         | string | `"openebs-zfs-controller"`  | Name of the ZFS controller component.                                                                  |
| `zfsController.initContainers`                                        | list   | `[]`                        | List of init containers to run before the ZFS controller pods.                                         |
| `zfsController.logLevel`                                              | int/null | `null`                    | ZFS controller plugin verbosity; inherits `logLevel`, otherwise the binary default (`0`). |
| `zfsController.nodeSelector`                                          | map    | `{}`                        | Node selector for scheduling ZFS controller pods.                                                      |
| `zfsController.podAnnotations`                                        | map    | `{}`                        | Annotations to add to the ZFS controller pods.                                                         |
| `zfsController.podLabels.name`                                        | string | `"openebs-zfs-controller"`  | Labels to add to the ZFS controller pods.                                                              |
| `zfsController.priorityClass.create`                                  | bool   | `true`                      | Specifies whether to create a priority class for the ZFS controller pods.                              |
| `zfsController.priorityClass.name`                                    | string | `"zfs-csi-controller-critical"` | The name of the priority class to use for the ZFS controller pods.                                      |
| `zfsController.provisioner.extraArgs`                                 | list   | `[]`                        | Additional arguments to pass to the CSI provisioner.                                                   |
| `zfsController.provisioner.image.pullPolicy`                          | string | `"IfNotPresent"`            | Image pull policy for the CSI provisioner.                                                             |
| `zfsController.provisioner.image.registry`                            | string | `"registry.k8s.io/"`        | Image registry for the CSI provisioner.                                                                |
| `zfsController.provisioner.image.repository`                          | string | `"sig-storage/csi-provisioner"` | Image repository for the CSI provisioner.                                                              |
| `zfsController.provisioner.image.tag`                                 | string | `"v5.2.0"`                  | Image tag for the CSI provisioner.                                                                      |
| `zfsController.provisioner.logLevel`                                  | int/null | `null`                    | CSI provisioner verbosity; inherits `logLevel`, otherwise `5`. |
| `zfsController.provisioner.name`                                      | string | `"csi-provisioner"`         | Name of the CSI provisioner container.                                                                 |
| `zfsController.replicas`                                              | int    | `1`                         | Number of replicas for the ZFS controller deployment.                                                  |
| `zfsController.resizer.extraArgs`                                     | list   | `[]`                        | Additional arguments to pass to the CSI resizer.                                                       |
| `zfsController.resizer.image.pullPolicy`                              | string | `"IfNotPresent"`            | Image pull policy for the CSI resizer.                                                                 |
| `zfsController.resizer.image.registry`                                | string | `"registry.k8s.io/"`        | Image registry for the CSI resizer.                                                                    |
| `zfsController.resizer.image.repository`                              | string | `"sig-storage/csi-resizer"` | Image repository for the CSI resizer.                                                                  |
| `zfsController.resizer.image.tag`                                     | string | `"v1.13.2"`                 | Image tag for the CSI resizer.                                                                          |
| `zfsController.resizer.logLevel`                                      | int/null | `null`                    | CSI resizer verbosity; inherits `logLevel`, otherwise `5`. |
| `zfsController.resizer.name`                                          | string | `"csi-resizer"`             | Name of the CSI resizer container.                                                                     |
| `zfsController.resources`                                             | map    | `{}`                        | Resource requests and limits for the ZFS controller pods.                                               |
| `zfsController.securityContext`                                       | map    | `{}`                        | Security context for the ZFS controller pods.                                                          |
| `zfsController.snapshotController.enabled`                            | bool   | `true`                      | Enables or disables the snapshot controller container. Set to false if the cluster already runs one.     |
| `zfsController.snapshotController.extraArgs`                          | list   | `[]`                        | Additional arguments to pass to the snapshot controller.                                                |
| `zfsController.snapshotController.image.pullPolicy`                   | string | `"IfNotPresent"`            | Image pull policy for the snapshot controller.                                                          |
| `zfsController.snapshotController.image.registry`                     | string | `"registry.k8s.io/"`        | Image registry for the snapshot controller.                                                             |
| `zfsController.snapshotController.image.repository`                   | string | `"sig-storage/snapshot-controller"` | Image repository for the snapshot controller.                                                           |
| `zfsController.snapshotController.image.tag`                          | string | `"v8.2.0"`                  | Image tag for the snapshot controller. 
| `zfsController.snapshotController.name`                               | string | `"snapshot-controller"`     | Name of the snapshot controller container. |
| `zfsController.snapshotController.logLevel`                           | int/null | `null`                    | Snapshot controller verbosity; inherits `logLevel`, otherwise `5`. |
| `zfsController.snapshotter.enabled`                                   | bool   | `true`                      | Enables or disables the csi-snapshotter container. Set to false if the VolumeSnapshot CRDs are not installed. |
| `zfsController.snapshotter.logLevel`                                  | int/null | `null`                    | CSI snapshotter verbosity; inherits `logLevel`, otherwise the binary default (`0`). |
| `zfsController.snapshotter.extraArgs`                                 | list   | `[]`                        | Additional arguments to pass to the CSI snapshotter. |
| `zfsController.snapshotter.image.pullPolicy`                          | string | `"IfNotPresent"`            | Image pull policy for the CSI snapshotter. |
| `zfsController.snapshotter.image.registry`                            | string | `"registry.k8s.io/"`        | Image registry for the CSI snapshotter. |
| `zfsController.snapshotter.image.repository`                          | string | `"sig-storage/csi-snapshotter"`| Image repository for the CSI snapshotter. |
| `zfsController.snapshotter.image.tag`                                 | string | `"v8.2.0"`                  | Image tag for the CSI snapshotter. |
| `zfsController.snapshotter.name`                                      | string | `"csi-snapshotter"`         | Name of the CSI snapshotter container. |
| `zfsController.tolerations`                                           | list   | `[]`                        | Tolerations for scheduling ZFS controller pods. |
| `zfsController.updateStrategy`                                        | map    | `{"type":"RollingUpdate"}`  | Deployment strategy of the ZFS controller. |
| `zfsNode.additionalVolumes`                                           | map    | `{}`                        | Extra volumes for the ZFS node pods, keyed by volume name; each value is the volume source and is rendered through `tpl`. |
| `zfsNode.allowedTopologyKeys`                                         | string | `"All"`                     | Comma separated node label keys the driver reports as topology (e.g. `kubernetes.io/hostname,openebs.io/rack`); `All` reports every node label. |
| `zfsNode.annotations`                                                 | map    | `{}`                        | Annotations to add to the ZFS node DaemonSet. |
| `zfsNode.componentName`                                               | string | `"openebs-zfs-node"`        | Name of the ZFS node component, used in its labels. |
| `zfsNode.defaultFormatOptions`                                        | map    | `{}`                        | Extra mkfs options per filesystem, used when the StorageClass of a volume does not set `formatOptions`. A StorageClass value replaces the default of its filesystem, the two are not merged. Keys are the formatted filesystem types (ext2, ext3, ext4, xfs, btrfs), values the space separated mkfs options of that filesystem as one string, e.g. `{"ext4": "-m 0", "xfs": "-i nrext64=0"}`. |
| `zfsNode.driverRegistrar.image.pullPolicy`                            | string | `"IfNotPresent"`            | Image pull policy for the CSI node driver registrar. |
| `zfsNode.driverRegistrar.image.registry`                              | string | `"registry.k8s.io/"`        | Image registry for the CSI node driver registrar. |
| `zfsNode.driverRegistrar.image.repository`                            | string | `"sig-storage/csi-node-driver-registrar"`| Image repository for the CSI node driver registrar. |
| `zfsNode.driverRegistrar.image.tag`                                   | string | `"v2.13.0"`                 | Image tag for the CSI node driver registrar. |
| `zfsNode.driverRegistrar.logLevel`                                    | int/null | `null`                    | CSI node driver registrar verbosity; inherits `logLevel`, otherwise `5`. |
| `zfsNode.driverRegistrar.name`                                        | string | `"csi-node-driver-registrar"`| Name of the CSI node driver registrar container. |
| `zfsNode.encrKeysDir`                                                 | string | `"/home/keys"`              | Host directory holding the keys of encrypted volumes, mounted at `/home/keys` in the plugin. |
| `zfsNode.initContainers`                                              | list   | `[]`                        | List of init containers to run before the ZFS node pods. |
| `zfsNode.kubeletDir`                                                  | string | `"/var/lib/kubelet/"`       | Kubelet root directory on the nodes; `global.kubeletDir` takes precedence when set. |
| `zfsNode.logLevel`                                                    | int/null | `null`                    | ZFS node plugin verbosity; inherits `logLevel`, otherwise the binary default (`0`). |
| `zfsNode.nodeSelector`                                                | map    | `{}`                        | Node selector for scheduling ZFS node pods. |
| `zfsNode.podAnnotations`                                              | map    | `{}`                        | Annotations to add to the ZFS node pods. |
| `zfsNode.podLabels`                                                   | map    | `{}`                        | Labels to add to the ZFS node pods. |
| `zfsNode.priorityClass.create`                                        | bool   | `true`                      | Creates a priority class for the ZFS node pods and sets it on them. |
| `zfsNode.priorityClass.name`                                          | string | `"zfs-csi-node-critical"`   | Name of the priority class, prefixed with the release name when created. |
| `zfsNode.resources`                                                   | map    | `{}`                        | Resource requests and limits for the ZFS node containers. |
| `zfsNode.securityContext`                                             | map    | `{}`                        | Security context for the ZFS node pods. |
| `zfsNode.tolerations`                                                 | list   | `[]`                        | Tolerations for scheduling ZFS node pods. |
| `zfsNode.updateStrategy`                                              | map    | `{"type":"RollingUpdate","rollingUpdate":{"maxUnavailable":"100%"}}` | Update strategy of the ZFS node DaemonSet. |
| `zfsPlugin.image.pullPolicy`                                          | string | `"IfNotPresent"`            | Image pull policy for the ZFS driver. |
| `zfsPlugin.image.registry`                                            | string | `"docker.io/"`              | Image registry for the ZFS driver. |
| `zfsPlugin.image.repository`                                          | string | `"openebs/zfs-driver"`      | Image repository for the ZFS driver. |
| `zfsPlugin.image.tag`                                                 | string | `"2.12.0-develop"`          | Image tag for the ZFS driver. |
| `zfsPlugin.name`                                                      | string | `"openebs-zfs-plugin"`      | Name of the ZFS driver container in the controller and node pods. |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install <release-name> -f values.yaml openebs/zfs-localpv
```

> **Tip**: You can use the default [values.yaml](values.yaml)

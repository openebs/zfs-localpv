
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
| `analytics.enabled`                                                  | bool   | `true`                      | Enables or disables analytics reporting for the chart.                                                 |
| `analytics.installerType`                                             | string | `"zfs-localpv-helm"`        | Specifies the installer type for analytics.                                                           |
| `backupGC.enabled`                                                    | bool   | `false`                     | Enables or disables garbage collection for backups.                                                   |
| `crds.csi.volumeSnapshots.enabled`                                    | bool   | `true`                      | Enables or disables the installation of Volume Snapshot CRDs.                                          |
| `crds.zfsLocalPv.enabled`                                             | bool   | `true`                      | Enables or disables the installation of ZFS Local Persistent Volume CRDs.                             |
| `enableHelmMetaLabels`                                                | bool   | `true`                      | Adds Helm-specific metadata labels to the components.                                                 |
| `feature.storageCapacity`                                             | bool   | `true`                      | Enables or disables storage capacity tracking feature.                                                |
| `imagePullSecrets`                                                    | list   | `[]`                        | List of secrets to use when pulling images from private registries.                                    |
| `loggingLabels."openebs.io/logging"`                                  | string | `"true"`                    | Enables or disables logging for OpenEBS components.                                                   |
| `rbac.pspEnabled`                                                     | bool   | `false`                     | Enables or disables the creation of PodSecurityPolicy resources.                                       |
| `role`                                                                | string | `"openebs-zfs"`             | Specifies the role for the OpenEBS ZFS component.                                                     |
| `serviceAccount.zfsController.create`                                 | bool   | `true`                      | Specifies whether a service account should be created for the ZFS controller.                          |
| `serviceAccount.zfsController.name`                                   | string | `"openebs-zfs-controller-sa"` | The name of the service account to use for the ZFS controller.                                         |
| `serviceAccount.zfsNode.create`                                       | bool   | `true`                      | Specifies whether a service account should be created for the ZFS node.                                |
| `serviceAccount.zfsNode.name`                                         | string | `"openebs-zfs-node-sa"`     | The name of the service account to use for the ZFS node.                                               |
| `zfs.bin`                                                             | string | `"zfs"`                     | Path to the ZFS binary.                                                                                |
| `zfsController.additionalVolumes`                                     | list   | `[]`                        | Additional volumes to mount into the ZFS controller pods.                                              |
| `zfsController.annotations`                                           | map    | `{}`                        | Annotations to add to the ZFS controller pods.                                                         |
| `zfsController.componentName`                                         | string | `"openebs-zfs-controller"`  | Name of the ZFS controller component.                                                                  |
| `zfsController.initContainers`                                        | list   | `[]`                        | List of init containers to run before the ZFS controller pods.                                         |
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
| `zfsController.provisioner.name`                                      | string | `"csi-provisioner"`         | Name of the CSI provisioner container.                                                                 |
| `zfsController.replicas`                                              | int    | `1`                         | Number of replicas for the ZFS controller deployment.                                                  |
| `zfsController.resizer.extraArgs`                                     | list   | `[]`                        | Additional arguments to pass to the CSI resizer.                                                       |
| `zfsController.resizer.image.pullPolicy`                              | string | `"IfNotPresent"`            | Image pull policy for the CSI resizer.                                                                 |
| `zfsController.resizer.image.registry`                                | string | `"registry.k8s.io/"`        | Image registry for the CSI resizer.                                                                    |
| `zfsController.resizer.image.repository`                              | string | `"sig-storage/csi-resizer"` | Image repository for the CSI resizer.                                                                  |
| `zfsController.resizer.image.tag`                                     | string | `"v1.13.2"`                 | Image tag for the CSI resizer.                                                                          |
| `zfsController.resizer.name`                                          | string | `"csi-resizer"`             | Name of the CSI resizer container.                                                                     |
| `zfsController.resources`                                             | map    | `{}`                        | Resource requests and limits for the ZFS controller pods.                                               |
| `zfsController.securityContext`                                       | map    | `{}`                        | Security context for the ZFS controller pods.                                                          |
| `zfsController.snapshotController.extraArgs`                          | list   | `[]`                        | Additional arguments to pass to the snapshot controller.                                                |
| `zfsController.snapshotController.image.pullPolicy`                   | string | `"IfNotPresent"`            | Image pull policy for the snapshot controller.                                                          |
| `zfsController.snapshotController.image.registry`                     | string | `"registry.k8s.io/"`        | Image registry for the snapshot controller.                                                             |
| `zfsController.snapshotController.image.repository`                   | string | `"sig-storage/snapshot-controller"` | Image repository for the snapshot controller.                                                           |
| `zfsController.snapshotController.image.tag`                          | string | `"v8.2.0"`                  | Image tag for the snapshot controller. 

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install <release-name> -f values.yaml openebs/zfs-localpv
```

> **Tip**: You can use the default [values.yaml](values.yaml)

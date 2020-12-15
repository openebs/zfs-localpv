## Introduction

This chart bootstraps OpenEBS ZFS Localpv deployment on a [Kubernetes](http://kubernetes.io) cluster using the 
[Helm](https://helm.sh) package manager.

## Installation

You can run OpenEBS ZFS Localpv on any Kubernetes 1.17+ cluster in a matter of seconds.

Please visit the [link](https://openebs.github.io/zfs-localpv/) for install instructions via helm3.

## Configuration

The following table lists the configurable parameters of the OpenEBS ZFS Localpv chart and their default values.

| Parameter                               | Description                                   | Default                                   |
| ----------------------------------------| --------------------------------------------- | ----------------------------------------- |
| `imagePullSecrets`                      | Provides image pull secrect                   | `""`                                      |
| `zfsPlugin.image.registry`                    | Registry for openebs-zfs-plugin image          | `""`                                      |
| `zfsPlugin.image.repository`                  | Image repository for openebs-zfs-plugin        | `openebs/zfs-driver`                        |
| `zfsPlugin.image.pullPolicy`                  | Image pull policy for openebs-zfs-plugin       | `IfNotPresent`                            |
| `zfsPlugin.image.tag`                         | Image tag for openebs-zfs-plugin               | `1.1.0`                                        |
| `zfsNode.driverRegistrar.image.registry`                    | Registry for csi-node-driver-registrar image          | `quay.io/`                                      |
| `zfsNode.driverRegistrar.image.repository`                  | Image repository for csi-node-driver-registrar        | `k8scsi/csi-node-driver-registrar`                        |
| `zfsNode.driverRegistrar.image.pullPolicy`                  | Image pull policy for csi-node-driver-registrar       | `IfNotPresent`                            |
| `zfsNode.driverRegistrar.image.tag`                         | Image tag for csi-node-driver-registrar               | `v1.2.0`                                        |
| `zfsNode.updateStrategy.type`               | Update strategy for zfsnode daemonset             | `RollingUpdate`                           |
| `zfsNode.kubeletDir`               | Kubelet mount point for zfsnode daemonset            | `"/var/lib/kubelet/"`                           |
| `zfsNode.annotations`                       | Annotations for zfsnode daemonset metadata        | `""`                                      |
| `zfsNode.podAnnotations`                    | Annotations for zfsnode daemonset's pods metadata | `""`                                      |
| `zfsNode.resources`                         | Resource and request and limit for zfsnode daemonset containers | `""`                                      |
| `zfsNode.labels`                    | Labels for zfsnode daemonset metadata | `""`                                      |
| `zfsNode.podLabels`                         | Appends labels to the zfsnode daemonset pods                    | `""`                                      |
| `zfsNode.nodeSelector`                      | Nodeselector for zfsnode daemonset pods               | `""`                                      |
| `zfsNode.tolerations`                       | zfsnode daemonset's pod toleration values         | `""`                                      |
| `zfsNode.securityContext`                   | Seurity context for zfsnode daemonset container                 | `""`                                      |
| `controller.resizer.image.registry`                    | Registry for csi-resizer image          | `quay.io/`                                      |
| `controller.resizer.image.repository`                  | Image repository for csi-resizer        | `k8scsi/csi-resizer`                        |
| `controller.resizer.image.pullPolicy`                  | Image pull policy for csi-resizer       | `IfNotPresent`                            |
| `controller.resizer.image.tag`                         | Image tag for csi-resizer               | `v0.4.0`                                        |
| `controller.snapshotter.image.registry`                    | Registry for csi-snapshotter image          | `quay.io/`                                      |
| `controller.snapshotter.image.repository`                  | Image repository for csi-snapshotter        | `k8scsi/csi-snapshotter`                        |
| `controller.snapshotter.image.pullPolicy`                  | Image pull policy for csi-snapshotter       | `IfNotPresent`                            |
| `controller.snapshotter.image.tag`                         | Image tag for csi-snapshotter               | `v2.0.1`                                        |
| `controller.snapshotController.image.registry`                    | Registry for snapshot-controller image          | `quay.io/`                                      |
| `controller.snapshotController.image.repository`                  | Image repository for snapshot-controller        | `k8scsi/snapshot-controller`                        |
| `controller.snapshotController.image.pullPolicy`                  | Image pull policy for snapshot-controller       | `IfNotPresent`                            |
| `controller.snapshotController.image.tag`                         | Image tag for snapshot-controller               | `v2.0.1`                                        |
| `controller.provisioner.image.registry`                    | Registry for csi-provisioner image          | `quay.io/`                                      |
| `controller.provisioner.image.repository`                  | Image repository for csi-provisioner        | `k8scsi/csi-provisioner`                        |
| `controller.provisioner.image.pullPolicy`                  | Image pull policy for csi-provisioner       | `IfNotPresent`                            |
| `controller.provisioner.image.tag`                         | Image tag for csi-provisioner               | `v1.6.0`                                        |
| `controller.updateStrategy.type`               | Update strategy for zfs localpv controller statefulset             | `RollingUpdate`                           |
| `controller.annotations`                       | Annotations for zfs localpv controller statefulset metadata        | `""`                                      |
| `controller.podAnnotations`                    | Annotations for zfs localpv controller statefulset's pods metadata | `""`                                      |
| `controller.resources`                         | Resource and request and limit for zfs localpv controller statefulset containers | `""`                                      |
| `controller.labels`                    | Labels for zfs localpv controller statefulset metadata | `""`                                      |
| `controller.podLabels`                         | Appends labels to the zfs localpv controller statefulset pods                    | `""`                                      |
| `controller.nodeSelector`                      | Nodeselector for zfs localpv controller statefulset pods               | `""`                                      |
| `controller.tolerations`                       | zfs localpv controller statefulset's pod toleration values         | `""`                                      |
| `controller.securityContext`                   | Seurity context for zfs localpv controller statefulset container                 | `""`                                      |
| `serviceAccount.zfsNode.create`                 | Create a service account for zfsnode or not               | `true`                                    |
| `serviceAccount.zfsNode.name`                   | Name for the zfsnode service account                  | `openebs-zfs-node-sa`                                    |
| `serviceAccount.controller.create`                 | Create a service account for zfs localpv controller or not               | `true`                                    |
| `serviceAccount.controller.name`                   | Name for the zfs localpv controller service account                  | `openebs-zfs-controller-sa`                                    |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install <release-name> -f values.yaml openebs/zfs-localpv
```

> **Tip**: You can use the default [values.yaml](values.yaml)

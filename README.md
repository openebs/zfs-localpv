## OpenEBS Local PV ZFS

[![CNCF Status](https://img.shields.io/badge/cncf%20status-sandbox-blue.svg)](https://www.cncf.io/projects/openebs/)
[![LICENSE](https://img.shields.io/github/license/openebs/openebs.svg)](./LICENSE)
[![Build Status](https://github.com/openebs/zfs-localpv/actions/workflows/build.yml/badge.svg)](https://github.com/openebs/zfs-localpv/actions/workflows/build.yml)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=shield&issueType=license)](https://app.fossa.com/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_shield&issueType=license)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/3523/badge)](https://www.bestpractices.dev/projects/3523)
[![Slack](https://img.shields.io/badge/chat-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs/)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://github.com/openebs/community/blob/HEAD/README.md#community)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/zfs-localpv)](https://goreportcard.com/report/github.com/openebs/zfs-localpv)
[![CLOMonitor](https://img.shields.io/endpoint?url=https://clomonitor.io/api/projects/cncf/openebs/badge)](https://clomonitor.io/projects/cncf/openebs)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/openebs)](https://artifacthub.io/packages/helm/openebs/openebs)

## Overview

OpenEBS Local PV ZFS is a [CSI](https://github.com/container-storage-interface/spec) plugin for implementation of [ZFS](https://en.wikipedia.org/wiki/ZFS) backed persistent volumes for Kubernetes. It is a local storage solution, which means the device, volume and the application are on the same host. It doesn't contain any dataplane, i.e only its simply a control-plane for the kernel zfs volumes. It mainly comprises of two components which are implemented in accordance to the CSI Specs:

1. CSI Controller - Frontends the incoming requests and initiates the operation.
2. CSI Node Plugin - Serves the requests by performing the operations and making the volume available for the initiator.

## Why OpenEBS Local PV ZFS?

1. Lightweight, easy to set up storage provisoner for host-local volumes in k8s ecosystem.
2. Makes ZFS stack available to K8s, allowing end users to use the ZFS functionalites like snapshot, restore, clone, thin provisioning, resize, encryption, compression, dedup, etc for their Persistent Volumes.
3. Cloud native, i.e based on CSI spec, so developed to run on K8s.

## Architecture

LocalPV refers to storage that is directly attached to a specific node in the Kubernetes cluster. It uses locally available disks (e.g., SSDs, HDDs) on the node.

<b>Use Case</b>: Ideal for workloads that require low-latency access to storage or when data locality is critical (e.g., databases, caching systems).

#### Characteristics:

- <b>Node-bound</b>: The volume is tied to the node where the disk is physically located.
- <b>No replication</b>: Data is not replicated across nodes, so if the node fails, the data may become inaccessible.
- <b>High performance</b>: Since the storage is local, it typically offers lower latency compared to network-attached storage.

The diagram below depicts the mapping to the host disks, the ZFS stack on top of the disks and the kubernetes persistent volumes to be consumed by the workload. Local PV ZFS CSI Controller upon creation of the Persistent Volume Claim, creates a ZFSVolume CR, which emits an event for Local PV ZFS CSI Node Plugin to create the zvol/dataset. When workloads are scheduled the Local PV ZFS CSI Node Plugin makes this zvol/dataset available via a mount point on the host.

```mermaid
graph TD;
  subgraph Node2["Node 2"]
    subgraph K8S_NODE1[" "]
      N1_PV1["PV"] --> N1_APP1["APP"]
      N1_PV2["PV"] --> N1_APP2["APP"]
    end
    subgraph LVM_Stack2["ZFS Stack"]
      V1_1 --> L1_1["ZVOL"]
      V1_1 --> L3_1["ZVOL"]
      L1_1 --> N1_PV1 
      L3_1 --> N1_PV2
    end
    subgraph Blockdevices1[" "]
      D1["/dev/sdc"] --> V1_1["ZPOOL"]
      D2["/dev/sdb"] --> V1_1["ZPOOL"]
    end
  end

  subgraph Node1["Node 1"]
    subgraph K8S_NODE2[" "]
      N2_PV1["PV"] --> N2_APP1["APP"]
    end
    subgraph LVM_Stack1["ZFS Stack"]
      V2_2 --> Z2_2["ZVOL"]
      Z2_2 --> N2_PV1 
    end
    subgraph Blockdevices2[" "]
      D3["/dev/sdb"] --> V2_2["ZPOOL"]
    end
  end

  classDef pv fill:#FFCC00,stroke:#FF9900,color:#000;
  classDef app fill:#99CC00,stroke:#66CC00,color:#000;
  classDef disk fill:#FF6666,stroke:#FF3333,color:#000;
  classDef zfs fill:#99CCFF,stroke:#6699FF,color:#000;

  class N1_PV1,N1_PV2,N2_PV1 pv;
  class N1_APP1,N1_APP2,N2_APP1 app;
  class D1,D2,D3 disk;
  class V1_1,V2_2 zfs;
  class L1_1,L3_1,Z2_2 zfs;

```

## Supported System

> | Name | Version |
> | :--- | :--- |
> | K8S | 1.23+ |
> | Distro | Alpine, Arch, CentOS, Debian, Fedora, NixOS, SUSE, RHEL, Ubuntu |
> | Kenel | oldest supported kernel is 2.6.32 |
> | ZFS | 0.7, 0.8, 2.2.3 |
> | Memory | ECC Memory is highly recommended |
> | RAM | 8GiB for best perf with Dedup enabled. (Will work with 2GiB or less without Dedup) |

Check the [features](./docs/features.md) supported for each k8s version.

## Documents

- [Prerequisites](./docs/quickstart.md#prerequisites)
- [Quickstart](./docs/quickstart.md#setup)
- [Developer Setup](./docs/developer-setup.md#development-workflow)
- [Testing](./docs/developer-setup.md#testing)
- [Contibuting Guidelines](./CONTRIBUTING.md)
- [Governance](./GOVERNANCE.md)
- [Changelog](./CHANGELOG.md)
- [Release Process](./RELEASE.md)

## Features

- [x] Access Modes
    - [x] ReadWriteOnce
    - ~~ReadOnlyMany~~
    - ~~ReadWriteMany~~
- [x] Volume modes
    - [x] `Filesystem` mode
    - [x] `Block` mode
- [x] Supports fsTypes: `ext4`, `btrfs`, `xfs`, `zfs`
- [x] Volume metrics
- [x] Snapshot
    - [x] [Create](docs/snapshot.md)
    - [x] [Restore](docs/clone.md#create-clone-from-snapshot)
- [x] [Clone](docs/clone.md#create-clone-from-volume)
- [x] [Volume Resize](docs/resize.md)
- [x] [Raw Block Volume](docs/raw-block-volume.md)
- [x] [Backup/Restore](docs/backup-restore.md)
- [ ] Ephemeral inline volume

## Dev Activity dashboard

![Alt](https://repobeats.axiom.co/api/embed/201b3c4bdd37f67c41fae1ea295d8a1887fba654.svg "Repobeats analytics image")

## License compliance

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=large&issueType=license)](https://app.fossa.com/projects/custom%2B162%2Fgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_large&issueType=license)

## OpenEBS is a [CNCF Sandbox Project](https://www.cncf.io/projects/openebs)

![OpenEBS is a CNCF Sandbox Project](https://github.com/cncf/artwork/blob/main/other/cncf/horizontal/color/cncf-color.png)

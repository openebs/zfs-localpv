v2.6.0 / 2024-07-03
========================
* feat(analytics): add heartbeat pinger ([#548](https://github.com/openebs/zfs-localpv/pull/548),[@niladrih](https://github.com/niladrih))
* fix: wrap k8s api error in GetNodeID ([#535](https://github.com/openebs/zfs-localpv/pull/535),[@aep](https://github.com/aep))

v2.5.0 / 2024-03-22
========================
* feat(deploy/helm): move volumesnapshot CRDs to the template dir ([#488](https://github.com/openebs/zfs-localpv/pull/488),[@hrudaya21](https://github.com/hrudaya21))
* fix(plugin): Fix ability to have custom value for openebs.io/nodeid ([#451](https://github.com/openebs/zfs-localpv/pull/451),[@jnels124](https://github.com/jnels124))
* fix(helm): Add extra args to zfsController containers and leader election inteligence ([#492](https://github.com/openebs/zfs-localpv/pull/492),[@trunet](https://github.com/trunet))
* chore(design): adding pv migration proposal ([#336](https://github.com/openebs/zfs-localpv/pull/336),[@pawanpraka1](https://github.com/pawanpraka1))
* fix(charts): correct default chart values ([#506](https://github.com/openebs/zfs-localpv/pull/506),[@jnels124](https://github.com/jnels124))
* chore: update protobuf deps ([#514](https://github.com/openebs/zfs-localpv/pull/514),[@niladrih](https://github.com/niladrih))
* chore: change zfs-controller to a deployment from statefulset ([#513](https://github.com/openebs/zfs-localpv/pull/513),[@Abhinandan-Purkait](https://github.com/Abhinandan-Purkait))

v2.4.0 / 2023-12-12
========================
* fix(localpv): restore size to return as part of snapshot create response ([#480](https://github.com/openebs/zfs-localpv/pull/480),[@hrudaya21](https://github.com/hrudaya21))
* feat(usedcapcity): kubectl describe zfsnode should show the used capacity information ([#485](https://github.com/openebs/zfs-localpv/pull/485),[@hrudaya21](https://github.com/hrudaya21))
* feat(event): update ua to ga4 analytics ([#490](https://github.com/openebs/zfs-localpv/pull/490),[@Abhinandan-Purkait](https://github.com/Abhinandan-Purkait))

v2.3.0 / 2023-07-23
========================
* feat(csi): bump up csi provisioner to v3.5.0 and other updates ([#457](https://github.com/openebs/zfs-localpv/pull/457),[@vharsh](https://github.com/vharsh))
* feat(helm): add support for providing additional volumes and adding init containers ([#455](https://github.com/openebs/zfs-localpv/pull/455),[@jnels124](https://github.com/jnels124))
* fix(helm): Possibility to override zfs encryption keys directory ([#487](https://github.com/openebs/zfs-localpv/pull/487),[@trunet](https://github.com/trunet))

v2.2.0 / 2023-05-29
========================
* perf(zfs): optimise pool listing for pools with many datasets ([#440](https://github.com/openebs/zfs-localpv/pull/440),[@lowjoel](https://github.com/lowjoel))
* feat(deps): Bump golang, k8s and lib-csi versions ([#444](https://github.com/openebs/zfs-localpv/pull/444),[@shubham14bajpai](https://github.com/shubham14bajpai))

v2.0.0 / 2022-01-11
========================
* fix(localpv): fixing CSIStorageCapacity when "poolname" param has child dataset ([#393](https://github.com/openebs/zfs-localpv/pull/393),[@netom](https://github.com/netom))

v1.7.0 / 2021-03-15
========================
* feat(migration): adding support to migrate the PV to a new node ([#304](https://github.com/openebs/zfs-localpv/pull/304),[@pawanpraka1](https://github.com/pawanpraka1))
* fix(topo): support old topology key for backward compatibility ([#320](https://github.com/openebs/zfs-localpv/pull/320),[@pawanpraka1](https://github.com/pawanpraka1))

v1.6.0 / 2021-04-14
========================
* refact(deps): bump k8s and client-go deps to version v0.20.2 ([#294](https://github.com/openebs/zfs-localpv/pull/294),[@prateekpandey14](https://github.com/prateekpandey14))
* remove finalizer that is owned by ZFS-LocalPV ([#303](https://github.com/openebs/zfs-localpv/pull/303),[@pawanpraka1](https://github.com/pawanpraka1))
* try volume creation on all the nodes that satisfy topology contraints ([#270](https://github.com/openebs/zfs-localpv/pull/270),[@pawanpraka1](https://github.com/pawanpraka1))
* With k8s v1.22 the v1beta1 for various resources will no longer be supported. Updating the storage and apiexention version to v1 for better support. ([#299](https://github.com/openebs/zfs-localpv/pull/299),[@shubham14bajpai](https://github.com/shubham14bajpai))

v1.6.0-RC2 / 2021-04-12
========================

v1.6.0-RC1 / 2021-04-06
========================
* refact(deps): bump k8s and client-go deps to version v0.20.2 ([#294](https://github.com/openebs/zfs-localpv/pull/294),[@prateekpandey14](https://github.com/prateekpandey14))
* remove finalizer that is owned by ZFS-LocalPV ([#303](https://github.com/openebs/zfs-localpv/pull/303),[@pawanpraka1](https://github.com/pawanpraka1))
* try volume creation on all the nodes that satisfy topology contraints ([#270](https://github.com/openebs/zfs-localpv/pull/270),[@pawanpraka1](https://github.com/pawanpraka1))
* With k8s v1.22 the v1beta1 for various resources will no longer be supported. Updating the storage and apiexention version to v1 for better support. ([#299](https://github.com/openebs/zfs-localpv/pull/299),[@shubham14bajpai](https://github.com/shubham14bajpai))


v1.5.0 / 2021-03-12
========================
* adding support to restore in an encrypted pool ([#292](https://github.com/openebs/zfs-localpv/pull/292),[@pawanpraka1](https://github.com/pawanpraka1))
* move the bdd test cases to github action ([#293](https://github.com/openebs/zfs-localpv/pull/293),[@shubham14bajpai](https://github.com/shubham14bajpai))

v1.5.0-RC2 / 2021-03-11
========================

v1.5.RC1 / 2021-03-09
========================
* adding support to restore in an encrypted pool ([#292](https://github.com/openebs/zfs-localpv/pull/292),[@pawanpraka1](https://github.com/pawanpraka1))
* move the bdd test cases to github action ([#293](https://github.com/openebs/zfs-localpv/pull/293),[@shubham14bajpai](https://github.com/shubham14bajpai))


v1.4.0 / 2021-02-13
========================

* update k8s sidecar images to gcr ([#284](https://github.com/openebs/zfs-localpv/pull/284),[@shubham14bajpai](https://github.com/shubham14bajpai))
* adding resize support for raw block volumes ([#281](https://github.com/openebs/zfs-localpv/pull/281),[@pawanpraka1](https://github.com/pawanpraka1))

v1.4.0-RC2 / 2021-02-11
========================


v1.4.0-RC1 / 2021-02-08
========================
* update k8s sidecar images to gcr ([#284](https://github.com/openebs/zfs-localpv/pull/284),[@shubham14bajpai](https://github.com/shubham14bajpai))
* adding resize support for raw block volumes ([#281](https://github.com/openebs/zfs-localpv/pull/281),[@pawanpraka1](https://github.com/pawanpraka1))


v1.3.0 / 2021-01-13
========================
* adding capacity weighted scheduler ([#266](https://github.com/openebs/zfs-localpv/pull/266),[@pawanpraka1](https://github.com/pawanpraka1))
* use common lib-csi imports ([#263](https://github.com/openebs/zfs-localpv/pull/263),[@shubham14bajpai](https://github.com/shubham14bajpai))
* Cross Build enviroment bug fixes ([#264](https://github.com/openebs/zfs-localpv/pull/264),[@praveengt](https://github.com/praveengt))
* bump k8s csi to latest stable container images ([#271](https://github.com/openebs/zfs-localpv/pull/271),[@shubham14bajpai](https://github.com/shubham14bajpai))
* creating directory with 0755 permission ([#262](https://github.com/openebs/zfs-localpv/pull/262),[@pawanpraka1](https://github.com/pawanpraka1))

v1.3.0-RC2 / 2021-01-11
========================

v1.3.0-RC1 / 2021-01-09
========================
* adding capacity weighted scheduler ([#266](https://github.com/openebs/zfs-localpv/pull/266),[@pawanpraka1](https://github.com/pawanpraka1))
* use common lib-csi imports ([#263](https://github.com/openebs/zfs-localpv/pull/263),[@shubham14bajpai](https://github.com/shubham14bajpai))
* Cross Build enviroment bug fixes ([#264](https://github.com/openebs/zfs-localpv/pull/264),[@praveengt](https://github.com/praveengt))
* bump k8s csi to latest stable container images ([#271](https://github.com/openebs/zfs-localpv/pull/271),[@shubham14bajpai](https://github.com/shubham14bajpai))
* creating directory with 0755 permission ([#262](https://github.com/openebs/zfs-localpv/pull/262),[@pawanpraka1](https://github.com/pawanpraka1))


v1.2.1 / 2020-12-15
========================
* fixing idempotency check for the mount path ([#260](https://github.com/openebs/zfs-localpv/pull/260),[@pawanpraka1](https://github.com/pawanpraka1))


v1.2.0 / 2020-12-13
========================

* removing quay from kustomization.yaml as we are using multiarch docker images ([#248](https://github.com/openebs/zfs-localpv/pull/248),[@pawanpraka1](https://github.com/pawanpraka1))
* move xfs and mount code out of zfs package ([#245](https://github.com/openebs/zfs-localpv/pull/245),[@pawanpraka1](https://github.com/pawanpraka1))
* move btrfs code out of zfs package ([#244](https://github.com/openebs/zfs-localpv/pull/244),[@pawanpraka1](https://github.com/pawanpraka1))
* add github action for chart test and release ([#250](https://github.com/openebs/zfs-localpv/pull/250),[@shubham14bajpai](https://github.com/shubham14bajpai))
* fixing flaky sanity test case ([#256](https://github.com/openebs/zfs-localpv/pull/256),[@pawanpraka1](https://github.com/pawanpraka1))
* refactor scheduler for ZFS-LocalPV ([#249](https://github.com/openebs/zfs-localpv/pull/249),[@pawanpraka1](https://github.com/pawanpraka1))
* moving to ubuntu bionic(18.04 LTS) docker image ([#255](https://github.com/openebs/zfs-localpv/pull/255),[@pawanpraka1](https://github.com/pawanpraka1))
* fixed the kustomize yaml name to kustomization.yaml ([#243](https://github.com/openebs/zfs-localpv/pull/243),[@pawanpraka1](https://github.com/pawanpraka1))
* adding CSI Sanity test for ZFS-LocalPV ([#232](https://github.com/openebs/zfs-localpv/pull/232),[@pawanpraka1](https://github.com/pawanpraka1))

v1.2.0-RC2 / 2020-12-12
========================

v1.2.0-RC1 / 2020-12-10
========================
* removing quay from kustomization.yaml as we are using multiarch docker images ([#248](https://github.com/openebs/zfs-localpv/pull/248),[@pawanpraka1](https://github.com/pawanpraka1))
* move xfs and mount code out of zfs package ([#245](https://github.com/openebs/zfs-localpv/pull/245),[@pawanpraka1](https://github.com/pawanpraka1))
* move btrfs code out of zfs package ([#244](https://github.com/openebs/zfs-localpv/pull/244),[@pawanpraka1](https://github.com/pawanpraka1))
* add github action for chart test and release ([#250](https://github.com/openebs/zfs-localpv/pull/250),[@shubham14bajpai](https://github.com/shubham14bajpai))
* fixing flaky sanity test case ([#256](https://github.com/openebs/zfs-localpv/pull/256),[@pawanpraka1](https://github.com/pawanpraka1))
* refactor scheduler for ZFS-LocalPV ([#249](https://github.com/openebs/zfs-localpv/pull/249),[@pawanpraka1](https://github.com/pawanpraka1))
* moving to ubuntu bionic(18.04 LTS) docker image ([#255](https://github.com/openebs/zfs-localpv/pull/255),[@pawanpraka1](https://github.com/pawanpraka1))
* fixed the kustomize yaml name to kustomization.yaml ([#243](https://github.com/openebs/zfs-localpv/pull/243),[@pawanpraka1](https://github.com/pawanpraka1))
* adding CSI Sanity test for ZFS-LocalPV ([#232](https://github.com/openebs/zfs-localpv/pull/232),[@pawanpraka1](https://github.com/pawanpraka1))


v1.1.0 / 2020-11-14
========================
* changing the zfs-driver images to multi-arch docker hub ([#237](https://github.com/openebs/zfs-localpv/pull/237),[@w3aman](https://github.com/w3aman))
* Remove MountInfo struct from the api files ([#225](https://github.com/openebs/zfs-localpv/pull/225),[@codegagan](https://github.com/codegagan))
* adding deployment yaml via kustomize ([#231](https://github.com/openebs/zfs-localpv/pull/231),[@pawanpraka1](https://github.com/pawanpraka1))
* add support for creating the Clone from volume as datasource ([#234](https://github.com/openebs/zfs-localpv/pull/234),[@pawanpraka1](https://github.com/pawanpraka1))
* add support for multi arch container image ([#233](https://github.com/openebs/zfs-localpv/pull/233),[@prateekpandey14](https://github.com/prateekpandey14))
* support parallel/faster upgrades for node daemonset ([#230](https://github.com/openebs/zfs-localpv/pull/230),[@pawanpraka1](https://github.com/pawanpraka1))

v1.1.0-RC2 / 2020-11-13
========================

v1.1.0-RC1 / 2020-11-12
========================
* Remove MountInfo struct from the api files ([#225](https://github.com/openebs/zfs-localpv/pull/225),[@codegagan](https://github.com/codegagan))
* adding deployment yaml via kustomize ([#231](https://github.com/openebs/zfs-localpv/pull/231),[@pawanpraka1](https://github.com/pawanpraka1))
* add support for creating the Clone from volume as datasource ([#234](https://github.com/openebs/zfs-localpv/pull/234),[@pawanpraka1](https://github.com/pawanpraka1))
* add support for multi arch container image ([#233](https://github.com/openebs/zfs-localpv/pull/233),[@prateekpandey14](https://github.com/prateekpandey14))
* support parallel/faster upgrades for node daemonset ([#230](https://github.com/openebs/zfs-localpv/pull/230),[@pawanpraka1](https://github.com/pawanpraka1))


v1.0.1 / 2020-10-14
========================
* removing centos yamls from the repo ([#211](https://github.com/openebs/zfs-localpv/pull/211),[@pawanpraka1](https://github.com/pawanpraka1))
* adding validation for backup and restore ([#221](https://github.com/openebs/zfs-localpv/pull/221),[@pawanpraka1](https://github.com/pawanpraka1))


v1.0.1-RC2 / 2020-10-12
========================

v1.0.1-RC1 / 2020-10-08
========================
* removing centos yamls from the repo ([#211](https://github.com/openebs/zfs-localpv/pull/211),[@pawanpraka1](https://github.com/pawanpraka1))
* adding validation for backup and restore ([#221](https://github.com/openebs/zfs-localpv/pull/221),[@pawanpraka1](https://github.com/pawanpraka1))


v1.0.0 / 2020-09-15
========================
* adding velero backup and restore support ([#162](https://github.com/openebs/zfs-localpv/pull/162),[@pawanpraka1](https://github.com/pawanpraka1))
* update go version to 1.14.7 ([#201](https://github.com/openebs/zfs-localpv/pull/201),[@pawanpraka1](https://github.com/pawanpraka1))
* mounting the root filesystem to remove the dependency on the Operating system ([#204](https://github.com/openebs/zfs-localpv/pull/204),[@pawanpraka1](https://github.com/pawanpraka1))
* Add license-check for .go , .sh , Dockerfile and Makefile ([#205](https://github.com/openebs/zfs-localpv/pull/205),[@ajeetrai707](https://github.com/AJEETRAI707))

v1.0.0-RC2 / 2020-09-14
========================

v1.0.0-RC1 / 2020-09-10
========================
* adding velero backup and restore support ([#162](https://github.com/openebs/zfs-localpv/pull/162),[@pawanpraka1](https://github.com/pawanpraka1))
* update go version to 1.14.7 ([#201](https://github.com/openebs/zfs-localpv/pull/201),[@pawanpraka1](https://github.com/pawanpraka1))
* mounting the root filesystem to remove the dependency on the Operating system ([#204](https://github.com/openebs/zfs-localpv/pull/204),[@pawanpraka1](https://github.com/pawanpraka1))
* Add license-check for .go , .sh , Dockerfile and Makefile ([#205](https://github.com/openebs/zfs-localpv/pull/205),[@ajeetrai707](https://github.com/AJEETRAI707))


v0.9.2 / 2020-08-26
========================
* Reverting back to old way of checking the volume status ([#196](https://github.com/openebs/zfs-localpv/pull/196),[@pawanpraka1](https://github.com/pawanpraka1))

v0.9.1 / 2020-08-14
========================
* mounting the volume if it is ready ([#184](https://github.com/openebs/zfs-localpv/pull/184),[@pawanpraka1](https://github.com/pawanpraka1))
* fixed uuid generation issue when mount fails ([#183](https://github.com/openebs/zfs-localpv/pull/183),[@pawanpraka1](https://github.com/pawanpraka1))
* rounding off the volume size to Gi and Mi ([#191](https://github.com/openebs/zfs-localpv/pull/191),[@pawanpraka1](https://github.com/pawanpraka1))
* removing volumeLifecycleModes from the operator yaml ([#186](https://github.com/openebs/zfs-localpv/pull/186),[@pawanpraka1](https://github.com/pawanpraka1))

v0.9.1-RC2 / 2020-08-12
========================

v0.9.1-RC1 / 2020-08-10
========================
* mounting the volume if it is ready ([#184](https://github.com/openebs/zfs-localpv/pull/184),[@pawanpraka1](https://github.com/pawanpraka1))
* fixed uuid generation issue when mount fails ([#183](https://github.com/openebs/zfs-localpv/pull/183),[@pawanpraka1](https://github.com/pawanpraka1))
* rounding off the volume size to Gi and Mi ([#191](https://github.com/openebs/zfs-localpv/pull/191),[@pawanpraka1](https://github.com/pawanpraka1))
* removing volumeLifecycleModes from the operator yaml ([#186](https://github.com/openebs/zfs-localpv/pull/186),[@pawanpraka1](https://github.com/pawanpraka1))


v0.9.0 / 2020-07-14
========================
* fixing xfs mounting issue on centos with ubuntu 20.04 image ([#179](https://github.com/openebs/zfs-localpv/pull/179),[@pawanpraka1](https://github.com/pawanpraka1))
* change logger from Sirupsen/logrus to klog ([#166](https://github.com/openebs/zfs-localpv/pull/166),[@vaniisgh](https://github.com/vaniisgh))
* Add checks to ensure zfs-driver status is running in BDD test ([#171](https://github.com/openebs/zfs-localpv/pull/171),[@vaniisgh](https://github.com/vaniisgh))
* fixing duplicate UUID issue with btrfs ([#172](https://github.com/openebs/zfs-localpv/pull/172),[@pawanpraka1](https://github.com/pawanpraka1))
* adding shared mount support ZFSPV volumes ([#164](https://github.com/openebs/zfs-localpv/pull/164),[@pawanpraka1](https://github.com/pawanpraka1))
* update docs to reflect gomod migration  ([#160](https://github.com/openebs/zfs-localpv/pull/160),[@vaniisgh](https://github.com/vaniisgh))
* add golint target to makefile ([#167](https://github.com/openebs/zfs-localpv/pull/167),[@vaniisgh](https://github.com/vaniisgh))
* adding support to have btrfs filesystem for ZFS-LocalPV ([#170](https://github.com/openebs/zfs-localpv/pull/170),[@pawanpraka1](https://github.com/pawanpraka1))
* adds a filter for grpc logs to reduce the pollution ([#161](https://github.com/openebs/zfs-localpv/pull/161),[@vaniisgh](https://github.com/vaniisgh))
* adding snapshot and clone releated test cases in BDD ([#174](https://github.com/openebs/zfs-localpv/pull/174),[@pawanpraka1](https://github.com/pawanpraka1))
* add golint to travis & fix linting ([#175](https://github.com/openebs/zfs-localpv/pull/175),[@vaniisgh](https://github.com/vaniisgh))


v0.9.0-RC2 / 2020-07-11
========================
* fixing xfs mounting issue on centos with ubuntu 20.04 image ([#179](https://github.com/openebs/zfs-localpv/pull/179),[@pawanpraka1](https://github.com/pawanpraka1))

v0.9.0-RC1 / 2020-07-08
========================
* change logger from Sirupsen/logrus to klog ([#166](https://github.com/openebs/zfs-localpv/pull/166),[@vaniisgh](https://github.com/vaniisgh))
* Add checks to ensure zfs-driver status is running in BDD test ([#171](https://github.com/openebs/zfs-localpv/pull/171),[@vaniisgh](https://github.com/vaniisgh))
* fixing duplicate UUID issue with btrfs ([#172](https://github.com/openebs/zfs-localpv/pull/172),[@pawanpraka1](https://github.com/pawanpraka1))
* adding shared mount support ZFSPV volumes ([#164](https://github.com/openebs/zfs-localpv/pull/164),[@pawanpraka1](https://github.com/pawanpraka1))
* update docs to reflect gomod migration  ([#160](https://github.com/openebs/zfs-localpv/pull/160),[@vaniisgh](https://github.com/vaniisgh))
* add golint target to makefile ([#167](https://github.com/openebs/zfs-localpv/pull/167),[@vaniisgh](https://github.com/vaniisgh))
* adding support to have btrfs filesystem for ZFS-LocalPV ([#170](https://github.com/openebs/zfs-localpv/pull/170),[@pawanpraka1](https://github.com/pawanpraka1))
* adds a filter for grpc logs to reduce the pollution ([#161](https://github.com/openebs/zfs-localpv/pull/161),[@vaniisgh](https://github.com/vaniisgh))
* adding snapshot and clone releated test cases in BDD ([#174](https://github.com/openebs/zfs-localpv/pull/174),[@pawanpraka1](https://github.com/pawanpraka1))
* add golint to travis & fix linting ([#175](https://github.com/openebs/zfs-localpv/pull/175),[@vaniisgh](https://github.com/vaniisgh))


v0.8.0 / 2020-06-13
========================

  * changing image pull policy to IfNotPresent to make it not pull the image again and again ([#124](https://github.com/openebs/zfs-localpv/pull/124),[@pawanpraka1](https://github.com/pawanpraka1))
  * moving to legacy mount ([#151](https://github.com/openebs/zfs-localpv/pull/151),[@pawanpraka1](https://github.com/pawanpraka1))
  * Fixes an issue where volumes meant to be filesystem datasets got created as zvols and generally makes storageclass parameter spelling insensitive to case ([#144](https://github.com/openebs/zfs-localpv/pull/144),[@cruwe](https://github.com/cruwe))
  * include pvc name in volume events ([#150](https://github.com/openebs/zfs-localpv/pull/150),[@pawanpraka1](https://github.com/pawanpraka1))
  * Fixes an issue where PVC was bound to unusable PV created using incorrect values provided in PVC/Storageclass ([#121](https://github.com/openebs/zfs-localpv/pull/121),[@pawanpraka1](https://github.com/pawanpraka1))
  * adding v1 CRD for ZFS-LocalPV  ([#140](https://github.com/openebs/zfs-localpv/pull/140),[@pawanpraka1](https://github.com/pawanpraka1))
  * add contributing checkout list ([#138](https://github.com/openebs/zfs-localpv/pull/138),[@Icedroid](https://github.com/Icedroid))
  * fixing golint warnings ([#133](https://github.com/openebs/zfs-localpv/pull/133),[@Icedroid](https://github.com/Icedroid))
  * removing unnecessary printer columns from ZFSVolume ([#128](https://github.com/openebs/zfs-localpv/pull/128),[@pawanpraka1](https://github.com/pawanpraka1))
  * fixing stale ZFSVolume resource issue when deleting the pvc in pending state ([#145](https://github.com/openebs/zfs-localpv/pull/145),[@pawanpraka1](https://github.com/pawanpraka1))
  * Updated the doc for custom-topology support ([#122](https://github.com/openebs/zfs-localpv/pull/122),[@w3aman](https://github.com/w3aman))
  * adding operator yaml for centos7 and centos8 ([#149](https://github.com/openebs/zfs-localpv/pull/149),[@pawanpraka1](https://github.com/pawanpraka1))
  * honouring readonly flag for ZFS-LocalPV ([#137](https://github.com/openebs/zfs-localpv/pull/137),[@pawanpraka1](https://github.com/pawanpraka1))

v0.8.0-RC2 / 2020-06-12
========================

v0.8.0-RC1 / 2020-06-10
========================

  * changing image pull policy to IfNotPresent to make it not pull the image again and again ([#124](https://github.com/openebs/zfs-localpv/pull/124),[@pawanpraka1](https://github.com/pawanpraka1))
  * moving to legacy mount ([#151](https://github.com/openebs/zfs-localpv/pull/151),[@pawanpraka1](https://github.com/pawanpraka1))
  * Fixes an issue where volumes meant to be filesystem datasets got created as zvols and generally makes storageclass parameter spelling insensitive to case ([#144](https://github.com/openebs/zfs-localpv/pull/144),[@cruwe](https://github.com/cruwe))
  * include pvc name in volume events ([#150](https://github.com/openebs/zfs-localpv/pull/150),[@pawanpraka1](https://github.com/pawanpraka1))
  * Fixes an issue where PVC was bound to unusable PV created using incorrect values provided in PVC/Storageclass ([#121](https://github.com/openebs/zfs-localpv/pull/121),[@pawanpraka1](https://github.com/pawanpraka1))
  * adding v1 CRD for ZFS-LocalPV  ([#140](https://github.com/openebs/zfs-localpv/pull/140),[@pawanpraka1](https://github.com/pawanpraka1))
  * add contributing checkout list ([#138](https://github.com/openebs/zfs-localpv/pull/138),[@Icedroid](https://github.com/Icedroid))
  * fixing golint warnings ([#133](https://github.com/openebs/zfs-localpv/pull/133),[@Icedroid](https://github.com/Icedroid))
  * removing unnecessary printer columns from ZFSVolume ([#128](https://github.com/openebs/zfs-localpv/pull/128),[@pawanpraka1](https://github.com/pawanpraka1))
  * fixing stale ZFSVolume resource issue when deleting the pvc in pending state ([#145](https://github.com/openebs/zfs-localpv/pull/145),[@pawanpraka1](https://github.com/pawanpraka1))
  * Updated the doc for custom-topology support ([#122](https://github.com/openebs/zfs-localpv/pull/122),[@w3aman](https://github.com/w3aman))
  * adding operator yaml for centos7 and centos8 ([#149](https://github.com/openebs/zfs-localpv/pull/149),[@pawanpraka1](https://github.com/pawanpraka1))
  * honouring readonly flag for ZFS-LocalPV ([#137](https://github.com/openebs/zfs-localpv/pull/137),[@pawanpraka1](https://github.com/pawanpraka1))


v0.7.0 / 2020-05-15
=======================

  * feat(grafana): adding basic grafana dashboard ([110](https://github.com/openebs/zfs-localpv/pull/110),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(version): use the travis tag for the version ([114](https://github.com/openebs/zfs-localpv/pull/114),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(README): Fix the link in README file to the raw-block-volume.md file ([109](https://github.com/openebs/zfs-localpv/pull/109),
  [@w3aman](https://github.com/w3aman))
  * chore(import-vol): adding steps to import existing volume to ZFS-LocalPV ([108](https://github.com/openebs/zfs-localpv/pull/108),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): adding raw block volume details in README ([106](https://github.com/openebs/zfs-localpv/pull/106),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * refact(build):trim leading v from image tag ([105](https://github.com/openebs/zfs-localpv/pull/105),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * refact(build): make the docker images configurable ([104](https://github.com/openebs/zfs-localpv/pull/104),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(block): adding block volume support for ZFSPV ([102](https://github.com/openebs/zfs-localpv/pull/102),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(topokey): changing topology key to unique name ([101](https://github.com/openebs/zfs-localpv/pull/101),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * docs(project): adding project specific files ([99](https://github.com/openebs/zfs-localpv/pull/99),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(doc , format): Fixed the formatting of ReadME file for upgrade ([98](https://github.com/openebs/zfs-localpv/pull/98),
  [@w3aman](https://github.com/w3aman))
  * feat(topology): adding support for custom topology keys ([94](https://github.com/openebs/zfs-localpv/pull/94),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * added developer environment examples ([#92](https://github.com/openebs/zfs-localpv/pull/92),
  [@filippobosi](https://github.com/filippobosi))

v0.7.0-RC2 / 2020-05-13
=======================


v0.7.0-RC1 / 2020-05-08
=======================

  * fix(README): Fix the link in README file to the raw-block-volume.md file ([109](https://github.com/openebs/zfs-localpv/pull/109),
  [@w3aman](https://github.com/w3aman))
  * chore(import-vol): adding steps to import existing volume to ZFS-LocalPV ([108](https://github.com/openebs/zfs-localpv/pull/108),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): adding raw block volume details in README ([106](https://github.com/openebs/zfs-localpv/pull/106),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * refact(build):trim leading v from image tag ([105](https://github.com/openebs/zfs-localpv/pull/105),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * refact(build): make the docker images configurable ([104](https://github.com/openebs/zfs-localpv/pull/104),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(block): adding block volume support for ZFSPV ([102](https://github.com/openebs/zfs-localpv/pull/102),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(topokey): changing topology key to unique name ([101](https://github.com/openebs/zfs-localpv/pull/101),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * docs(project): adding project specific files ([99](https://github.com/openebs/zfs-localpv/pull/99),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(doc , format): Fixed the formatting of ReadME file for upgrade ([98](https://github.com/openebs/zfs-localpv/pull/98),
  [@w3aman](https://github.com/w3aman))
  * feat(topology): adding support for custom topology keys ([94](https://github.com/openebs/zfs-localpv/pull/94),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * added developer environment examples ([#92](https://github.com/openebs/zfs-localpv/pull/92),
  [@filippobosi](https://github.com/filippobosi))

0.6.1 / 2020-04-23
=======================

  * potential data loss in case of pod deletion ([#89](https://github.com/openebs/zfs-localpv/pull/89),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * avoid creation of volumeattachment object to fix slow volume attachment ([#85](https://github.com/openebs/zfs-localpv/pull/85),
  [@pawanpraka1](https://github.com/pawanpraka1))

0.6.0 / 2020-04-14
=======================

  * feat(validation): adding validation for ZFSPV CR parameters ([#66](https://github.com/openebs/zfs-localpv/pull/66),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(zfspv): adding poolname info to the PV volumeattributes ([#80](https://github.com/openebs/zfs-localpv/pull/80),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(zfspv): handling unmounted volume ([#78](https://github.com/openebs/zfs-localpv/pull/78),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(crd-gen): automate the CRDs generation with validations for APIs ([#75](https://github.com/openebs/zfs-localpv/pull/75),
  [@prateekpandey14](https://github.com/prateekpandey14))
  * feat(crd): scripts to help migrating to new CRDs ([#73](https://github.com/openebs/zfs-localpv/pull/73),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * refactor(crd): move CR from openebs.io to zfs.openebs.io ([#70](https://github.com/openebs/zfs-localpv/pull/70),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(zfspv): Upgrade the base ubuntu package ([#68](https://github.com/openebs/zfs-localpv/pull/68),
  [@stevefan1999-personal](https://github.com/stevefan1999-personal))
  * fix(test): fixing resize flaky test case ([#71](https://github.com/openebs/zfs-localpv/pull/71),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(xfs): clearing the xfs log before generating UUID ([#64](https://github.com/openebs/zfs-localpv/pull/64),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(readme): adding e2e project link in README ([#65](https://github.com/openebs/zfs-localpv/pull/65),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(xfs): fixing xfs duplicate uuid for cloned volumes ([#63](https://github.com/openebs/zfs-localpv/pull/63),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(version): Makefile and version enhancement ([#62](https://github.com/openebs/zfs-localpv/pull/62),
  [@pawanpraka1](https://github.com/pawanpraka1))

v0.6-RC2 / 2020-04-11
=======================

  * feat(zfspv): handling unmounted volume ([#78](https://github.com/openebs/zfs-localpv/pull/78),
  [@pawanpraka1](https://github.com/pawanpraka1))

v0.6-RC1 / 2020-04-08
=======================

  * feat(crd-gen): automate the CRDs generation with validations for APIs ([#75](https://github.com/openebs/zfs-localpv/pull/75),
  [@prateekpandey14](https://github.com/prateekpandey14))
  * feat(crd): scripts to help migrating to new CRDs ([#73](https://github.com/openebs/zfs-localpv/pull/73),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * refactor(crd): move CR from openebs.io to zfs.openebs.io ([#70](https://github.com/openebs/zfs-localpv/pull/70),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(zfspv): Upgrade the base ubuntu package ([#68](https://github.com/openebs/zfs-localpv/pull/68),
  [@stevefan1999-personal](https://github.com/stevefan1999-personal))
  * fix(test): fixing resize flaky test case ([#71](https://github.com/openebs/zfs-localpv/pull/71),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(xfs): clearing the xfs log before generating UUID ([#64](https://github.com/openebs/zfs-localpv/pull/64),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(readme): adding e2e project link in README ([#65](https://github.com/openebs/zfs-localpv/pull/65),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(xfs): fixing xfs duplicate uuid for cloned volumes ([#63](https://github.com/openebs/zfs-localpv/pull/63),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(version): Makefile and version enhancement ([#62](https://github.com/openebs/zfs-localpv/pull/62),
  [@pawanpraka1](https://github.com/pawanpraka1))

v0.5 / 2020-03-14
=======================

  * fix(clone): setting properties on the clone volume ([#57](https://github.com/openebs/zfs-localpv/pull/57),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): adding resize details in README ([#53](https://github.com/openebs/zfs-localpv/pull/53),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(resize): adding BDD test for Online volume expansion ([#52](https://github.com/openebs/zfs-localpv/pull/52),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(resize): adding Online volume expansion support for ZFSPV ([#51](https://github.com/openebs/zfs-localpv/pull/51),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(analytics): adding google analytics for ZFSPV ([#49](https://github.com/openebs/zfs-localpv/pull/49),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): updating readme with snapshot and clone details ([#48](https://github.com/openebs/zfs-localpv/pull/48),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc):  Adding the list of e2e test cases ([#50](https://github.com/openebs/zfs-localpv/pull/50),
  [@w3aman](https://github.com/w3aman))
  * fix(operator): update provisioner image to support snapshot datasource ([#46](https://github.com/openebs/zfs-localpv/pull/46),
  [@prateekpandey14](https://github.com/prateekpandey14))

v0.5-RC2 / 2020-03-12
=======================

  * fix(clone): setting properties on the clone volume ([#57](https://github.com/openebs/zfs-localpv/pull/57),
  [@pawanpraka1](https://github.com/pawanpraka1))

v0.5-RC1 / 2020-03-06
=======================

  * chore(doc): adding resize details in README ([#53](https://github.com/openebs/zfs-localpv/pull/53),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(resize): adding BDD test for Online volume expansion ([#52](https://github.com/openebs/zfs-localpv/pull/52),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(resize): adding Online volume expansion support for ZFSPV ([#51](https://github.com/openebs/zfs-localpv/pull/51),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(analytics): adding google analytics for ZFSPV ([#49](https://github.com/openebs/zfs-localpv/pull/49),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): updating readme with snapshot and clone details ([#48](https://github.com/openebs/zfs-localpv/pull/48),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc):  Adding the list of e2e test cases ([#50](https://github.com/openebs/zfs-localpv/pull/50),
  [@w3aman](https://github.com/w3aman))
  * fix(operator): update provisioner image to support snapshot datasource ([#46](https://github.com/openebs/zfs-localpv/pull/46),
  [@prateekpandey14](https://github.com/prateekpandey14))

v0.4 / 2020-02-13
=======================

  * feat(zfspv): adding snapshot and clone support for ZFSPV ([#39](https://github.com/openebs/zfs-localpv/pull/32),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(zfspv): do not destroy the dataset with -R option ([#40](https://github.com/openebs/zfs-localpv/pull/40),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix(doc): Resolving the typo error in README doc ([#38](https://github.com/openebs/zfs-localpv/pull/38),
  [@w3aman](https://github.com/w3aman))
  * chore(metrics): adding list of zfs metrics exposed by prometheus ([#36](https://github.com/openebs/zfs-localpv/pull/36),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * refactor(version): bumping the version to 0.4 ([#37](https://github.com/openebs/zfs-localpv/pull/37),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): adding v0.3 changelog in the repo ([#35](https://github.com/openebs/zfs-localpv/pull/35),
  [@pawanpraka1](https://github.com/pawanpraka1))

v0.3 / 2020-01-15
=======================

  * feat(alert): adding sample prometheus rules for ZFSPV ([#32](https://github.com/openebs/zfs-localpv/pull/32),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(HA): adding support to have controller in HA  ([#31](https://github.com/openebs/zfs-localpv/pull/31),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): adding contributing and faq doc ([#29](https://github.com/openebs/zfs-localpv/pull/29),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * feat(stats): adding volume usage stats ([#27](https://github.com/openebs/zfs-localpv/pull/27),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): making zfs-localpv repository CNCF compatible ([#26](https://github.com/openebs/zfs-localpv/pull/26),
  [#28](https://github.com/openebs/zfs-localpv/pull/28), [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): adding roadmap in the README ([#25](https://github.com/openebs/zfs-localpv/pull/25),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * test(zfspv): adding test cases to verify zfs property update ([#24](https://github.com/openebs/zfs-localpv/pull/24),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * chore(doc): adding changelog in the repo ([#23](https://github.com/openebs/zfs-localpv/pull/23),
  [@pawanpraka1](https://github.com/pawanpraka1))

v0.2 / 2019-12-09
=======================

  * making test cases to run on forked repo ([#22](https://github.com/openebs/zfs-localpv/pull/22),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * integration test cases for ZFSPV ([#21](https://github.com/openebs/zfs-localpv/pull/21),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * changing image pull policy to IfNotPresent ([#20](https://github.com/openebs/zfs-localpv/pull/20),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * renamed watcher to mgmt package ([#19](https://github.com/openebs/zfs-localpv/pull/19),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fixing mongo yaml ([#18](https://github.com/openebs/zfs-localpv/pull/18),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * updating readme with latest details ([#17](https://github.com/openebs/zfs-localpv/pull/17),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fix scheduling algorithm doc ([#16](https://github.com/openebs/zfs-localpv/pull/16),
  [@akhilerm](https://github.com/akhilerm))
  * adding support for applications to create "zfs" filesystem ([#15](https://github.com/openebs/zfs-localpv/pull/15),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * fixed a typo for thinprovision json name. ([#14](https://github.com/openebs/zfs-localpv/pull/14),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * updating sample ZFSVolume CR ([#13](https://github.com/openebs/zfs-localpv/pull/13),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * remove unnecessary deploy from travis ([#12](https://github.com/openebs/zfs-localpv/pull/12),
  [@pawanpraka1](https://github.com/pawanpraka1))

v0.1 / 2019-11-07
=======================

  * updating readme with latest details ([#11](https://github.com/openebs/zfs-localpv/pull/11),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * adding xfs filesystem support for zfs-localpv ([#10](https://github.com/openebs/zfs-localpv/pull/10),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * volume count based scheduler for ZFSPV ([#8](https://github.com/openebs/zfs-localpv/pull/8),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * adding topology support for zfspv ([#7](https://github.com/openebs/zfs-localpv/pull/7),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * adding encryption in ZFSVolume CR ([#6](https://github.com/openebs/zfs-localpv/pull/6),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * updating README with volume property usage ([#5](https://github.com/openebs/zfs-localpv/pull/5),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * not able to deploy on rancher with ZFS 0.8 ([#4](https://github.com/openebs/zfs-localpv/pull/4),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * Add license scan report and status ([#3](https://github.com/openebs/zfs-localpv/pull/3),
  [@fossabot](https://github.com/fossabot))
  * adding README for ZFSPV ([#2](https://github.com/openebs/zfs-localpv/pull/2),
  [@pawanpraka1](https://github.com/pawanpraka1))
  * Initial commit for provisioning and deprovisioning the volumes ([#1](https://github.com/openebs/zfs-localpv/pull/1),
  [@pawanpraka1](https://github.com/pawanpraka1))

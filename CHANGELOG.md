v2.8.0 / To be Released
========================

## New Features

- **Backup garbage collector**
  - Disabled by default. This feature clean up ZFSBackups which `prevSnapName` is missing in the cluster. Fixing [openebs/velero-plugin#119](https://github.com/openebs/velero-plugin/issues/119) deleting orphaned backups and fixing frequency of full backups creation.


v2.7.1 / 2025-02-12
========================

The `2.7.1` release of OpenEBS Local PV ZFS focuses fixing a bug which prevented volume clones upon upgrade. This bug got introduced because of a new feature being  added for the selection of the `quota` and `requota`, but the backwards compatibility was not taken care of. This has also highlighted the gaps on upgrade testing, which we would like to work upon in the upcoming releases

---

## Bug Fixes and Stability Improvements
- fix: quota property empty on upgrade to ensure backwards compatibility (#629 by @Abhinandan-Purkait)


v2.7.0 / 2025-02-10
========================

The 2.7.0 release of zfs‑localpv focuses on enhancing the stability, feature set, and developer experience. Key highlights include new configuration options for quota management, support for faster compression algorithms, fixes to race conditions and resource cleanup, and a raft of improvements to tests, CI workflows, and documentation. In addition, several maintenance tasks and dependency updates have been performed to keep the project robust and secure.

**NOTE: The 2.7.0 release of OpenEBS Local PV ZFS has bug which affects the volume clone feature upon upgrade from previous version. It has been fixed via a patch release [2.7.1](https://github.com/openebs/zfs-localpv/releases/tag/v2.7.1). We would recommend users to upgrade directly to 2.7.1, if coming from 2.6.x and below**

---

## New Features

- **Configurable Quota Options**  
  - An option has been added to let users choose between using _refquota_ and _quota_ for ZFS volumes. This change (#542 by @cinapm) gives administrators more flexibility when managing resource limits.

- **Enhanced Compression Support**  
  - Support for the _zstd‑fast_ algorithm has been introduced (#597 by @Abhinandan‑Purkait). This new option improves performance when compression is desired on ZFS volumes.

---

## Bug Fixes and Stability Improvements

- **Volume Provisioning and Controller Fixes**  
  - The plugin now correctly retrieves the owner node id (#549).  
  - A fix ensures that if a ZFS volume already exists, the controller will provision the volume without error (#576 by @AChangFeng).  
  - Several race conditions in the CSI controller have been addressed:
    - A per‑volume mutex was introduced to prevent simultaneous CSI controller calls that might cause the volume CR to be inadvertently deleted (#588 by @Lucaber and #613 by @sinhaashish).  
    - The ZFS timer used during volume creation is now properly stopped after volume creation completes (#600 by @rfyiamcool).

- **YAML and CRD Corrections**  
  - VolumeSnapshot CRDs now have identation fixes (#620 by @nilroy).  
  - Minor formatting adjustments (such as indent fixes for imagePullSecrets in the deployment charts - see #596 by @chris199512).

- **Reservation and Deployment Fixes**  
  - A bug in the reservation logic during volume expansion (with refquota settings) has been resolved (#595 by @abuisine).  
  - Controller fixes prevent accidental deletion of volume CRs when snapshots exist (#613 by @sinhaashish).

---

## Testing Enhancements

- **BDD and Integration Tests**  
  - New BDD tests for CI have been added (#551 and #556 by @sinhaashish), improving confidence in test outcomes.
  
- **Snapshot and Clone Testing**  
  - Snapshot and clone tests for raw block volumes have been introduced (#559, #570 by @sinhaashish).  
  - Thin provision and volume type specific tests (#572 by @sinhaashish).  
  - Additional tests now verify quota and refquota parameters (#608 by @Abhinandan‑Purkait).  
  - Clone tests have been corrected so that the clone deployment uses the proper clone app and volume (#612 by @tiagolobocastro).

- **Local Testing Improvements**  
  - Various enhancements to local testing setups have been applied (#609 by @tiagolobocastro) to help developers run tests without elevated privileges and with more predictable behavior.

---

## Continuous Integration and Deployment

- **Workflow Enhancements**  
  - The pull_request workflow has been enhanced to improve reliability (#557 by @Abhinandan‑Purkait).  
  - The build.yml workflow received improvements for efficiency (#565 by @Abhinandan‑Purkait).  
  - CI now includes branch preparation changes (#567 by @Abhinandan‑Purkait) and explicit namespace settings (#580 by @sinhaashish).

- **New CI Features**  
  - A Fossa CLI workflow has been integrated to automatically check licensing issues (#599 by @Abhinandan‑Purkait).  
  - The CI environment now pins Ubuntu to 22.04 to ensure consistency for minikube (#604 by @Abhinandan‑Purkait).  
  - Updates to the build_and_push action (addressing missing changes) are included (#618 by @tiagolobocastro).

---

## Documentation and Contributor Workflow

- **Documentation Updates**  
  - Typos and minor errors in the README have been fixed (#554 by @druesendieb), and several docs (such as the localpv parameter explanations in docs and backup‑restore guides) have been updated (#563, #585).  
  - A security section now cross‑references relevant security documents (#611 by @tiagolobocastro).

- **Contributor and Release Process Improvements**  
  - The contributor workflow documentation has been improved to provide clearer guidelines on how to contribute (#616 by @tiagolobocastro).  
  - The overall README and additional documentation have been tidied up, and outdated assets have been moved or removed (#619 by @Abhinandan‑Purkait).  
  - Changes to the RBAC configuration were applied (#603 by @d4rkfella) to ensure that all components (including CSI snapshotter) have the correct permissions.

---

## Dependency and Maintenance Updates

- **CRD Generation and Cleanup**  
  - The CRDs have been replaced with an auto‑generated copy (#564 by @niladrih) to reduce manual errors.  
  - Unused scripts have been removed and the make manifests updated (#569 by @Abhinandan‑Purkait).

- **Dependency Bumps**  
  - The analytics dependency has been updated (#578 by @niladrih).  
  - The Go networking package has been bumped from 0.28.0 to 0.33.0 (#610 by dependabot).

- **Chart and Label Adjustments**  
  - A fix was made to move the `app=componentName` label out of the csi‑node matchLabels section to prevent upgrade issues (#605 by @niladrih).
---

v2.6.2 / 2024-09-25
========================
* fix(chart): handle trailing slash (/) in csi plugin kubelet directory ([#532](https://github.com/openebs/zfs-localpv/pull/532),[@w3aman](https://github.com/w3aman))
* fix(chart): remove anti-affinity from csi controller ([#552](https://github.com/openebs/zfs-localpv/pull/552),[@Abhinandan-Purkait](https://github.com/Abhinandan-Purkait))

v2.6.1 / 2024-09-17
========================
* ci: enhance pull_request workflow ([#557](https://github.com/openebs/zfs-localpv/pull/557),[@Abhinandan-Purkait](https://github.com/Abhinandan-Purkait))
* ci: add branch preparation and release CI changes ([#567](https://github.com/openebs/zfs-localpv/pull/567),[@Abhinandan-Purkait](https://github.com/Abhinandan-Purkait))
* chore: cleanup unused scripts, update make manifests ([#569](https://github.com/openebs/zfs-localpv/pull/569),[@Abhinandan-Purkait](https://github.com/Abhinandan-Purkait))
* chore: replace CRD with auto-generated copy ([#564](https://github.com/openebs/zfs-localpv/pull/548),[@niladrih](https://github.com/niladrih))
* chore(deps): update analytics dependency ([#578](https://github.com/openebs/zfs-localpv/pull/578),[@niladrih](https://github.com/niladrih))
* fix: chart.yaml indentation ([#586](https://github.com/openebs/zfs-localpv/pull/586),[@Abhinandan-Purkait](https://github.com/Abhinandan-Purkait))

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

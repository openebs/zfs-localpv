/*
Copyright 2020 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
The restore flow is as follows:
- plugin creates a restore storage volume(zvol or dataset)

At the backup time, the plugin backs up the ZFSVolume CR and at while doing the restore we have all the information related to that volume. The plugin first creates the restore destination to store the data.

- plugin then creates the ZFSRestore CR with the destination volume and remote location from where the data needs to be read

- restore controller (on node) keeps a watch for new CRs associated with the node id. This node ID will be same as the Node ID present in the ZFSVolume resource.

- if Restore status == init and not marked for deletion, Restore controller will execute the `remote-read | zfs recv` command.


Limitation with the Initial Version :-

- The destination cluster should have same node ID and Zpool present.

- If volume was thick provisioned, then destination Zpool should have enough space for that volume.

- destination volume should be present before starting the Restore Operation.

- If the restore fails due to network issues and
        * the status update succeed, the Restore will not be re-attempted.
	* the status update fails, the Restore will be re-attempted from the beginning (TODO optimize it).

- If the restore doesn't have the specified backup, the plugin itself fails that restore request as there is no Backup to Restore from.

- If the same volume is restored twice, the data will be written again. The plugin should fail this kind of request.

*/

package restore

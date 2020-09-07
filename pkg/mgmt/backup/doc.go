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
The Backup flow is as follows:

- plugin takes the backup of ZFSVolume CR so that it can be restored.

- It will save the namespace information where the pvc is created also while taking the backup. Plugin will use this info if restoring without a namespace mapping to find if volume has already been restored.

- plugin then creates the ZFSBackup CR with the destination volume and remote location where the data needs to be send.

- Backup controller (on node) keeps a watch for new CRs associated with the node id. This node ID will be same as the Node ID present in the ZFSVolume resource.

- The Backup controller will take a snapshot and then send the data to the remote location.

- if Backup status == init and not marked for deletion, Backup controller will execute the `zfs send | remote-write` command.

- If Backup is deleted then corresponsing snapshot also gets deleted.


Limitation :-

- there should be enough space in the pool to accomodate the snapshot.

- if there is a network error and backup failed and :
    * Backup status update also failed, then backup will be retried from the beginning (TODO optimize it)
    * Backup status update is successful, the Backup operation will fail.

- A snapshot will exist as long as Backup is be present and it will be cleaned up when the Backup is deleted.

*/

package backup

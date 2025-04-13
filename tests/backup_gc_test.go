/*
Copyright © 2025 The OpenEBS Authors

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

package tests

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	zfsapi "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[zfspv] TEST BACKUP GARBAGE COLLECTOR", func() {

    Context("ZFS backups with broken references or chain dependencies exist", func() {

        It("Running orphaned backups test",Label("backup-gc"), orphanedBackupTest)
        It("Running chain dependency test", Label("backup-gc"), chainDependencyTest)
    })
})

// orphanedBackupTest checks orphaned backups are removed
func orphanedBackupTest() {
    By("####### Creating the default storage class #######")
    createFstypeStorageClass(nil)

    By("creating and verifying PVC bound status", func() {createAndVerifyPVC(pvcNameFS)})

    By("creating a regular backup with no previous snapshot reference")
    regularBackup := createBackup(
        "regular-backup",
        "snap1",
        "",
    )

    By("creating an orphaned backup that refers to non-existent previous snapshot")
    orphanedBackup := createBackup(
        "orphaned-backup",
        "snap2",
        "non-existent-snap",
    )

    By("creating a dependent backup that references the regular backup")
    dependentBackup := createBackup(
        "dependent-backup",
        "snap3",
        "snap1",
    )

    By("verifying orphaned backup is deleted")
    waitForBackupDeletion(orphanedBackup.Name)

    By("verifying regular backup still exists")
    Expect(backupExists(regularBackup.Name)).To(BeTrue())

    By("verifying dependent backup still exists")
    Expect(backupExists(dependentBackup.Name)).To(BeTrue())

    By("cleaning up backups and PVC")
    deleteBackup(regularBackup.Name)
    deleteBackup(dependentBackup.Name)
    deletePVC(pvcNameFS)
}

// chainDependencyTest checks dependent backups are cascaded on broken chain
func chainDependencyTest() {
    By("####### Creating the default storage class #######")
    createFstypeStorageClass(nil)

    By("creating and verifying PVC bound status")
    createAndVerifyPVC(pvcNameFS)

    By("creating a root backup with no previous snapshot")
    rootBackup := createBackup(
        "root-backup",
        "snap1",
        "",
    )

    By("creating a child backup referencing root backup")
    childBackup := createBackup(
        "child-backup",
        "snap2",
        "snap1",
    )

    By("creating a grandchild backup referencing the child backup")
    grandChildBackup := createBackup(
        "grandchild-backup",
        "snap3",
        "snap2",
    )

    By("deleting the root backup to break the chain")
    deleteBackup(rootBackup.Name)

    By("verifying child backup is deleted due to broken chain")
    waitForBackupDeletion(childBackup.Name)

    By("verifying grandchild backup is also deleted")
    waitForBackupDeletion(grandChildBackup.Name)

    By("cleaning up PVC")
    deletePVC(pvcNameFS)
}

func createBackup(name, snapName, prevSnapName string) *zfsapi.ZFSBackup {
	By("creating backup " + name)
    bkp := &zfsapi.ZFSBackup{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: OpenEBSNamespace,
        },
        Spec: zfsapi.ZFSBackupSpec{
            VolumeName:   backupVolumeName,
            OwnerNodeID:  backupNodeID,
            SnapName:     snapName,
            PrevSnapName: prevSnapName,
            BackupDest:   "10.10.10.10:1234",
        },
        Status: zfsapi.BKPZFSStatusDone,
    }

    createdBkp, err := BKPClient.WithNamespace(OpenEBSNamespace).Create(bkp)
    Expect(err).To(BeNil(), "Failed to create ZFSBackup CR")
    return createdBkp
}

func deleteBackup(name string) {
	By("deleting backup " + name)
    err := BKPClient.WithNamespace(OpenEBSNamespace).Delete(name)
    Expect(err).To(BeNil(), "Failed to delete ZFSBackup CR")
}

func backupExists(name string) bool {
	By("checking if backup " + name + " exists")
    _, err := BKPClient.WithNamespace(OpenEBSNamespace).Get(name, metav1.GetOptions{})
    return err == nil
}

func waitForBackupDeletion(name string) {
	By("waiting for backup " + name + " to be deleted")
    Eventually(func() bool {
        return !backupExists(name)
    }, 2*time.Second, 1*time.Second).Should(BeTrue(), fmt.Sprintf("Backup %s should be deleted", name))
}

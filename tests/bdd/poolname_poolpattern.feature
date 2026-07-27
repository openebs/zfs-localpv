Feature: Validate pool pattern (poolpattern) based volume provisioning
         Validate thick fit filter and capacity-aware fail-fast
         Validate snapshot, restore and clone under a poolpattern storage class
         Validate the VolumeWeighted, CapacityWeighted and SpaceWeighted scheduling algorithms

  Background:
    Given a single zfs cluster is configured
    And the node advertises its pools through the ZFSNode resource
    And pools zfspv-pool-a and zfspv-pool-b are created on the node
    And zfspv-pool-a has more free space than zfspv-pool-b
    And every storage class is created with volumeBindingMode as Immediate unless stated otherwise

  #########################################################################################################
  ################## Provisioning with poolpattern ########################################################
  #########################################################################################################

  Scenario : Volume provision using a matching poolpattern
    Given a storage class is created with poolpattern as "zfspv-pool.*"
    When pvc is created referencing this same storage class
    Then pvc should be in bound state
    And zfsvolume must be created and should be in Ready state to use by any application
    And the zfsvolume PoolName must be zfspv-pool-a, the matching pool with the largest free space

  Scenario: Fixed poolname with a pool dataset path
    Given a storage class is created with poolname as zfspv-pool-a/k8s/localpv
    When pvc is created referencing this same storage class
    Then pvc should be in bound state
    And the zfsvolume PoolName must be exactly zfspv-pool-a/k8s/localpv
    And the volume must be created under the dataset zfspv-pool-a/k8s/localpv on the node

  #########################################################################################################
  ################## poolname and poolpattern combinations ################################################
  #########################################################################################################

  Scenario Outline: A storage class must set exactly one of poolname and poolpattern
    Given a storage class is created with poolname as <poolname> and poolpattern as <poolpattern>
    When pvc is created referencing this same storage class
    Then the pvc must remain in pending state
    And the zfsvolume must not be created
    And a ProvisioningFailed event on the pvc must report InvalidArgument
    Examples:
      | poolname      | poolpattern    |
      | zfspv-pool-a  | "zfspv-pool.*" |
      | not specified | not specified  |

  #########################################################################################################
  ################## Thick fit filter and capacity-aware fail-fast ########################################
  #########################################################################################################

  Scenario Outline: Reserving volume must fit in a single matching pool
    Given a storage class is created with poolpattern as "zfspv-pool.*" and fsType as <fsType> and thinprovision as <thinprovision>
    And every matching pool on the node has less free space than the requested volume size
    When pvc is created referencing this same storage class
    Then the pvc must remain in pending state
    And the zfsvolume must not be created
    And a ProvisioningFailed event on the pvc must report ResourceExhausted
    Examples:
      | fsType | thinprovision |
      |  ext4  |               |
      |  zfs   |    no         |

  Scenario Outline: Non reserving volume bypasses the fit filter
    Given a storage class is created with poolpattern as "zfspv-pool.*" and fsType as <fsType> and thinprovision as <thinprovision> and quotatype as <quotatype>
    And every matching pool on the node has less free space than the requested volume size
    When pvc is created referencing this same storage class
    Then pvc should be in bound state
    And zfsvolume must be created and should be in Ready state to use by any application
    Examples:
      | fsType | thinprovision | quotatype |
      |  ext4  |    yes        |           |
      |  zfs   |    yes        |           |
      |  zfs   |               |           |
      |  zfs   |               | refquota  |

  Scenario: A reserving volume is placed on the matching pool that has enough free space
    Given a storage class is created with poolpattern as "zfspv-pool.*" and thinprovision as no
    And one matching pool on the node has enough free space and the other does not
    When pvc is created referencing this same storage class
    Then pvc should be in bound state
    And the zfsvolume PoolName must be the matching pool that has enough free space

  Scenario: A pending reserving volume binds once a matching pool has enough free space
    Given a storage class is created with poolpattern as "zfspv-pool.*" and thinprovision as no
    And every matching pool on the node has less free space than the requested volume size
    When pvc is created referencing this same storage class
    Then the pvc must remain in pending state
    And the zfsvolume must not be created
    When free space is made available on a matching pool
    Then pvc should be in bound state
    And the zfsvolume must be created in the pool that now has enough free space

  #########################################################################################################
  ################## Snapshot, restore, clone and resize under a poolpattern storage class ########################
  #########################################################################################################

  Scenario : Create a snapshot of a pattern provisioned volume
    Given a storage class is created with poolpattern as "zfspv-pool.*"
    And pvc is created referencing this storage class and a deployment using the same pvc
    And a zfsvolume is created and used by the application
    When the snapshot create command for the given pvc is run in the same namespace
    Then the snapshot must be created for the corresponding pvc with status as ready to use

  Scenario Outline: Create a clone from a <source> under a poolpattern storage class
    Given a storage class is created with poolpattern as "zfspv-pool.*" and fsType as <fsType>
    And pvc is created referencing this storage class and a deployment using the same pvc
    And a zfsvolume is created and used by the application
    And a snapshot is created for the corresponding pvc when the <source> is a snapshot
    When the clone create command is run for this <source> and a deployment is created to use this clone
    Then the clone must be created from the <source> and the created deployment must use the cloned volume
    And the clone PoolName must be the same as the source volume pool
    Examples:
      | fsType | source    |
      |  zfs   | snapshot  |
      |  ext4  | snapshot  |
      |  zfs   | volume    |
      |  ext4  | volume    |

  Scenario: Clone inherits the source pool even when another matching pool has more free space
    Given a source volume exists in zfspv-pool-b which matches the pattern
    And zfspv-pool-a also matches the pattern and has more free space than zfspv-pool-b
    And a storage class is created with poolpattern as "zfspv-pool.*"
    When a clone is created from that source referencing this storage class
    Then the clone must be provisioned in zfspv-pool-b, the same pool as the source

  Scenario: Raw block volume provisioning with poolpattern
    Given a storage class is created with poolpattern as "zfspv-pool.*" and without any fstype
    And a pvc with volumeMode as Block and a size that fits in a matching pool is created referencing this storage class
    When a deployment is created using the same pvc
    Then zfsvolume must be created and used by the application
    And the zfsvolume PoolName must be one of the pools matching the pattern

  Scenario Outline: Resize a pattern provisioned volume
    Given a storage class is created with poolpattern as "zfspv-pool.*" and allowVolumeExpansion and fsType as <fsType>
    And pvc is created referencing this storage class and a deployment using the same pvc
    When the pvc size is increased
    Then the zfsvolume capacity must be updated in the pool that the pattern resolved to
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |

############################################################################################################
################## The below bdd need to be implemented in e2e tests (multi node) ##########################
############################################################################################################

  Scenario: Provisioning across nodes with differently named pools
    Given a multi node cluster where each node carries a differently named pool matching the pattern
    And a storage class is created with poolpattern as "zfspv-pool.*"
    When several pvcs are created referencing this same storage class
    Then each pvc must be bound and provisioned on whichever node carries a matching pool

  Scenario: CapacityWeighted scheduling orders nodes by real pool usage
    Given a multi node cluster with a matching pool on each node
    And a storage class is created with poolpattern as "zfspv-pool.*" and scheduler as CapacityWeighted
    And the matching pool on one node has less used capacity than the matching pool on another node
    And that pool holds data that was not provisioned through the driver
    When pvc is created referencing this same storage class
    Then the zfsvolume must be provisioned on the node whose matching pool has the least used capacity

  Scenario: VolumeWeighted scheduling orders nodes by volume count
    Given a multi node cluster with a matching pool on each node
    And a storage class is created with poolpattern as "zfspv-pool.*" and scheduler as VolumeWeighted
    And the matching pool on one node has fewer zfsvolumes than the matching pool on another node
    When pvc is created referencing this same storage class
    Then the zfsvolume must be provisioned on the node with the fewest matching zfsvolumes

  # SpaceWeighted orders by what a pool has left rather than by what has been written into
  # it, so the two nodes below are deliberately set up to be ranked the opposite way by
  # SpaceWeighted and by CapacityWeighted
  Scenario: SpaceWeighted scheduling orders nodes by free space
    Given a multi node cluster with a matching pool on each node
    And a storage class is created with poolpattern as "zfspv-pool.*" and scheduler as SpaceWeighted
    And the matching pool on one node has more free space but also more used capacity than the matching pool on another node
    When pvc is created referencing this same storage class
    Then the zfsvolume must be provisioned on the node whose matching pool has the most free space
    And not on the node that CapacityWeighted would have picked for its lower used capacity

  Scenario: CapacityWeighted stays the default when no scheduler is set
    Given a multi node cluster with a matching pool on each node
    And a storage class is created with poolpattern as "zfspv-pool.*" and without any scheduler
    And the matching pool on one node has more free space but also more used capacity than the matching pool on another node
    When pvc is created referencing this same storage class
    Then the zfsvolume must be provisioned on the node whose matching pool has the least used capacity
    And not on the node that SpaceWeighted would have picked for its larger free space

  # a thin volume is used on purpose: it skips the fit filter, so the node ordering alone
  # decides where it lands. A node missing from the weight map is not excluded by the
  # scheduler but moved to the front of the list, so a full pool that is dropped instead of
  # being weighted last would be preferred over a pool with room
  Scenario: SpaceWeighted orders a node whose matching pool is full last, not first
    Given a multi node cluster with a matching pool on each node
    And the matching pool on one node is full and the matching pool on another node has free space
    And a storage class is created with poolpattern as "zfspv-pool.*" and scheduler as SpaceWeighted and thinprovision as yes
    When pvc is created referencing this same storage class
    Then the zfsvolume must be provisioned on the node whose matching pool has free space
    And the node whose matching pool is full must not be preferred over itdoe

  Scenario: Capacity tracking reports capacity for a poolpattern storage class
    Given the storage capacity feature is enabled on the CSIDriver
    And a storage class is created with poolpattern as "zfspv-pool.*" and volumeBindingMode as WaitForFirstConsumer
    When the external provisioner queries GetCapacity for the storage class
    Then a CSIStorageCapacity object must report the maximum free space across the matching pools per topology segment
    And the value must not be zero

  Scenario: Pod is scheduled only onto a node that has capacity for a poolpattern storage class
    Given a multi node cluster and a storage class is created with poolpattern as "zfspv-pool.*" and volumeBindingMode as WaitForFirstConsumer
    And only one node has a matching pool with enough free space
    When a pod that consumes the pvc is created
    Then the pod and its volume must be scheduled onto the node whose matching pool has enough free space

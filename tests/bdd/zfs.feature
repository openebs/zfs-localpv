Feature: Validate volume provisioning for fsType zfs, ext4, xfs, btrfs
         Validate snapshot creation and clone
         Validate volume resize

  Background:
    Given a single zfs cluster is configured
    And zfs pool pool created on the node

  Scenario Outline: Volume provision for file system
    Given a storage class is created with fsType as <fsType>
    When pvc is created referencing this same storage class
    Then pvc should be in bound state
    And zfsvolume must be created and should be in Ready state to use by any application
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |
      |  xfs   |
      |  btrfs |

  Scenario Outline: Volume provision for file system
    Given a storage class is created with fsType as <fsType>
    And pvc is created referencing this same storage class with zfsvolume in Ready state to use by any application
    When a deployment is created using the same pvc
    Then deployment pod should be in Running state
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |
      |  xfs   |
      |  btrfs |

  Scenario Outline: Volume property change 
    Given a storage class is created with fsType as <fsType>
    And pvc is created referencing this storage class and a deployment using the same pvc
    When the zfsvolume properties like <properties> are updated
    Then the zfsvolume properties must reflect the updated value
    Examples:
      | fsType | properties     |
      |  zfs   | compression    |
      |  ext4  | compression    |
      |  xfs   | compression    |
      |  btrfs | compression    |
      |  zfs   | dedup          |
      |  ext4  | dedup          |
      |  xfs   | dedup          |
      |  btrfs | dedup          |


  Scenario Outline: Volume data size change
    Given a storage class is created with fsType as <fsType>
    And pvc is created referencing this storage class and a deployment using the same pvc
    When the zfsvolume volume type is <volume_type> and record size is updated to <record_size> and volume block size to <block_size> size is updated
    Then the volume type is <volume_type> must have the record sise as  <record_size> and volume block size as <block_size> size
    Examples:
      | fsType | volume_type | record_size   | block_size |
      |  zfs   |  DATASET    | 4096          | 8192       |
      |  ext4  |  DATASET    | 4096          | 8192       |
      |  xfs   |  DATASET    | 4096          | 8192       |
      |  btrfs |  DATASET    | 4096          | 8192       |
      |  zfs   |  ZVOL       | 8192          | 16384      |
      |  ext4  |  ZVOL       | 8192          | 16384      |
      |  xfs   |  ZVOL       | 8192          | 16384      |
      |  btrfs |  ZVOL       | 8192          | 16384      |

  Scenario Outline: Create a snapshot from a pvc
    Given a storage class is created with fsType as <fsType>
    And pvc is created referencing this storage class and a deployment using the same pvc
    And a zfsvolume is created and used by the application
    When the snapshot create command for the given pvc is run in the same namespace
    Then the snapshot must be created for the corresponding pvc with status as ready to use
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |
      |  xfs   |
      |  btrfs |
    
  Scenario Outline: Create a clone from a <source>
    Given a storage class is created with fsType as <fsType>
    And pvc is created referencing this storage class and a deployment using the same pvc
    And a zfsvolume is created and used by the application
    And a snapshot is be created for the corresponding pvc 
    When the clone create command is run for this snapshot and a deployment is created to use this clone
    Then the clone must be created from the <source> and the created deployment must use the cloned volume
    Examples:
      | fsType | source    |
      |  zfs   | snapshot  |
      |  ext4  | snapshot  |
      |  xfs   | snapshot  |
      |  btrfs | snapshot  |
      |  zfs   | volume    |
      |  ext4  | volume    |
      |  xfs   | volume    |
      |  btrfs | volume    |

  Scenario Outline: Resize of pvc
    Given a storage class is created with fsType as <fsType>
    And pvc is created referencing the this storage class and a deployment using the same pvc
    When the size of pvc is update to a <new_capacity>
    Then the pvc must be of size <new_capacity> 
    Examples:
      | fsType | new_capacity |
      |  zfs   |  8Gi         |
      |  ext4  |  8Gi         |
      |  xfs   |  8Gi         |


  Scenario: Volume provision for a raw block volume
    Given a storage class is created without any fstype
    And a pvc with volumeMode as Bolck is created referencing this storage class
    When a deployment is created using the same pvc
    Then zfsvolume must be created and used by the application


  Scenario Outline: Volume data size change for a raw block volume
    Given a storage class is created without any fstype
    And a pvc with volumeMode as Bolck is created referencing this storage class
    When the volume type is <volume_type> and record size is updated to <record_size> and volume block size to <block_size> size
    Then the volume type is <volume_type> must have the record sise as  <record_size> and volume block size as <block_size> size
    Examples:
      | volume_type | record_size   | block_size |
      |  DATASET    | 4096          | 8192       |
      |  DATASET    | 4096          | 8192       |
      |  DATASET    | 4096          | 8192       |
      |  DATASET    | 4096          | 8192       |
      |  ZVOL       | 8192          | 16384      |
      |  ZVOL       | 8192          | 16384      |
      |  ZVOL       | 8192          | 16384      |
      |  ZVOL       | 8192          | 16384      |

  Scenario Outline: Online volume resize
    Given a storage class is created with fsType as <fsType>
    And pvc is created referencing this storage class
    And a deployment is created using the same pvc which is in running state
    When resize volume requested is by updating the PVC resource
    Then corresponding pvc must be resized
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |
      |  xfs   |

  Scenario Outline: Thin provisioning
    Given a storage class is created with fsType as <fsType> and thinprovision as true
    And pvc is created referencing this storage class and a deployment using the same pvc
    And a zfsvolume is created and used by the application
    When the ZPOOL size is 1GB and the requested storage in PVC is 10 GB
    Then the volume must be provisioned even if the ZPOOL does not have the enough capacity
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |
      |  xfs   |
      |  btrfs |

   Scenario Outline: Supported compression algorithm
    Given a storage class is created with fsType as zfs, ext4, xfs, btrfs and compression as <compression_type>
    And pvc is created referencing this storage class and a deployment using the same pvc
    When the zfsvolume compression is present 
    Then the volume ust be provisioned and the compression type of volume must be <compression_type>
    Examples:
      | compression_type|
      |     lzjb        |  
      |     zstd        |
      |     zstd-1      |
      |     zstd-2      |
      |     zstd-3      |
      |     zstd-4      |
      |     zstd-5      |
      |     zstd-7      |
      |     zstd-8      |
      |     zstd-9      |
      |     zstd-10     |
      |     zstd-11     |
      |     zstd-12     |
      |     zstd-13     |
      |     zstd-14     |
      |     zstd-15     |
      |     zstd-16     |
      |     zstd-17     |
      |     zstd-18     |
      |     zstd-19     |
      |     gzip        | 
      |     gzip-1      |
      |     gzip-2      |
      |     gzip-3      |
      |     gzip-4      |
      |     gzip-5      |
      |     gzip-6      |
      |     gzip-7      |
      |     gzip-8      |
      |     gzip-9      |
      |     zle         |
      |     lz4         | 
    
############################################################################################################
################## The below bdd need to be implemented in ci tests ########################################
############################################################################################################

  Scenario Outline: Create a clone from a <source>
    Given a storage class is created without any fstype as Raw block volume does not have any fstype
    And a pvc with volumeMode as Bolck is created referencing this storage class
    And a deployment is created using the same pvc
    And a snapshot is be created for the corresponding <source>
    When the clone create command is run for this snapshot and a deployment is created to use this clone
    Then the clone must be created from the snapshot and the created deployment must use the cloned volume
    Examples:
      | source    |
      | snapshot  |
      | volume    |

  Scenario Outline: Shared volume
    Given a storage class is created with fsType as <fsType> and shared is set to yes
    And pvc is created referencing this storage class and a deployment using the same pvc
    And a zfsvolume is created and used by the application
    When two pods were deployed on same node and they both are running
    Then the LocalPV-ZFS Driver will allow the volumes to be mounted by more than one pods
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |
      |  xfs   |
      |  btrfs |


############################################################################################################
################## The below bdd need to be implemented in e2e tests ########################################
############################################################################################################

  Scenario Outline: Allowed topologies 
    Given a storage class with allowedTopologies is created with fsType as <fsType>
    And pvc is created referencing this storage class and a deployment using the same pvc
    When the topology for specific zpool is defined in the storage class
    Then the volumes must be provisioned on the nodes which has the required zpool
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |
      |  xfs   |
      |  btrfs |

  Scenario Outline: Import Existing Volumes to LocalPV-ZFS
    Given a storage class is created with fsType as <fsType>
    And pvc is created referencing this storage class and a deployment using the same pvc
    When a node fails and the disks is attached to a different node
    Then the zfs pool should be imported on that pool
    And the zfsvolume resource should reflect that upon creation of a volume referencing the pool on that node
    And the application should keep running
    Examples:
      | fsType |
      |  zfs   |
      |  ext4  |
      |  xfs   |
      |  btrfs |

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
    And pvc is created referencing this storage class with a deployment using the same pvc
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
    And pvc is created referencing this storage class with a deployment using the same pvc
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
    And pvc is created referencing this storage class with a deployment using the same pvc
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
    And pvc is created referencing this storage class with a deployment using the same pvc
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
    And pvc is created referencing the this storage class with a deployment using the same pvc
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

  Scenario Outline: Volume property change for a raw block volume
    Given a storage class is created without any fstype
    And a pvc with volumeMode as Bolck is created referencing this storage class
    When a deployment is created using the same pvc
    When the zfsvolume properties <properties> are updated
    Then the zfsvolume properties must reflect the changed value for the <properties>
    Examples:
      | properties     |
      |  compression   |
      |  dedup         |

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

    
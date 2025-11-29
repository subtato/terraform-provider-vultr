---
layout: "vultr"
page_title: "Vultr: vultr_virtual_file_system_storage"
sidebar_current: "docs-vultr-resource-virtual-file-system-storage"
description: |-
  Provides a Vultr virtual file system storage resource. This can be used to create, read, modify and delete a virtual file system storage.
---

# vultr_virtual_file_system_storage

Provides a Vultr virtual file system storage resource. This can be used to create, read, modify and delete a virtual file system storage.

## Example Usage

Define a virtual file system storage resource:

```hcl
resource "vultr_virtual_file_system_storage" "my_vfs_storage" {
  label = "vultr-vfs-storage"
  size_gb = 10
  region = "ewr"
  tags = ["terraform", "important"]
}
```

Create a virtual file system storage with attached instances:

```hcl
resource "vultr_instance" "my_instance" {
  plan   = "vc2-1c-1gb"
  region = "ewr"
  os_id  = 387
}

resource "vultr_virtual_file_system_storage" "my_vfs_storage" {
  label             = "vultr-vfs-storage"
  size_gb           = 100
  region            = "ewr"
  disk_type         = "nvme"
  attached_instances = [vultr_instance.my_instance.id]
  tags              = ["production", "shared-storage"]
}
```

## Argument Reference

~> Updating `tags` will cause a `force new`.

The following arguments are supported:

* `size_gb` - (Required) The size of the given virtual file system storage subscription.
* `region` - (Required) The region in which this virtual file system storage will reside.
* `label` - (Required) The label to give to the virtual file system storage subscription.
* `tags` - (Optional) A list of tags to be used on the virtual file system storage subscription.
* `attached_instances` - (Optional) A list of UUIDs to attach to the virtual file system storage subscription.
* `disk_type` - (Optional) The underlying disk type to use for the virtual file system storage.  Default is `nvme`.

## Attributes Reference

The following attributes are exported:

* `attached_instances` - A list of instance IDs currently attached to the virtual file system storage.
* `attachments` - A list of attachment states for instances currently attached to the virtual file system storage. Each attachment contains:
  * `instance_id` - The ID of the attached instance.
  * `state` - The current state of the attachment (e.g., "ATTACHED").
  * `mount` - The mount tag number for this attachment.
* `charges` - The current pending charges for the virtual file system storage subscription in USD.
* `cost` - The cost per month of the virtual file system storage subscription in USD.
* `date_created` - The date the virtual file system storage subscription was added to your Vultr account.
* `disk_type` - The underlying disk type used by the virtual file system storage subscription.
* `label` - The label of the virtual file system storage subscription.
* `region` - The region ID of the virtual file system storage subscription.
* `status` - The status of the virtual file system storage subscription.
* `size_gb` - The size of the virtual file system storage subscription in GB.
* `tags` - A list of tags used on the virtual file system storage subscription.

## Timeouts

The `timeouts` block allows you to specify timeouts for certain operations:

* `create` - (Defaults to 30 minutes) Used when creating the virtual file system storage and attaching instances
* `update` - (Defaults to 30 minutes) Used when updating the virtual file system storage or changing attachments
* `delete` - (Defaults to 10 minutes) Used when destroying the virtual file system storage

## Import

Virtual file system storage can be imported using the `ID`, e.g.

```
terraform import vultr_virtual_file_system_storage.my_vfs_storage 79210a84-bc58-494f-8dd1-953685654f7f
```

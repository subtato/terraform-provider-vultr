---
layout: "vultr"
page_title: "Vultr: vultr_cdn_push_zone"
sidebar_current: "docs-vultr-resource-cdn-push-zone"
description: |-
  Provides a Vultr CDN Push Zone resource. This can be used to create, read, modify, and delete CDN Push Zones.
---

# vultr_cdn_push_zone

Provides a Vultr CDN Push Zone resource. This can be used to create, read, modify, and delete CDN Push Zones.

## Example Usage

Create a new CDN Push Zone:

```hcl
resource "vultr_cdn_push_zone" "my_cdn" {
  label = "my-cdn-push-zone"
}
```

## Argument Reference

The following arguments are supported:

* `label` - (Required) The label for the CDN push zone.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the CDN push zone.
* `label` - The label of the CDN push zone.
* `cdn_domain` - The CDN domain URL for accessing content through the CDN.
* `date_created` - The date the CDN push zone was created.

## Import

CDN Push Zones can be imported using the CDN Push Zone `ID`, e.g.

```
terraform import vultr_cdn_push_zone.my_cdn abc123
```


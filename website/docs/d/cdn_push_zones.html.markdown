---
layout: "vultr"
page_title: "Vultr: vultr_cdn_push_zones"
sidebar_current: "docs-vultr-datasource-cdn-push-zones"
description: |-
  Get information about your Vultr CDN Push Zones.
---

# vultr_cdn_push_zones

Get information about your Vultr CDN Push Zones. This data source provides a list of all CDN push zones associated with your Vultr account.

## Example Usage

Get all CDN Push Zones:

```hcl
data "vultr_cdn_push_zones" "my_zones" {}
```

Get filtered CDN Push Zones:

```hcl
data "vultr_cdn_push_zones" "my_zones" {
  filter {
    name   = "label"
    values = ["production", "staging"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) Query parameters for finding CDN push zones.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `cdn_push_zones` - A list of CDN push zones. Each zone contains:
  * `id` - The ID of the CDN push zone.
  * `label` - The label of the CDN push zone.
  * `cdn_domain` - The CDN domain URL for accessing content through the CDN.
  * `date_created` - The date the CDN push zone was created.


---
layout: "vultr"
page_title: "Vultr: vultr_cdn_pull_zones"
sidebar_current: "docs-vultr-datasource-cdn-pull-zones"
description: |-
  Get information about your Vultr CDN Pull Zones.
---

# vultr_cdn_pull_zones

Get information about your Vultr CDN Pull Zones. This data source provides a list of all CDN pull zones associated with your Vultr account.

## Example Usage

Get all CDN Pull Zones:

```hcl
data "vultr_cdn_pull_zones" "my_zones" {}
```

Get filtered CDN Pull Zones:

```hcl
data "vultr_cdn_pull_zones" "my_zones" {
  filter {
    name   = "label"
    values = ["production", "staging"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) Query parameters for finding CDN pull zones.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `cdn_pull_zones` - A list of CDN pull zones. Each zone contains:
  * `id` - The ID of the CDN pull zone.
  * `label` - The label of the CDN pull zone.
  * `origin_domain` - The origin domain for the CDN pull zone.
  * `cdn_domain` - The CDN domain URL for accessing content through the CDN.
  * `date_created` - The date the CDN pull zone was created.


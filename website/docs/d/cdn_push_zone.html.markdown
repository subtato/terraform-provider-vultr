---
layout: "vultr"
page_title: "Vultr: vultr_cdn_push_zone"
sidebar_current: "docs-vultr-datasource-cdn-push-zone"
description: |-
  Get information about a specific Vultr CDN Push Zone.
---

# vultr_cdn_push_zone

Get information about a specific Vultr CDN Push Zone. This data source provides details about a single CDN push zone by filtering the push zone list.

## Example Usage

Get a specific CDN Push Zone by label:

```hcl
data "vultr_cdn_push_zone" "my_cdn" {
  filter {
    name   = "label"
    values = ["my-cdn-push-zone"]
  }
}
```

Get a CDN Push Zone by ID:

```hcl
data "vultr_cdn_pull_zone" "my_cdn" {
  filter {
    name   = "id"
    values = ["abc123"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Required) Query parameters for finding the CDN push zone. Must return exactly one result.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the CDN push zone.
* `label` - The label of the CDN push zone.
* `cdn_domain` - The CDN domain URL for accessing content through the CDN.
* `date_created` - The date the CDN push zone was created.


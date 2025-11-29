---
layout: "vultr"
page_title: "Vultr: vultr_cdn_pull_zone"
sidebar_current: "docs-vultr-datasource-cdn-pull-zone"
description: |-
  Get information about a specific Vultr CDN Pull Zone.
---

# vultr_cdn_pull_zone

Get information about a specific Vultr CDN Pull Zone. This data source provides details about a single CDN pull zone by filtering the pull zone list.

## Example Usage

Get a specific CDN Pull Zone by label:

```hcl
data "vultr_cdn_pull_zone" "my_cdn" {
  filter {
    name   = "label"
    values = ["my-cdn-zone"]
  }
}
```

Get a CDN Pull Zone by ID:

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

* `filter` - (Required) Query parameters for finding the CDN pull zone. Must return exactly one result.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the CDN pull zone.
* `label` - The label of the CDN pull zone.
* `origin_domain` - The origin domain for the CDN pull zone.
* `cdn_domain` - The CDN domain URL for accessing content through the CDN.
* `date_created` - The date the CDN pull zone was created.


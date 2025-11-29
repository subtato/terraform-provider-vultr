---
layout: "vultr"
page_title: "Vultr: vultr_cdn_pull_zone"
sidebar_current: "docs-vultr-resource-cdn-pull-zone"
description: |-
  Provides a Vultr CDN Pull Zone resource. This can be used to create, read, modify, and delete CDN Pull Zones.
---

# vultr_cdn_pull_zone

Provides a Vultr CDN Pull Zone resource. This can be used to create, read, modify, and delete CDN Pull Zones.

## Example Usage

Create a new CDN Pull Zone:

```hcl
resource "vultr_cdn_pull_zone" "my_cdn" {
  label         = "my-cdn-zone"
  origin_domain = "example.com"
}
```

## Argument Reference

The following arguments are supported:

* `label` - (Required) The label for the CDN pull zone.
* `origin_domain` - (Required) The origin domain that the CDN will pull content from.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the CDN pull zone.
* `label` - The label of the CDN pull zone.
* `origin_domain` - The origin domain for the CDN pull zone.
* `cdn_domain` - The CDN domain URL for accessing content through the CDN.
* `date_created` - The date the CDN pull zone was created.

## Import

CDN Pull Zones can be imported using the CDN Pull Zone `ID`, e.g.

```
terraform import vultr_cdn_pull_zone.my_cdn abc123
```


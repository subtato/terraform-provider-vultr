---
layout: "vultr"
page_title: "Vultr: vultr_pending_charges"
sidebar_current: "docs-vultr-datasource-pending-charges"
description: |-
  Get information about your Vultr pending charges.
---

# vultr_pending_charges

Get information about your Vultr pending charges. This data source provides the total amount of pending charges on your account.

## Example Usage

Get pending charges:

```hcl
data "vultr_pending_charges" "my_charges" {}

output "pending_charges" {
  value = data.vultr_pending_charges.my_charges.pending_charges
}
```

## Argument Reference

This data source does not take any arguments. It will return the total pending charges associated with the Vultr API key you have set.

## Attributes Reference

The following attributes are exported:

* `pending_charges` - The total amount of pending charges on your Vultr account in USD.


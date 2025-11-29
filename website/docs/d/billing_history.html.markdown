---
layout: "vultr"
page_title: "Vultr: vultr_billing_history"
sidebar_current: "docs-vultr-datasource-billing-history"
description: |-
  Get information about your Vultr billing history.
---

# vultr_billing_history

Get information about your Vultr billing history. This data source provides a list of all billing history entries associated with your Vultr account.

## Example Usage

Get all billing history entries:

```hcl
data "vultr_billing_history" "my_history" {}
```

Get filtered billing history entries:

```hcl
data "vultr_billing_history" "my_history" {
  filter {
    name   = "type"
    values = ["charge"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) Query parameters for finding billing history entries.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `billing_history` - A list of billing history entries. Each entry contains:
  * `id` - The ID of the billing history entry.
  * `date` - The date of the billing history entry.
  * `type` - The type of billing entry (e.g., "charge", "payment").
  * `description` - A description of the billing entry.
  * `amount` - The amount of the billing entry in USD.
  * `balance` - The account balance after this entry in USD.


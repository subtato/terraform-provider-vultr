---
layout: "vultr"
page_title: "Vultr: vultr_invoices"
sidebar_current: "docs-vultr-datasource-invoices"
description: |-
  Get information about your Vultr invoices.
---

# vultr_invoices

Get information about your Vultr invoices. This data source provides a list of all invoices associated with your Vultr account.

## Example Usage

Get all invoices:

```hcl
data "vultr_invoices" "my_invoices" {}
```

Get filtered invoices:

```hcl
data "vultr_invoices" "my_invoices" {
  filter {
    name   = "description"
    values = ["Monthly Invoice"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) Query parameters for finding invoices.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `invoices` - A list of invoices. Each invoice contains:
  * `id` - The ID of the invoice.
  * `date` - The date of the invoice.
  * `description` - A description of the invoice.
  * `amount` - The amount of the invoice in USD.
  * `balance` - The remaining balance on the invoice in USD.


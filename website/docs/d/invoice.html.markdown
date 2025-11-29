---
layout: "vultr"
page_title: "Vultr: vultr_invoice"
sidebar_current: "docs-vultr-datasource-invoice"
description: |-
  Get information about a specific Vultr invoice.
---

# vultr_invoice

Get information about a specific Vultr invoice. This data source provides details about a single invoice by filtering the invoice list.

## Example Usage

Get a specific invoice by ID:

```hcl
data "vultr_invoice" "my_invoice" {
  filter {
    name   = "id"
    values = ["12345"]
  }
}
```

Get an invoice by description:

```hcl
data "vultr_invoice" "my_invoice" {
  filter {
    name   = "description"
    values = ["Monthly Invoice - January 2024"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Required) Query parameters for finding the invoice. Must return exactly one result.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the invoice.
* `date` - The date of the invoice.
* `description` - A description of the invoice.
* `amount` - The amount of the invoice in USD.
* `balance` - The remaining balance on the invoice in USD.


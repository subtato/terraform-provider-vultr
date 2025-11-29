---
layout: "vultr"
page_title: "Vultr: vultr_invoice_items"
sidebar_current: "docs-vultr-datasource-invoice-items"
description: |-
  Get information about items on a specific Vultr invoice.
---

# vultr_invoice_items

Get information about items on a specific Vultr invoice. This data source provides a detailed list of all line items for a given invoice.

## Example Usage

Get all items for a specific invoice:

```hcl
data "vultr_invoice_items" "my_items" {
  invoice_id = 12345
}
```

Get filtered invoice items:

```hcl
data "vultr_invoice_items" "my_items" {
  invoice_id = 12345
  filter {
    name   = "product"
    values = ["Compute Instance"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `invoice_id` - (Required) The ID of the invoice to get items for.
* `filter` - (Optional) Query parameters for finding invoice items.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `invoice_items` - A list of invoice items. Each item contains:
  * `description` - A description of the invoice item.
  * `product` - The product name for this item.
  * `start_date` - The start date for this item's billing period.
  * `end_date` - The end date for this item's billing period.
  * `units` - The number of units for this item.
  * `unit_type` - The type of unit (e.g., "hours", "days").
  * `unit_price` - The price per unit in USD.
  * `amount` - The total amount for this item in USD.


---
layout: "vultr"
page_title: "Vultr: vultr_logs"
sidebar_current: "docs-vultr-datasource-logs"
description: |-
  Get information about your Vultr logs.
---

# vultr_logs

Get information about your Vultr logs. This data source provides a list of available logs.

~> **Note:** This data source is currently a placeholder. The Log service is not yet implemented in the govultr SDK. This will be fully functional once the SDK adds support for the Logs API endpoint.

## Example Usage

Get all logs:

```hcl
data "vultr_logs" "my_logs" {}
```

Get filtered logs:

```hcl
data "vultr_logs" "my_logs" {
  filter {
    name   = "name"
    values = ["application"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) Query parameters for finding logs.

The `filter` block supports:

* `name` - Attribute name to filter on.
* `values` - One or more values to filter on.

## Attributes Reference

The following attributes are exported:

* `logs` - A list of logs. Each log contains:
  * `id` - The ID of the log.
  * `name` - The name of the log.
  * `date_created` - The date the log was created.


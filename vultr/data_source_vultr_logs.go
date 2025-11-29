package vultr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVultrLogs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrLogsRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"date_created": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVultrLogsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Note: Log service is not yet implemented in the govultr SDK
	// This data source is a placeholder for when the SDK adds support
	_ = meta
	_ = ctx

	return diag.Errorf("Log service is not yet implemented in the govultr SDK. Please check the SDK documentation for updates.")
}

package vultr

import (
	"context"
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vultr/govultr/v3"
)

func dataSourceVultrPendingCharges() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrPendingChargesRead,
		Schema: map[string]*schema.Schema{
			"pending_charges": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
	}
}

func dataSourceVultrPendingChargesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	options := &govultr.ListOptions{}
	pendingCharges, _, err := client.Billing.ListPendingCharges(ctx, options)
	if err != nil {
		return diag.Errorf("error getting pending charges: %v", err)
	}

	// Calculate total from invoice items
	var total float64
	for _, item := range pendingCharges {
		total += float64(item.Total)
	}

	d.SetId("pending_charges")
	if err := d.Set("pending_charges", math.Round(total*100)/100); err != nil {
		return diag.Errorf("unable to set `pending_charges` read value: %v", err)
	}

	return nil
}

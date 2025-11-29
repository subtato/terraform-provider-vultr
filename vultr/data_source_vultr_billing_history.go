package vultr

import (
	"context"
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vultr/govultr/v3"
)

func dataSourceVultrBillingHistory() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrBillingHistoryRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"billing_history": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"amount": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"balance": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVultrBillingHistoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	var billingHistoryList []map[string]interface{}
	filters, filtersOk := d.GetOk("filter")
	f := buildVultrDataSourceFilter(filters.(*schema.Set))
	options := &govultr.ListOptions{}

	for {
		billingHistory, meta, _, err := client.Billing.ListHistory(ctx, options)
		if err != nil {
			return diag.Errorf("error getting billing history: %v", err)
		}

		for _, item := range billingHistory {
			sm, err := structToMap(item)
			if err != nil {
				return diag.FromErr(err)
			}

			// If filters exist, check if this item matches
			if filtersOk && !filterLoop(f, sm) {
				continue
			}

			billingHistoryList = append(billingHistoryList, map[string]interface{}{
				"id":          item.ID,
				"date":        item.Date,
				"type":        item.Type,
				"description": item.Description,
				"amount":      math.Round(float64(item.Amount)*100) / 100,
				"balance":     math.Round(float64(item.Balance)*100) / 100,
			})
		}

		if meta.Links.Next == "" {
			break
		} else {
			options.Cursor = meta.Links.Next
			continue
		}
	}

	d.SetId("billing_history")
	if err := d.Set("billing_history", billingHistoryList); err != nil {
		return diag.Errorf("unable to set `billing_history` read value: %v", err)
	}

	return nil
}


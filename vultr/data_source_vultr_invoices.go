package vultr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vultr/govultr/v3"
)

func dataSourceVultrInvoices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrInvoicesRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"invoices": {
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

func dataSourceVultrInvoicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	var invoiceList []map[string]interface{}
	filters, filtersOk := d.GetOk("filter")
	f := buildVultrDataSourceFilter(filters.(*schema.Set))
	options := &govultr.ListOptions{}

	for {
		invoices, meta, _, err := client.Billing.ListInvoices(ctx, options)
		if err != nil {
			return diag.Errorf("error getting invoices: %v", err)
		}

		for _, invoice := range invoices {
			sm, err := structToMap(invoice)
			if err != nil {
				return diag.FromErr(err)
			}

			// If filters exist, check if this invoice matches
			if filtersOk && !filterLoop(f, sm) {
				continue
			}

			invoiceList = append(invoiceList, map[string]interface{}{
				"id":          invoice.ID,
				"date":        invoice.Date,
				"description": invoice.Description,
				"amount":      invoice.Amount,
				"balance":     invoice.Balance,
			})
		}

		if meta.Links.Next == "" {
			break
		} else {
			options.Cursor = meta.Links.Next
			continue
		}
	}

	d.SetId("invoices")
	if err := d.Set("invoices", invoiceList); err != nil {
		return diag.Errorf("unable to set `invoices` read value: %v", err)
	}

	return nil
}


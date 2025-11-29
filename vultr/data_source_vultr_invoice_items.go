package vultr

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vultr/govultr/v3"
)

func dataSourceVultrInvoiceItems() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrInvoiceItemsRead,
		Schema: map[string]*schema.Schema{
			"invoice_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The invoice ID to get items for",
			},
			"filter": dataSourceFiltersSchema(),
			"invoice_items": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"product": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"units": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"unit_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"unit_price": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"amount": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVultrInvoiceItemsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	invoiceID := d.Get("invoice_id").(int)
	var invoiceItemList []map[string]interface{}
	filters, filtersOk := d.GetOk("filter")
	f := buildVultrDataSourceFilter(filters.(*schema.Set))
	options := &govultr.ListOptions{}

	for {
		items, meta, _, err := client.Billing.ListInvoiceItems(ctx, invoiceID, options)
		if err != nil {
			return diag.Errorf("error getting invoice items: %v", err)
		}

		for _, item := range items {
			sm, err := structToMap(item)
			if err != nil {
				return diag.FromErr(err)
			}

			// If filters exist, check if this item matches
			if filtersOk && !filterLoop(f, sm) {
				continue
			}

			invoiceItemList = append(invoiceItemList, map[string]interface{}{
				"description": item.Description,
				"product":     item.Product,
				"start_date":  item.StartDate,
				"end_date":    item.EndDate,
				"units":       item.Units,
				"unit_type":   item.UnitType,
				"unit_price":  item.UnitPrice,
				"amount":      item.Total,
			})
		}

		if meta.Links.Next == "" {
			break
		} else {
			options.Cursor = meta.Links.Next
			continue
		}
	}

	d.SetId(strconv.Itoa(invoiceID))
	if err := d.Set("invoice_items", invoiceItemList); err != nil {
		return diag.Errorf("unable to set `invoice_items` read value: %v", err)
	}

	return nil
}

package vultr

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vultr/govultr/v3"
)

func dataSourceVultrInvoice() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrInvoiceRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
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
	}
}

func dataSourceVultrInvoiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	filters, filtersOk := d.GetOk("filter")

	if !filtersOk {
		return diag.Errorf("issue with filter: %v", filtersOk)
	}

	invoiceList := []govultr.Invoice{}
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

			if filterLoop(f, sm) {
				invoiceList = append(invoiceList, invoice)
			}
		}

		if meta.Links.Next == "" {
			break
		} else {
			options.Cursor = meta.Links.Next
			continue
		}
	}
	if len(invoiceList) > 1 {
		return diag.Errorf("your search returned too many results. Please refine your search to be more specific")
	}

	if len(invoiceList) < 1 {
		return diag.Errorf("no results were found")
	}

	d.SetId(strconv.Itoa(invoiceList[0].ID))
	if err := d.Set("id", invoiceList[0].ID); err != nil {
		return diag.Errorf("unable to set invoice `id` read value: %v", err)
	}
	if err := d.Set("date", invoiceList[0].Date); err != nil {
		return diag.Errorf("unable to set invoice `date` read value: %v", err)
	}
	if err := d.Set("description", invoiceList[0].Description); err != nil {
		return diag.Errorf("unable to set invoice `description` read value: %v", err)
	}
	if err := d.Set("amount", invoiceList[0].Amount); err != nil {
		return diag.Errorf("unable to set invoice `amount` read value: %v", err)
	}
	if err := d.Set("balance", invoiceList[0].Balance); err != nil {
		return diag.Errorf("unable to set invoice `balance` read value: %v", err)
	}
	return nil
}

package vultr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vultr/govultr/v3"
)

func dataSourceVultrCDNPushZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrCDNPushZoneRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cdn_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVultrCDNPushZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	filters, filtersOk := d.GetOk("filter")

	if !filtersOk {
		return diag.Errorf("issue with filter: %v", filtersOk)
	}

	pushZoneList := []govultr.CDNZone{}
	f := buildVultrDataSourceFilter(filters.(*schema.Set))

	pushZones, _, _, err := client.CDN.ListPushZones(ctx)
	if err != nil {
		return diag.Errorf("error getting CDN push zones: %v", err)
	}

	for _, zone := range pushZones {
		sm, err := structToMap(zone)

		if err != nil {
			return diag.FromErr(err)
		}

		if filterLoop(f, sm) {
			pushZoneList = append(pushZoneList, zone)
		}
	}
	if len(pushZoneList) > 1 {
		return diag.Errorf("your search returned too many results. Please refine your search to be more specific")
	}

	if len(pushZoneList) < 1 {
		return diag.Errorf("no results were found")
	}

	d.SetId(pushZoneList[0].ID)
	if err := d.Set("label", pushZoneList[0].Label); err != nil {
		return diag.Errorf("unable to set cdn_push_zone `label` read value: %v", err)
	}
	if err := d.Set("cdn_domain", pushZoneList[0].CDNURL); err != nil {
		return diag.Errorf("unable to set cdn_push_zone `cdn_domain` read value: %v", err)
	}
	if err := d.Set("date_created", pushZoneList[0].DateCreated); err != nil {
		return diag.Errorf("unable to set cdn_push_zone `date_created` read value: %v", err)
	}
	return nil
}


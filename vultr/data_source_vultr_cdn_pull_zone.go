package vultr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vultr/govultr/v3"
)

func dataSourceVultrCDNPullZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrCDNPullZoneRead,
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
			"origin_domain": {
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

func dataSourceVultrCDNPullZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	filters, filtersOk := d.GetOk("filter")

	if !filtersOk {
		return diag.Errorf("issue with filter: %v", filtersOk)
	}

	pullZoneList := []govultr.CDNZone{}
	f := buildVultrDataSourceFilter(filters.(*schema.Set))

	pullZones, _, _, err := client.CDN.ListPullZones(ctx)
	if err != nil {
		return diag.Errorf("error getting CDN pull zones: %v", err)
	}

	for _, zone := range pullZones {
		sm, err := structToMap(zone)

		if err != nil {
			return diag.FromErr(err)
		}

		if filterLoop(f, sm) {
			pullZoneList = append(pullZoneList, zone)
		}
	}
	if len(pullZoneList) > 1 {
		return diag.Errorf("your search returned too many results. Please refine your search to be more specific")
	}

	if len(pullZoneList) < 1 {
		return diag.Errorf("no results were found")
	}

	d.SetId(pullZoneList[0].ID)
	if err := d.Set("label", pullZoneList[0].Label); err != nil {
		return diag.Errorf("unable to set cdn_pull_zone `label` read value: %v", err)
	}
	if err := d.Set("origin_domain", pullZoneList[0].OriginDomain); err != nil {
		return diag.Errorf("unable to set cdn_pull_zone `origin_domain` read value: %v", err)
	}
	if err := d.Set("cdn_domain", pullZoneList[0].CDNURL); err != nil {
		return diag.Errorf("unable to set cdn_pull_zone `cdn_domain` read value: %v", err)
	}
	if err := d.Set("date_created", pullZoneList[0].DateCreated); err != nil {
		return diag.Errorf("unable to set cdn_pull_zone `date_created` read value: %v", err)
	}
	return nil
}

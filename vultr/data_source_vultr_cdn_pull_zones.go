package vultr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVultrCDNPullZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrCDNPullZonesRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"cdn_pull_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
				},
			},
		},
	}
}

func dataSourceVultrCDNPullZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	var pullZoneList []map[string]interface{}
	filters, filtersOk := d.GetOk("filter")
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

			// If filters exist, check if this zone matches
			if filtersOk && !filterLoop(f, sm) {
				continue
			}

			pullZoneList = append(pullZoneList, map[string]interface{}{
				"id":           zone.ID,
				"label":        zone.Label,
				"origin_domain": zone.OriginDomain,
				"cdn_domain":    zone.CDNURL,
				"date_created":  zone.DateCreated,
			})
		}

	d.SetId("cdn_pull_zones")
	if err := d.Set("cdn_pull_zones", pullZoneList); err != nil {
		return diag.Errorf("unable to set `cdn_pull_zones` read value: %v", err)
	}

	return nil
}


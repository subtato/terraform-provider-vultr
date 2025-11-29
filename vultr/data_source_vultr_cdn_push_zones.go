package vultr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVultrCDNPushZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVultrCDNPushZonesRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"cdn_push_zones": {
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

func dataSourceVultrCDNPushZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	var pushZoneList []map[string]interface{}
	filters, filtersOk := d.GetOk("filter")
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

			// If filters exist, check if this zone matches
			if filtersOk && !filterLoop(f, sm) {
				continue
			}

			pushZoneList = append(pushZoneList, map[string]interface{}{
				"id":          zone.ID,
				"label":       zone.Label,
				"cdn_domain":  zone.CDNURL,
				"date_created": zone.DateCreated,
			})
		}

	d.SetId("cdn_push_zones")
	if err := d.Set("cdn_push_zones", pushZoneList); err != nil {
		return diag.Errorf("unable to set `cdn_push_zones` read value: %v", err)
	}

	return nil
}


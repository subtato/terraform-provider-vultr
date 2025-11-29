package vultr

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vultr/govultr/v3"
)

func resourceVultrCDNPushZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVultrCDNPushZoneCreate,
		ReadContext:   resourceVultrCDNPushZoneRead,
		UpdateContext: resourceVultrCDNPushZoneUpdate,
		DeleteContext: resourceVultrCDNPushZoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"label": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "The label for the CDN push zone.",
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

func resourceVultrCDNPushZoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()
	pushZoneReq := &govultr.CDNZoneReq{
		Label: d.Get("label").(string),
	}

	zone, _, err := client.CDN.CreatePushZone(ctx, pushZoneReq)
	if err != nil {
		return diag.Errorf("error creating CDN push zone: %v", err)
	}

	d.SetId(zone.ID)
	log.Printf("[INFO] CDN Push Zone ID: %s", d.Id())

	return resourceVultrCDNPushZoneRead(ctx, d, meta)
}

func resourceVultrCDNPushZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	zone, _, err := client.CDN.GetPushZone(ctx, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Invalid") {
			tflog.Warn(ctx, fmt.Sprintf("Removing CDN push zone (%s) because it is gone", d.Id()))
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting CDN push zone: %v", err)
	}

	if err := d.Set("label", zone.Label); err != nil {
		return diag.Errorf("unable to set resource cdn_push_zone `label` read value: %v", err)
	}
	if err := d.Set("cdn_domain", zone.CDNURL); err != nil {
		return diag.Errorf("unable to set resource cdn_push_zone `cdn_domain` read value: %v", err)
	}
	if err := d.Set("date_created", zone.DateCreated); err != nil {
		return diag.Errorf("unable to set resource cdn_push_zone `date_created` read value: %v", err)
	}

	return nil
}

func resourceVultrCDNPushZoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	updateReq := &govultr.CDNZoneReq{}

	if d.HasChange("label") {
		updateReq.Label = d.Get("label").(string)
	}

	log.Printf("[INFO] Updating CDN Push Zone: %s", d.Id())
	if _, _, err := client.CDN.UpdatePushZone(ctx, d.Id(), updateReq); err != nil {
		return diag.Errorf("error updating CDN push zone (%s): %v", d.Id(), err)
	}

	return resourceVultrCDNPushZoneRead(ctx, d, meta)
}

func resourceVultrCDNPushZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()
	log.Printf("[INFO] Deleting CDN Push Zone: %s", d.Id())

	if err := client.CDN.DeletePushZone(ctx, d.Id()); err != nil {
		return diag.Errorf("error destroying CDN push zone (%s): %v", d.Id(), err)
	}

	return nil
}

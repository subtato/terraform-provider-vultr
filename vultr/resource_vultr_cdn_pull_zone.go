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

func resourceVultrCDNPullZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVultrCDNPullZoneCreate,
		ReadContext:   resourceVultrCDNPullZoneRead,
		UpdateContext: resourceVultrCDNPullZoneUpdate,
		DeleteContext: resourceVultrCDNPullZoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"label": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "The label for the CDN pull zone.",
			},
			"origin_domain": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "The origin domain that the CDN will pull content from. Must be a valid domain name.",
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

func resourceVultrCDNPullZoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()
	pullZoneReq := &govultr.CDNZoneReq{
		Label:        d.Get("label").(string),
		OriginDomain: d.Get("origin_domain").(string),
	}

	zone, _, err := client.CDN.CreatePullZone(ctx, pullZoneReq)
	if err != nil {
		return diag.Errorf("error creating CDN pull zone: %v", err)
	}

	d.SetId(zone.ID)
	log.Printf("[INFO] CDN Pull Zone ID: %s", d.Id())

	return resourceVultrCDNPullZoneRead(ctx, d, meta)
}

func resourceVultrCDNPullZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	zone, _, err := client.CDN.GetPullZone(ctx, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Invalid") {
			tflog.Warn(ctx, fmt.Sprintf("Removing CDN pull zone (%s) because it is gone", d.Id()))
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting CDN pull zone: %v", err)
	}

	if err := d.Set("label", zone.Label); err != nil {
		return diag.Errorf("unable to set resource cdn_pull_zone `label` read value: %v", err)
	}
	if err := d.Set("origin_domain", zone.OriginDomain); err != nil {
		return diag.Errorf("unable to set resource cdn_pull_zone `origin_domain` read value: %v", err)
	}
	if err := d.Set("cdn_domain", zone.CDNURL); err != nil {
		return diag.Errorf("unable to set resource cdn_pull_zone `cdn_domain` read value: %v", err)
	}
	if err := d.Set("date_created", zone.DateCreated); err != nil {
		return diag.Errorf("unable to set resource cdn_pull_zone `date_created` read value: %v", err)
	}

	return nil
}

func resourceVultrCDNPullZoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	updateReq := &govultr.CDNZoneReq{}

	if d.HasChange("label") {
		updateReq.Label = d.Get("label").(string)
	}
	if d.HasChange("origin_domain") {
		updateReq.OriginDomain = d.Get("origin_domain").(string)
	}

	log.Printf("[INFO] Updating CDN Pull Zone: %s", d.Id())
	if _, _, err := client.CDN.UpdatePullZone(ctx, d.Id(), updateReq); err != nil {
		return diag.Errorf("error updating CDN pull zone (%s): %v", d.Id(), err)
	}

	return resourceVultrCDNPullZoneRead(ctx, d, meta)
}

func resourceVultrCDNPullZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()
	log.Printf("[INFO] Deleting CDN Pull Zone: %s", d.Id())

	if err := client.CDN.DeletePullZone(ctx, d.Id()); err != nil {
		return diag.Errorf("error destroying CDN pull zone (%s): %v", d.Id(), err)
	}

	return nil
}


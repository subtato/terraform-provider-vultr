package vultr

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vultr/govultr/v3"
)

func resourceVultrBlockStorage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVultrBlockStorageCreate,
		ReadContext:   resourceVultrBlockStorageRead,
		UpdateContext: resourceVultrBlockStorageUpdate,
		DeleteContext: resourceVultrBlockStorageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"size_gb": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(10),
				Description:  "The size of the block storage in GB. Minimum size is 10 GB.",
			},
			"region": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: IgnoreCase,
			},
			"attached_to_instance": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"live": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cost": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mount_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"storage_opt", "high_perf"}, false),
				ForceNew:     true,
			},
			"attachment_info": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mount_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"attached": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceVultrBlockStorageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	bsReq := &govultr.BlockStorageCreate{
		Region:    d.Get("region").(string),
		SizeGB:    d.Get("size_gb").(int),
		Label:     d.Get("label").(string),
		BlockType: d.Get("block_type").(string),
	}

	bs, _, err := client.BlockStorage.Create(ctx, bsReq)
	if err != nil {
		return diag.Errorf("error creating block storage: %v", err)
	}

	d.SetId(bs.ID)
	log.Printf("[INFO] Block Storage ID: %s", d.Id())

	if _, err = waitForBlockAvailable(ctx, d, "active", []string{"pending"}, "status", meta); err != nil {
		return diag.Errorf("error while waiting for block %s to be completed: %s", d.Id(), err)
	}

	if instanceID, ok := d.GetOk("attached_to_instance"); ok {
		log.Printf("[INFO] Attaching block storage %s to instance %s", d.Id(), instanceID.(string))

		// Wait for the BS state to become active before attaching
		// Use the existing waitForBlockAvailable function
		if _, err := waitForBlockAvailable(ctx, d, "active", []string{"pending"}, "status", meta); err != nil {
			return diag.Errorf("error waiting for block storage to be ready: %v", err)
		}

		attachReq := &govultr.BlockStorageAttach{
			InstanceID: instanceID.(string),
			Live:       govultr.BoolToBoolPtr(d.Get("live").(bool)),
		}

		if err := client.BlockStorage.Attach(ctx, d.Id(), attachReq); err != nil {
			return diag.Errorf("error attaching block storage %s to instance %s: %v", d.Id(), instanceID.(string), err)
		}

		// Wait for attachment to complete by checking the block storage status
		if err := waitForBlockStorageAttachment(ctx, client, d.Id(), instanceID.(string), 30*time.Second); err != nil {
			return diag.Errorf("error waiting for block storage attachment: %v", err)
		}
	}

	return resourceVultrBlockStorageRead(ctx, d, meta)
}

func resourceVultrBlockStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	bs, _, err := client.BlockStorage.Get(ctx, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Invalid block storage ID") {
			tflog.Warn(ctx, fmt.Sprintf("Removing block storage (%s) because it is gone", d.Id()))
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting block storage: %v", err)
	}

	// Note: 'live' is a configuration option, not returned by the API
	// We preserve the configured value in state
	if d.Get("live") != nil {
		if err := d.Set("live", d.Get("live").(bool)); err != nil {
			return diag.Errorf("unable to set resource block_storage `live` read value: %v", err)
		}
	}
	if err := d.Set("date_created", bs.DateCreated); err != nil {
		return diag.Errorf("unable to set resource block_storage `date_created` read value: %v", err)
	}
	if err := d.Set("cost", bs.Cost); err != nil {
		return diag.Errorf("unable to set resource block_storage `cost` read value: %v", err)
	}
	if err := d.Set("status", bs.Status); err != nil {
		return diag.Errorf("unable to set resource block_storage `status` read value: %v", err)
	}
	if err := d.Set("size_gb", bs.SizeGB); err != nil {
		return diag.Errorf("unable to set resource block_storage `size_gb` read value: %v", err)
	}
	if err := d.Set("region", bs.Region); err != nil {
		return diag.Errorf("unable to set resource block_storage `region` read value: %v", err)
	}
	if err := d.Set("attached_to_instance", bs.AttachedToInstance); err != nil {
		return diag.Errorf("unable to set resource block_storage `attached_to_instance` read value: %v", err)
	}
	if err := d.Set("label", bs.Label); err != nil {
		return diag.Errorf("unable to set resource block_storage `label` read value: %v", err)
	}
	if err := d.Set("mount_id", bs.MountID); err != nil {
		return diag.Errorf("unable to set resource block_storage `mount_id` read value: %v", err)
	}
	if err := d.Set("block_type", bs.BlockType); err != nil {
		return diag.Errorf("unable to set resource block_storage `block_type` read value: %v", err)
	}

	// Set attachment info
	attachmentInfo := []map[string]interface{}{}
	if bs.AttachedToInstance != "" {
		attachmentInfo = append(attachmentInfo, map[string]interface{}{
			"instance_id": bs.AttachedToInstance,
			"mount_id":    bs.MountID,
			"attached":    true,
		})
	} else {
		attachmentInfo = append(attachmentInfo, map[string]interface{}{
			"instance_id": "",
			"mount_id":    "",
			"attached":    false,
		})
	}
	if err := d.Set("attachment_info", attachmentInfo); err != nil {
		return diag.Errorf("unable to set resource block_storage `attachment_info` read value: %v", err)
	}

	return nil
}

func resourceVultrBlockStorageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	blockReq := &govultr.BlockStorageUpdate{}
	if d.HasChange("label") {
		blockReq.Label = d.Get("label").(string)
	}

	if d.HasChange("size_gb") {
		blockReq.SizeGB = d.Get("size_gb").(int)
	}

	if err := client.BlockStorage.Update(ctx, d.Id(), blockReq); err != nil {
		return diag.Errorf("error getting block storage: %v", err)
	}

	if d.HasChange("attached_to_instance") {
		old, newVal := d.GetChange("attached_to_instance")
		oldInstanceID := old.(string)
		newInstanceID := newVal.(string)

		// Detach from old instance if needed
		if oldInstanceID != "" && oldInstanceID != newInstanceID {
			// The following check is necessary so we do not erroneously detach
			// after a formerly attached server has been tainted and/or
			// destroyed.
			bs, _, err := client.BlockStorage.Get(ctx, d.Id())
			if err != nil {
				return diag.Errorf("error getting block storage: %v", err)
			}

			if bs.AttachedToInstance != "" && bs.AttachedToInstance == oldInstanceID {
				log.Printf("[INFO] Detaching block storage %s from instance %s", d.Id(), oldInstanceID)

				blockReq := &govultr.BlockStorageDetach{Live: govultr.BoolToBoolPtr(d.Get("live").(bool))}
				err := client.BlockStorage.Detach(ctx, d.Id(), blockReq)
				if err != nil {
					return diag.Errorf("error detaching block storage %s from instance %s: %v", d.Id(), oldInstanceID, err)
				}

				// Wait for detachment to complete
				log.Printf("[INFO] Waiting for block storage detachment to complete...")
				detached := false
				for i := 0; i <= 30; i++ {
					bState, _, err := client.BlockStorage.Get(ctx, d.Id())
					if err != nil {
						return diag.Errorf("error checking detachment status: %s", err.Error())
					}
					if bState.AttachedToInstance == "" {
						detached = true
						log.Printf("[INFO] Block storage successfully detached")
						break
					}
					time.Sleep(1 * time.Second)
				}

				if !detached {
					return diag.Errorf("block storage detachment did not complete within 30 seconds")
				}
			}
		}

		// Attach to new instance if needed
		if newInstanceID != "" && newInstanceID != oldInstanceID {
			log.Printf("[INFO] Attaching block storage %s to instance %s", d.Id(), newInstanceID)
			blockReq := &govultr.BlockStorageAttach{
				InstanceID: newInstanceID,
				Live:       govultr.BoolToBoolPtr(d.Get("live").(bool)),
			}
			if err := client.BlockStorage.Attach(ctx, d.Id(), blockReq); err != nil {
				return diag.Errorf("error attaching block storage %s to instance %s: %v", d.Id(), newInstanceID, err)
			}

			// Wait for attachment to complete
			if err := waitForBlockStorageAttachment(ctx, client, d.Id(), newInstanceID, 30*time.Second); err != nil {
				return diag.Errorf("error waiting for block storage attachment: %v", err)
			}
		}
	}

	return resourceVultrBlockStorageRead(ctx, d, meta)
}

func resourceVultrBlockStorageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	log.Printf("[INFO] Deleting block storage: %s", d.Id())

	// Check if block storage is attached and detach if necessary
	bs, _, err := client.BlockStorage.Get(ctx, d.Id())
	if err != nil {
		// If we can't get it, it might already be deleted, try to delete anyway
		if strings.Contains(err.Error(), "Invalid block storage ID") {
			log.Printf("[INFO] Block storage %s appears to already be deleted", d.Id())
			return nil
		}
		return diag.Errorf("error getting block storage %s during deletion: %v", d.Id(), err)
	}

	// Detach if attached
	if bs.AttachedToInstance != "" {
		log.Printf("[INFO] Detaching block storage %s from instance %s before deletion", d.Id(), bs.AttachedToInstance)
		blockReq := &govultr.BlockStorageDetach{Live: govultr.BoolToBoolPtr(d.Get("live").(bool))}
		if err := client.BlockStorage.Detach(ctx, d.Id(), blockReq); err != nil {
			// If detach fails, still try to delete (might already be detached)
			log.Printf("[WARN] Error detaching block storage %s (will attempt deletion anyway): %v", d.Id(), err)
		} else {
			// Wait for detachment to complete
			if err := waitForBlockStorageDetachment(ctx, client, d.Id(), 30*time.Second); err != nil {
				log.Printf("[WARN] Block storage detachment did not complete within timeout, attempting deletion anyway: %v", err)
			}
		}
	}

	// Delete the block storage
	if err := client.BlockStorage.Delete(ctx, d.Id()); err != nil {
		// Check if error is due to still being attached
		if strings.Contains(err.Error(), "attached") || strings.Contains(err.Error(), "attached to") {
			return diag.Errorf("error deleting block storage %s: storage is still attached. Please detach manually: %v", d.Id(), err)
		}
		return diag.Errorf("error deleting block storage %s: %v", d.Id(), err)
	}

	return nil
}

func waitForBlockAvailable(ctx context.Context, d *schema.ResourceData, target string, pending []string, attribute string, meta interface{}) (interface{}, error) { //nolint:lll
	log.Printf(
		"[INFO] Waiting for Server (%s) to have %s of %s",
		d.Id(), attribute, target)

	stateConf := &retry.StateChangeConf{
		Pending:        pending,
		Target:         []string{target},
		Refresh:        newBlockStateRefresh(ctx, d, meta, attribute),
		Timeout:        60 * time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     3 * time.Second,
		NotFoundChecks: 60,
	}

	return stateConf.WaitForStateContext(ctx)
}

func newBlockStateRefresh(ctx context.Context, d *schema.ResourceData, meta interface{}, attr string) retry.StateRefreshFunc { //nolint:lll
	client := meta.(*Client).govultrClient()
	return func() (interface{}, string, error) {
		log.Printf("[INFO] Creating Block")
		block, _, err := client.BlockStorage.Get(ctx, d.Id())
		if err != nil {
			return nil, "", fmt.Errorf("error retrieving block %s : %s", d.Id(), err)
		}

		if attr == "status" {
			log.Printf("[INFO] The Block Status is %s", block.Status)
			return block, block.Status, nil
		} else {
			return nil, "", nil
		}
	}
}

// waitForBlockStorageAttachment waits for a block storage to be attached to an instance
func waitForBlockStorageAttachment(ctx context.Context, client *govultr.Client, blockID, instanceID string, timeout time.Duration) error {
	log.Printf("[INFO] Waiting for block storage %s to attach to instance %s", blockID, instanceID)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			bState, _, err := client.BlockStorage.Get(ctx, blockID)
			if err != nil {
				return fmt.Errorf("error checking attachment status: %w", err)
			}
			if bState.AttachedToInstance == instanceID && bState.MountID != "" {
				log.Printf("[INFO] Block storage successfully attached with mount_id: %s", bState.MountID)
				return nil
			}
		}
	}

	return fmt.Errorf("block storage attachment did not complete within %v", timeout)
}

// waitForBlockStorageDetachment waits for a block storage to be detached from an instance
func waitForBlockStorageDetachment(ctx context.Context, client *govultr.Client, blockID string, timeout time.Duration) error {
	log.Printf("[INFO] Waiting for block storage %s to detach", blockID)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			bState, _, err := client.BlockStorage.Get(ctx, blockID)
			if err != nil {
				// If we can't get it, it might be deleted already
				if strings.Contains(err.Error(), "Invalid block storage ID") {
					log.Printf("[INFO] Block storage %s appears to be deleted", blockID)
					return nil
				}
				return fmt.Errorf("error checking detachment status: %w", err)
			}
			if bState.AttachedToInstance == "" {
				log.Printf("[INFO] Block storage successfully detached")
				return nil
			}
		}
	}

	return fmt.Errorf("block storage detachment did not complete within %v", timeout)
}

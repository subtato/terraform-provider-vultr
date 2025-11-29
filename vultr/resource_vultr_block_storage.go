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

// isNothingToChangeError checks if the response indicates "Nothing to change"
// This is not actually an error - it means the state is already correct
// The API returns: {"error":"Nothing to change","status":400}
func isNothingToChangeError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	errLower := strings.ToLower(errStr)

	// Check for various formats of the "Nothing to change" response
	// The API may return this in different formats, so we check multiple patterns
	return strings.Contains(errStr, "Nothing to change") ||
		strings.Contains(errStr, `"error":"Nothing to change"`) ||
		strings.Contains(errStr, `error":"Nothing to change`) ||
		strings.Contains(errStr, `{"error":"Nothing to change"`) ||
		strings.Contains(errLower, "nothing to change") ||
		(strings.Contains(errStr, `{"error"`) && strings.Contains(errLower, "nothing to change"))
}

func resourceVultrBlockStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).govultrClient()

	bs, _, err := client.BlockStorage.Get(ctx, d.Id())
	if err != nil {
		errStr := err.Error()

		// "Nothing to change" is not an error - it means the state is already correct
		// This commonly occurs after detachment operations when the API is confirming no changes needed
		isNothingToChange := isNothingToChangeError(err) || (strings.Contains(errStr, "Nothing") && strings.Contains(errStr, "change"))

		if isNothingToChange {
			log.Printf("[INFO] Block storage %s returned 'Nothing to change' - state is already correct, no update needed", d.Id())
			// This is not an error - the state is already in the desired condition
			// Return successfully without updating state, as it's already correct
			return nil
		}

		// Handle actual errors
		if strings.Contains(errStr, "Invalid block storage ID") || strings.Contains(errStr, "not found") {
			tflog.Warn(ctx, fmt.Sprintf("Removing block storage (%s) because it is gone", d.Id()))
			d.SetId("")
			return nil
		}

		// For other actual errors, return them
		log.Printf("[DEBUG] Block storage read error: %s", errStr)
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

		// Detach from old instance if needed (either changing to different instance or removing attachment)
		if oldInstanceID != "" && oldInstanceID != newInstanceID {
			// The following check is necessary so we do not erroneously detach
			// after a formerly attached server has been tainted and/or
			// destroyed.
			bs, _, err := client.BlockStorage.Get(ctx, d.Id())
			if err != nil {
				// "Nothing to change" means the state is already correct
				if isNothingToChangeError(err) {
					log.Printf("[INFO] Block storage %s returned 'Nothing to change' - state is already correct", d.Id())
					// If we're removing attachment, "Nothing to change" likely means it's already detached
					if newInstanceID == "" {
						log.Printf("[INFO] Assuming block storage is already detached (Nothing to change response)")
						bs = nil // Set to nil to skip attachment check below
					} else {
						// If we're attaching to a new instance, "Nothing to change" might mean it's already in the desired state
						// Try to get the state one more time, but if it still says "Nothing to change", proceed
						log.Printf("[INFO] Retrying to get block storage state for attachment operation")
						time.Sleep(1 * time.Second)
						bs, _, err = client.BlockStorage.Get(ctx, d.Id())
						if err != nil && isNothingToChangeError(err) {
							log.Printf("[INFO] Still getting 'Nothing to change', assuming current state is acceptable")
							bs = nil
						} else if err != nil {
							return diag.Errorf("error getting block storage after retry: %v", err)
						}
					}
				} else {
					return diag.Errorf("error getting block storage: %v", err)
				}
			}

			// Only detach if it's actually attached to the old instance
			if bs != nil && bs.AttachedToInstance != "" && bs.AttachedToInstance == oldInstanceID {
				log.Printf("[INFO] Detaching block storage %s from instance %s", d.Id(), oldInstanceID)

				blockReq := &govultr.BlockStorageDetach{Live: govultr.BoolToBoolPtr(d.Get("live").(bool))}
				err := client.BlockStorage.Detach(ctx, d.Id(), blockReq)
				if err != nil {
					errStr := err.Error()
					// "Nothing to change" means it's already detached - not an error
					if isNothingToChangeError(err) {
						log.Printf("[INFO] Block storage %s already detached (Nothing to change response)", d.Id())
					} else if newInstanceID == "" && strings.Contains(errStr, "not attached") {
						// Already detached - not an error when removing attachment
						log.Printf("[INFO] Block storage %s already detached", d.Id())
					} else {
						return diag.Errorf("error detaching block storage %s from instance %s: %v", d.Id(), oldInstanceID, err)
					}
				} else {
					// Wait for detachment to complete
					if err := waitForBlockStorageDetachment(ctx, client, d.Id(), 30*time.Second); err != nil {
						// If detachment didn't complete but we're just removing the attachment (setting to empty),
						// we can continue - the read will handle the state
						if newInstanceID == "" {
							log.Printf("[WARN] Block storage detachment did not complete within timeout, but continuing with removal: %v", err)
						} else {
							return diag.Errorf("error waiting for block storage detachment: %v", err)
						}
					}
				}
			} else if bs != nil && bs.AttachedToInstance == "" {
				// Already detached, nothing to do
				log.Printf("[INFO] Block storage %s is already detached", d.Id())
			} else if bs == nil {
				// Couldn't read state due to "Nothing to change" - assume already detached if removing attachment
				if newInstanceID == "" {
					log.Printf("[INFO] Assuming block storage %s is already detached (could not read state)", d.Id())
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
		errStr := err.Error()
		// "Nothing to change" means the state is already correct - proceed with deletion
		if isNothingToChangeError(err) {
			log.Printf("[INFO] Block storage %s returned 'Nothing to change' - proceeding with deletion", d.Id())
			// Continue to deletion - "Nothing to change" means it's in the desired state
		} else if strings.Contains(errStr, "Invalid block storage ID") || strings.Contains(errStr, "not found") {
			// If we can't get it, it might already be deleted, try to delete anyway
			log.Printf("[INFO] Block storage %s appears to already be deleted", d.Id())
			return nil
		} else {
			return diag.Errorf("error getting block storage %s during deletion: %v", d.Id(), err)
		}
		// Set bs to nil if we got "Nothing to change" so we skip the detachment check
		if isNothingToChangeError(err) {
			bs = nil
		}
	}

	// Detach if attached
	if bs != nil && bs.AttachedToInstance != "" {
		log.Printf("[INFO] Detaching block storage %s from instance %s before deletion", d.Id(), bs.AttachedToInstance)
		blockReq := &govultr.BlockStorageDetach{Live: govultr.BoolToBoolPtr(d.Get("live").(bool))}
		if err := client.BlockStorage.Detach(ctx, d.Id(), blockReq); err != nil {
			// "Nothing to change" means it's already detached - not an error
			if isNothingToChangeError(err) {
				log.Printf("[INFO] Block storage %s already detached (Nothing to change response)", d.Id())
			} else {
				// If detach fails for other reasons, still try to delete (might already be detached)
				log.Printf("[WARN] Error detaching block storage %s (will attempt deletion anyway): %v", d.Id(), err)
			}
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
				// "Nothing to change" means the state is already correct (attached in this case)
				if isNothingToChangeError(err) {
					log.Printf("[INFO] Block storage %s returned 'Nothing to change' - assuming already attached", blockID)
					return nil
				}
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
				errStr := err.Error()
				// Handle "Nothing to change" error - retry after brief delay
				if isNothingToChangeError(err) {
					log.Printf("[DEBUG] Received 'Nothing to change' error, retrying...")
					time.Sleep(2 * time.Second)
					bState, _, err = client.BlockStorage.Get(ctx, blockID)
					if err != nil {
						errStr = err.Error()
						// If still failing with "Nothing to change", assume it's detached
						if isNothingToChangeError(err) {
							log.Printf("[INFO] Block storage %s appears to be detached (API returned 'Nothing to change')", blockID)
							return nil
						}
					}
				}
				// If we can't get it, it might be deleted already
				if err != nil {
					errStr = err.Error()
					if strings.Contains(errStr, "Invalid block storage ID") || strings.Contains(errStr, "not found") {
						log.Printf("[INFO] Block storage %s appears to be deleted", blockID)
						return nil
					}
					// For other errors, continue retrying
					log.Printf("[DEBUG] Error checking detachment status (will retry): %v", err)
					continue
				}
			}
			if bState != nil && bState.AttachedToInstance == "" {
				log.Printf("[INFO] Block storage successfully detached")
				return nil
			}
		}
	}

	return fmt.Errorf("block storage detachment did not complete within %v", timeout)
}

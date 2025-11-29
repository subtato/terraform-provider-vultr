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

func resourceVultrVirtualFileSystemStorage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVultrVirtualFileSystemStorageCreate,
		ReadContext:   resourceVultrVirtualFileSystemStorageRead,
		UpdateContext: resourceVultrVirtualFileSystemStorageUpdate,
		DeleteContext: resourceVultrVirtualFileSystemStorageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: IgnoreCase,
			},
			"size_gb": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(10),
				Description:  "The size of the virtual file system storage in GB. Minimum size is 10 GB.",
			},
			"label": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Default:  nil,
			},
			"attached_instances": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"disk_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "nvme",
				ValidateFunc: validation.StringInSlice([]string{"nvme", "ssd"}, false),
				Description:  "The underlying disk type. Options are `nvme` (default) or `ssd`.",
			},
			// computed fields
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cost": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"charges": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"attachments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mount": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceVultrVirtualFileSystemStorageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { //nolint:lll
	client := meta.(*Client).govultrClient()

	req := govultr.VirtualFileSystemStorageReq{
		Region: d.Get("region").(string),
		Label:  d.Get("label").(string),
		StorageSize: govultr.VirtualFileSystemStorageSize{
			SizeGB: d.Get("size_gb").(int),
		},
	}

	if tagsData, tagsOK := d.GetOk("tags"); tagsOK {
		tags := tagsData.(*schema.Set).List()
		for i := range tags {
			req.Tags = append(req.Tags, tags[i].(string))
		}
	}

	storage, _, err := client.VirtualFileSystemStorage.Create(ctx, &req)
	if err != nil {
		return diag.Errorf("error creating virtual file system storage: %v", err)
	}

	d.SetId(storage.ID)
	log.Printf("[INFO] Virtual File System Storage ID: %s", d.Id())

	if _, err = waitForVirtualFileSystemStorageAvailable(ctx, d, "active", []string{"pending"}, "status", meta); err != nil { //nolint:lll
		return diag.Errorf("error while waiting for virtual file system storage %s to be completed: %s", d.Id(), err)
	}

	if attached, ok := d.GetOk("attached_instances"); ok {
		ids := attached.(*schema.Set).List()
		for i := range ids {
			instanceID := ids[i].(string)
			log.Printf("[INFO] Attaching virtual file system storage %s to instance %s", d.Id(), instanceID)
			attachment, _, err := client.VirtualFileSystemStorage.Attach(ctx, d.Id(), instanceID)
			if err != nil {
				return diag.Errorf(
					"error attaching virtual file system storage %s to instance %s: %v",
					d.Id(),
					instanceID,
					err,
				)
			}

			// Log attachment details if available
			if attachment != nil {
				log.Printf("[INFO] VFS attachment created: state=%s, mount_tag=%d", attachment.State, attachment.MountTag)
			}
		}
	}

	return resourceVultrVirtualFileSystemStorageRead(ctx, d, meta)
}

func resourceVultrVirtualFileSystemStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { //nolint:lll
	client := meta.(*Client).govultrClient()

	storage, _, err := client.VirtualFileSystemStorage.Get(ctx, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Subscription ID Not Found.") {
			tflog.Warn(ctx, fmt.Sprintf("removing virtual file system storage (%s) because it is gone", d.Id()))
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting virtual file system storage: %v", err)
	}

	if err := d.Set("region", storage.Region); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `region` read value: %v", err)
	}
	if err := d.Set("size_gb", storage.StorageSize.SizeGB); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `size_gb` read value: %v", err)
	}
	if err := d.Set("label", storage.Label); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `label` read value: %v", err)
	}
	if err := d.Set("tags", storage.Tags); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `tags` read value: %v", err)
	}
	if err := d.Set("date_created", storage.DateCreated); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `date_created` read value: %v", err)
	}
	if err := d.Set("status", storage.Status); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `status` read value: %v", err)
	}
	if err := d.Set("size_gb", storage.StorageSize.SizeGB); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `size_gb` read value: %v", err)
	}
	if err := d.Set("disk_type", storage.DiskType); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `disk_type` read value: %v", err)
	}
	if err := d.Set("cost", storage.Billing.Monthly); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `cost` read value: %v", err)
	}
	if err := d.Set("charges", storage.Billing.Charges); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `charges` read value: %v", err)
	}

	attachments, _, err := client.VirtualFileSystemStorage.AttachmentList(ctx, d.Id())
	if err != nil {
		return diag.Errorf("unable to retrieve attachments for virtual file system storage %s", d.Id())
	}

	var attInstIDs []string
	var attStates []map[string]interface{}
	if len(attachments) != 0 {
		for i := range attachments {
			attInstIDs = append(attInstIDs, attachments[i].TargetID)
			attStates = append(attStates, map[string]interface{}{
				"instance_id": attachments[i].TargetID,
				"state":       attachments[i].State,
				"mount":       attachments[i].MountTag,
			})
		}
	}

	if err := d.Set("attached_instances", attInstIDs); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `attached_instances` read value: %v", err)
	}
	if err := d.Set("attachments", attStates); err != nil {
		return diag.Errorf("unable to set resource virtual_file_system_storage `attachments` read value: %v", err)
	}

	return nil
}

func resourceVultrVirtualFileSystemStorageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { //nolint:lll
	client := meta.(*Client).govultrClient()

	updateReq := &govultr.VirtualFileSystemStorageUpdateReq{}
	if d.HasChange("label") {
		updateReq.Label = d.Get("label").(string)
	}

	if d.HasChange("size_gb") {
		updateReq.StorageSize.SizeGB = d.Get("size_gb").(int)
	}

	if _, _, err := client.VirtualFileSystemStorage.Update(ctx, d.Id(), updateReq); err != nil {
		return diag.Errorf("error updating virtual file system storage : %v", err)
	}

	if d.HasChange("attached_instances") {
		attOld, attNew := d.GetChange("attached_instances")
		elemsOld := attOld.(*schema.Set).List()
		elemsNew := attNew.(*schema.Set).List()

		var idOld, idNew []string
		for i := range elemsOld {
			idOld = append(idOld, elemsOld[i].(string))
		}

		for i := range elemsNew {
			idNew = append(idNew, elemsNew[i].(string))
		}

		// Find instances to detach: in old but not in new
		idDetach := diffSlice(idNew, idOld)
		// Find instances to attach: in new but not in old
		idAttach := diffSlice(idOld, idNew)

		// Detach instances that are no longer in the list
		for i := range idDetach {
			log.Printf(`[INFO] Detaching virtual file system storage %s from instance %s`, d.Id(), idDetach[i])
			if err := client.VirtualFileSystemStorage.Detach(ctx, d.Id(), idDetach[i]); err != nil {
				return diag.Errorf("error detaching instance %s from virtual file system storage %s: %v", idDetach[i], d.Id(), err)
			}
		}

		// Attach new instances
		for i := range idAttach {
			log.Printf(`[INFO] Attaching virtual file system storage %s to instance %s`, d.Id(), idAttach[i])
			attachment, _, err := client.VirtualFileSystemStorage.Attach(ctx, d.Id(), idAttach[i])
			if err != nil {
				return diag.Errorf("error attaching instance %s to virtual file system storage %s: %v", idAttach[i], d.Id(), err)
			}

			// Wait for attachment to be in ATTACHED state
			if attachment != nil && attachment.State != "ATTACHED" {
				log.Printf("[INFO] Waiting for VFS attachment to instance %s to be in ATTACHED state", idAttach[i])
				if err := waitForVFSAttachment(ctx, client, d.Id(), idAttach[i], 30*time.Second); err != nil {
					return diag.Errorf("error waiting for VFS attachment: %v", err)
				}
			}
		}
	}

	return resourceVultrVirtualFileSystemStorageRead(ctx, d, meta)
}

func resourceVultrVirtualFileSystemStorageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { //nolint:lll
	client := meta.(*Client).govultrClient()

	log.Printf("[INFO] Deleting virtual file system storage: %s", d.Id())

	attachments, _, err := client.VirtualFileSystemStorage.AttachmentList(ctx, d.Id())
	if err != nil {
		return diag.Errorf("unable to retrieve attachments for virtual file system storage %s during deletion", d.Id())
	}

	if len(attachments) != 0 {
		for i := range attachments {
			if err := client.VirtualFileSystemStorage.Detach(ctx, d.Id(), attachments[i].TargetID); err != nil {
				return diag.Errorf(
					"error detaching instance %s from virtual file system storage %s during deletion: %v",
					attachments[i].TargetID,
					d.Id(),
					err,
				)
			}
		}
	}

	retryErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *retry.RetryError {
		err := client.VirtualFileSystemStorage.Delete(ctx, d.Id())
		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "Can not delete this subscription until it is detatched from all machines") { //nolint:misspell,lll
			return retry.RetryableError(fmt.Errorf("virtual file system storage is still attached: %s", err.Error()))
		}

		return retry.NonRetryableError(err)
	})

	if retryErr != nil {
		return diag.Errorf("error destroying virtual file system storage %s: %v", d.Id(), retryErr)
	}

	return nil
}

func waitForVirtualFileSystemStorageAvailable(ctx context.Context, d *schema.ResourceData, target string, pending []string, attribute string, meta interface{}) (interface{}, error) { //nolint:lll
	log.Printf(
		"[INFO] Waiting for virtual file system storage (%s) to have %s of %s",
		d.Id(),
		attribute,
		target,
	)

	stateConf := &retry.StateChangeConf{
		Pending:        pending,
		Target:         []string{target},
		Refresh:        newVirtualFileSystemStorageStateRefresh(ctx, d, meta, attribute),
		Timeout:        60 * time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     3 * time.Second,
		NotFoundChecks: 60,
	}

	return stateConf.WaitForStateContext(ctx)
}

func newVirtualFileSystemStorageStateRefresh(ctx context.Context, d *schema.ResourceData, meta interface{}, attr string) retry.StateRefreshFunc { //nolint:lll
	client := meta.(*Client).govultrClient()
	return func() (interface{}, string, error) {
		log.Printf("[INFO] Checking new virtual file system storage")
		storage, _, err := client.VirtualFileSystemStorage.Get(ctx, d.Id())
		if err != nil {
			return nil, "", fmt.Errorf("error retrieving virtual file system storage %s : %s", d.Id(), err)
		}

		if attr == "status" {
			log.Printf("[INFO] The virtual file system storage status is %s", storage.Status)
			return storage, storage.Status, nil
		} else {
			return nil, "", nil
		}
	}
}

// waitForVFSAttachment waits for a VFS attachment to be in ATTACHED state
func waitForVFSAttachment(ctx context.Context, client *govultr.Client, vfsID, instanceID string, timeout time.Duration) error {
	log.Printf("[INFO] Waiting for VFS %s attachment to instance %s to be in ATTACHED state", vfsID, instanceID)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			attachment, _, err := client.VirtualFileSystemStorage.AttachmentGet(ctx, vfsID, instanceID)
			if err != nil {
				// Attachment might not exist yet, continue waiting
				continue
			}
			if attachment != nil && attachment.State == "ATTACHED" {
				log.Printf("[INFO] VFS attachment successfully attached with mount_tag: %d", attachment.MountTag)
				return nil
			}
		}
	}

	return fmt.Errorf("VFS attachment did not reach ATTACHED state within %v", timeout)
}

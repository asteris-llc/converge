package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

func resourceScalewayVolumeAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayVolumeAttachmentCreate,
		Read:   resourceScalewayVolumeAttachmentRead,
		Delete: resourceScalewayVolumeAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"server": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"volume": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalewayVolumeAttachmentCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	var startServerAgain = false
	server, err := scaleway.GetServer(d.Get("server").(string))
	if err != nil {
		fmt.Printf("Failed getting server: %q", err)
		return err
	}

	// volumes can only be modified when the server is powered off
	if server.State != "stopped" {
		startServerAgain = true

		if err := scaleway.PostServerAction(server.Identifier, "poweroff"); err != nil {
			return err
		}

		if err := waitForServerState(scaleway, server.Identifier, "stopped"); err != nil {
			return err
		}
	}

	volumes := make(map[string]api.ScalewayVolume)
	for i, volume := range server.Volumes {
		volumes[i] = volume
	}

	vol, err := scaleway.GetVolume(d.Get("volume").(string))
	if err != nil {
		return err
	}
	volumes[fmt.Sprintf("%d", len(volumes)+1)] = *vol

	// the API request requires most volume attributes to be unset to succeed
	for k, v := range volumes {
		v.Size = 0
		v.CreationDate = ""
		v.Organization = ""
		v.ModificationDate = ""
		v.VolumeType = ""
		v.Server = nil
		v.ExportURI = ""

		volumes[k] = v
	}

	var req = api.ScalewayServerPatchDefinition{
		Volumes: &volumes,
	}
	if err := scaleway.PatchServer(d.Get("server").(string), req); err != nil {
		return fmt.Errorf("Failed attaching volume to server: %q", err)
	}

	if startServerAgain {
		if err := scaleway.PostServerAction(d.Get("server").(string), "poweron"); err != nil {
			return err
		}

		if err := waitForServerState(scaleway, d.Get("server").(string), "running"); err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprintf("scaleway-server:%s/volume/%s", d.Get("server").(string), d.Get("volume").(string)))

	return resourceScalewayVolumeAttachmentRead(d, m)
}

func resourceScalewayVolumeAttachmentRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	server, err := scaleway.GetServer(d.Get("server").(string))
	if err != nil {
		if serr, ok := err.(api.ScalewayAPIError); ok {
			log.Printf("[DEBUG] Error reading server: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if _, err := scaleway.GetVolume(d.Get("volume").(string)); err != nil {
		if serr, ok := err.(api.ScalewayAPIError); ok {
			log.Printf("[DEBUG] Error reading volume: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	for _, volume := range server.Volumes {
		if volume.Identifier == d.Get("volume").(string) {
			return nil
		}
	}

	log.Printf("[DEBUG] Volume %q not attached to server %q\n", d.Get("volume").(string), d.Get("server").(string))
	d.SetId("")
	return nil
}

func resourceScalewayVolumeAttachmentDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway
	var startServerAgain = false

	server, err := scaleway.GetServer(d.Get("server").(string))
	if err != nil {
		return err
	}

	// volumes can only be modified when the server is powered off
	if server.State != "stopped" {
		startServerAgain = true

		if err := scaleway.PostServerAction(server.Identifier, "poweroff"); err != nil {
			return err
		}

		if err := waitForServerState(scaleway, server.Identifier, "stopped"); err != nil {
			return err
		}
	}

	volumes := make(map[string]api.ScalewayVolume)
	for _, volume := range server.Volumes {
		if volume.Identifier != d.Get("volume").(string) {
			volumes[fmt.Sprintf("%d", len(volumes))] = volume
		}
	}

	// the API request requires most volume attributes to be unset to succeed
	for k, v := range volumes {
		v.Size = 0
		v.CreationDate = ""
		v.Organization = ""
		v.ModificationDate = ""
		v.VolumeType = ""
		v.Server = nil
		v.ExportURI = ""

		volumes[k] = v
	}

	var req = api.ScalewayServerPatchDefinition{
		Volumes: &volumes,
	}
	if err := scaleway.PatchServer(d.Get("server").(string), req); err != nil {
		return err
	}

	if startServerAgain {
		if err := scaleway.PostServerAction(d.Get("server").(string), "poweron"); err != nil {
			return err
		}

		if err := waitForServerState(scaleway, d.Get("server").(string), "running"); err != nil {
			return err
		}
	}

	d.SetId("")

	return nil
}

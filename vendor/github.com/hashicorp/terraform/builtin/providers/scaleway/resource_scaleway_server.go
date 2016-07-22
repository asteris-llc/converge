package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

func resourceScalewayServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayServerCreate,
		Read:   resourceScalewayServerRead,
		Update: resourceScalewayServerUpdate,
		Delete: resourceScalewayServerDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"image": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bootscript": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"ipv4_address_private": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv4_address_public": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"dynamic_ip_required": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"state_detail": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceScalewayServerCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	image := d.Get("image").(string)
	var server = api.ScalewayServerDefinition{
		Name:         d.Get("name").(string),
		Image:        String(image),
		Organization: scaleway.Organization,
	}

	server.DynamicIPRequired = Bool(d.Get("dynamic_ip_required").(bool))
	server.CommercialType = d.Get("type").(string)

	if bootscript, ok := d.GetOk("bootscript"); ok {
		server.Bootscript = String(bootscript.(string))
	}

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			server.Tags = append(server.Tags, tag.(string))
		}
	}

	id, err := scaleway.PostServer(server)
	if err != nil {
		return err
	}

	d.SetId(id)
	if d.Get("state").(string) != "stopped" {
		err = scaleway.PostServerAction(id, "poweron")
		if err != nil {
			return err
		}

		err = waitForServerState(scaleway, id, "running")
	}

	if err != nil {
		return err
	}

	return resourceScalewayServerRead(d, m)
}

func resourceScalewayServerRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway
	server, err := scaleway.GetServer(d.Id())

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

	d.Set("ipv4_address_private", server.PrivateIP)
	d.Set("ipv4_address_public", server.PublicAddress.IP)
	d.Set("state", server.State)
	d.Set("state_detail", server.StateDetail)
	d.Set("tags", server.Tags)

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": server.PublicAddress.IP,
	})

	return nil
}

func resourceScalewayServerUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	var req api.ScalewayServerPatchDefinition

	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}

	if d.HasChange("tags") {
		if raw, ok := d.GetOk("tags"); ok {
			var tags []string
			for _, tag := range raw.([]interface{}) {
				tags = append(tags, tag.(string))
			}
			req.Tags = &tags
		}
	}

	if d.HasChange("dynamic_ip_required") {
		req.DynamicIPRequired = Bool(d.Get("dynamic_ip_required").(bool))
	}

	if err := scaleway.PatchServer(d.Id(), req); err != nil {
		return fmt.Errorf("Failed patching scaleway server: %q", err)
	}

	return resourceScalewayServerRead(d, m)
}

func resourceScalewayServerDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	def, err := scaleway.GetServer(d.Id())
	if err != nil {
		if serr, ok := err.(api.ScalewayAPIError); ok {
			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	err = deleteServerSafe(scaleway, def.Identifier)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

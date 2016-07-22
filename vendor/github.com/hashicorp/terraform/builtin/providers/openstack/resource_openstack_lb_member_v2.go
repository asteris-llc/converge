package openstack

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools"
)

func resourceMemberV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceMemberV2Create,
		Read:   resourceMemberV2Read,
		Update: resourceMemberV2Update,
		Delete: resourceMemberV2Delete,

		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_REGION_NAME", ""),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"tenant_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"protocol_port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"weight": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(int)
					if value < 1 {
						errors = append(errors, fmt.Errorf(
							"Only numbers greater than 0 are supported values for 'weight'"))
					}
					return
				},
			},

			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"admin_state_up": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},

			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceMemberV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	subnetID := d.Get("subnet_id").(string)
	adminStateUp := d.Get("admin_state_up").(bool)
	createOpts := pools.MemberCreateOpts{
		Name:         d.Get("name").(string),
		TenantID:     d.Get("tenant_id").(string),
		Address:      d.Get("address").(string),
		ProtocolPort: d.Get("protocol_port").(int),
		Weight:       d.Get("weight").(int),
		AdminStateUp: &adminStateUp,
	}
	// Must omit if not set
	if subnetID != "" {
		createOpts.SubnetID = subnetID
	}

	poolID := d.Get("pool_id").(string)

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	member, err := pools.CreateAssociateMember(networkingClient, poolID, createOpts).ExtractMember()
	if err != nil {
		return fmt.Errorf("Error creating OpenStack LBaaSV2 member: %s", err)
	}
	log.Printf("[INFO] member ID: %s", member.ID)

	log.Printf("[DEBUG] Waiting for Openstack LBaaSV2 member (%s) to become available.", member.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"PENDING_CREATE"},
		Target:     []string{"ACTIVE"},
		Refresh:    waitForMemberActive(networkingClient, poolID, member.ID),
		Timeout:    2 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	d.SetId(member.ID)

	return resourceMemberV2Read(d, meta)
}

func resourceMemberV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	member, err := pools.GetAssociateMember(networkingClient, d.Get("pool_id").(string), d.Id()).ExtractMember()
	if err != nil {
		return CheckDeleted(d, err, "LBV2 Member")
	}

	log.Printf("[DEBUG] Retreived OpenStack LBaaSV2 Member %s: %+v", d.Id(), member)

	d.Set("name", member.Name)
	d.Set("weight", member.Weight)
	d.Set("admin_state_up", member.AdminStateUp)
	d.Set("tenant_id", member.TenantID)
	d.Set("subnet_id", member.SubnetID)
	d.Set("address", member.Address)
	d.Set("protocol_port", member.ProtocolPort)
	d.Set("id", member.ID)

	return nil
}

func resourceMemberV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	var updateOpts pools.MemberUpdateOpts
	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}
	if d.HasChange("weight") {
		updateOpts.Weight = d.Get("weight").(int)
	}
	if d.HasChange("admin_state_up") {
		asu := d.Get("admin_state_up").(bool)
		updateOpts.AdminStateUp = &asu
	}

	log.Printf("[DEBUG] Updating OpenStack LBaaSV2 Member %s with options: %+v", d.Id(), updateOpts)

	_, err = pools.UpdateAssociateMember(networkingClient, d.Get("pool_id").(string), d.Id(), updateOpts).ExtractMember()
	if err != nil {
		return fmt.Errorf("Error updating OpenStack LBaaSV2 Member: %s", err)
	}

	return resourceMemberV2Read(d, meta)
}

func resourceMemberV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE", "PENDING_DELETE"},
		Target:     []string{"DELETED"},
		Refresh:    waitForMemberDelete(networkingClient, d.Get("pool_id").(string), d.Id()),
		Timeout:    2 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error deleting OpenStack LBaaSV2 Member: %s", err)
	}

	d.SetId("")
	return nil
}

func waitForMemberActive(networkingClient *gophercloud.ServiceClient, poolID string, memberID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		member, err := pools.GetAssociateMember(networkingClient, poolID, memberID).ExtractMember()
		if err != nil {
			return nil, "", err
		}

		// The member resource has no Status attribute, so a successful Get is the best we can do
		log.Printf("[DEBUG] OpenStack LBaaSV2 Member: %+v", member)
		return member, "ACTIVE", nil
	}
}

func waitForMemberDelete(networkingClient *gophercloud.ServiceClient, poolID string, memberID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete OpenStack LBaaSV2 Member %s", memberID)

		member, err := pools.GetAssociateMember(networkingClient, poolID, memberID).ExtractMember()
		if err != nil {
			errCode, ok := err.(*gophercloud.UnexpectedResponseCodeError)
			if !ok {
				return member, "ACTIVE", err
			}
			if errCode.Actual == 404 {
				log.Printf("[DEBUG] Successfully deleted OpenStack LBaaSV2 Member %s", memberID)
				return member, "DELETED", nil
			}
		}

		log.Printf("[DEBUG] Openstack LBaaSV2 Member: %+v", member)
		err = pools.DeleteMember(networkingClient, poolID, memberID).ExtractErr()
		if err != nil {
			errCode, ok := err.(*gophercloud.UnexpectedResponseCodeError)
			if !ok {
				return member, "ACTIVE", err
			}
			if errCode.Actual == 404 {
				log.Printf("[DEBUG] Successfully deleted OpenStack LBaaSV2 Member %s", memberID)
				return member, "DELETED", nil
			}
		}

		log.Printf("[DEBUG] OpenStack LBaaSV2 Member %s still active.", memberID)
		return member, "ACTIVE", nil
	}
}

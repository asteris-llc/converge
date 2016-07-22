package azurerm

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceArmPublicIp() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmPublicIpCreate,
		Read:   resourceArmPublicIpRead,
		Update: resourceArmPublicIpCreate,
		Delete: resourceArmPublicIpDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: azureRMNormalizeLocation,
			},

			"resource_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"public_ip_address_allocation": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validatePublicIpAllocation,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},

			"idle_timeout_in_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(int)
					if value < 4 || value > 30 {
						errors = append(errors, fmt.Errorf(
							"The idle timeout must be between 4 and 30 minutes"))
					}
					return
				},
			},

			"domain_name_label": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validatePublicIpDomainNameLabel,
			},

			"reverse_fqdn": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmPublicIpCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	publicIPClient := client.publicIPClient

	log.Printf("[INFO] preparing arguments for Azure ARM Public IP creation.")

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resGroup := d.Get("resource_group_name").(string)
	tags := d.Get("tags").(map[string]interface{})

	properties := network.PublicIPAddressPropertiesFormat{
		PublicIPAllocationMethod: network.IPAllocationMethod(d.Get("public_ip_address_allocation").(string)),
	}

	dnl, hasDnl := d.GetOk("domain_name_label")
	rfqdn, hasRfqdn := d.GetOk("reverse_fqdn")

	if hasDnl || hasRfqdn {
		dnsSettings := network.PublicIPAddressDNSSettings{}

		if hasRfqdn {
			reverse_fqdn := rfqdn.(string)
			dnsSettings.ReverseFqdn = &reverse_fqdn
		}

		if hasDnl {
			domain_name_label := dnl.(string)
			dnsSettings.DomainNameLabel = &domain_name_label

		}

		properties.DNSSettings = &dnsSettings
	}

	if v, ok := d.GetOk("idle_timeout_in_minutes"); ok {
		idle_timeout := v.(int32)
		properties.IdleTimeoutInMinutes = &idle_timeout
	}

	publicIp := network.PublicIPAddress{
		Name:       &name,
		Location:   &location,
		Properties: &properties,
		Tags:       expandTags(tags),
	}

	_, err := publicIPClient.CreateOrUpdate(resGroup, name, publicIp, make(chan struct{}))
	if err != nil {
		return err
	}

	read, err := publicIPClient.Get(resGroup, name, "")
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Public IP %s (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmPublicIpRead(d, meta)
}

func resourceArmPublicIpRead(d *schema.ResourceData, meta interface{}) error {
	publicIPClient := meta.(*ArmClient).publicIPClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["publicIPAddresses"]

	resp, err := publicIPClient.Get(resGroup, name, "")
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error making Read request on Azure public ip %s: %s", name, err)
	}

	d.Set("location", resp.Location)
	d.Set("name", resp.Name)
	d.Set("public_ip_address_allocation", strings.ToLower(string(resp.Properties.PublicIPAllocationMethod)))

	if resp.Properties.DNSSettings != nil && resp.Properties.DNSSettings.Fqdn != nil && *resp.Properties.DNSSettings.Fqdn != "" {
		d.Set("fqdn", resp.Properties.DNSSettings.Fqdn)
	}

	if resp.Properties.IPAddress != nil && *resp.Properties.IPAddress != "" {
		d.Set("ip_address", resp.Properties.IPAddress)
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmPublicIpDelete(d *schema.ResourceData, meta interface{}) error {
	publicIPClient := meta.(*ArmClient).publicIPClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["publicIPAddresses"]

	_, err = publicIPClient.Delete(resGroup, name, make(chan struct{}))

	return err
}

func validatePublicIpAllocation(v interface{}, k string) (ws []string, errors []error) {
	value := strings.ToLower(v.(string))
	allocations := map[string]bool{
		"static":  true,
		"dynamic": true,
	}

	if !allocations[value] {
		errors = append(errors, fmt.Errorf("Public IP Allocation can only be Static of Dynamic"))
	}
	return
}

func validatePublicIpDomainNameLabel(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q: %q",
			k, value))
	}

	if len(value) > 61 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 61 characters: %q", k, value))
	}

	if len(value) == 0 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be an empty string: %q", k, value))
	}
	if regexp.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen: %q", k, value))
	}

	return
}

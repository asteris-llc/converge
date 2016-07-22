package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAmiDataSource_natInstance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsAmiDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami.nat_ami"),
					// Check attributes. Some attributes are tough to test - any not contained here should not be considered
					// stable and should not be used in interpolation. Exception to block_device_mappings which should both
					// show up consistently and break if certain references are not available. However modification of the
					// snapshot ID which is bound to happen on the NAT AMIs will cause testing to break consistently, so
					// deep inspection is not included, simply the count is checked.
					// Tags and product codes may need more testing, but I'm having a hard time finding images with
					// these attributes set.
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "block_device_mappings.#", "1"),
					resource.TestMatchResourceAttr("data.aws_ami.nat_ami", "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr("data.aws_ami.nat_ami", "description", regexp.MustCompile("^Amazon Linux AMI")),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "hypervisor", "xen"),
					resource.TestMatchResourceAttr("data.aws_ami.nat_ami", "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr("data.aws_ami.nat_ami", "image_location", regexp.MustCompile("^amazon/")),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "image_owner_alias", "amazon"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "image_type", "machine"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "most_recent", "true"),
					resource.TestMatchResourceAttr("data.aws_ami.nat_ami", "name", regexp.MustCompile("^amzn-ami-vpc-nat")),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "owner_id", "137112412989"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "public", "true"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "product_codes.#", "0"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "root_device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "root_device_type", "ebs"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "state", "available"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "tags.#", "0"),
					resource.TestCheckResourceAttr("data.aws_ami.nat_ami", "virtualization_type", "hvm"),
				),
			},
		},
	})
}
func TestAccAWSAmiDataSource_windowsInstance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsAmiDataSourceWindowsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami.windows_ami"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "block_device_mappings.#", "27"),
					resource.TestMatchResourceAttr("data.aws_ami.windows_ami", "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr("data.aws_ami.windows_ami", "description", regexp.MustCompile("^Microsoft Windows Server 2012")),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "hypervisor", "xen"),
					resource.TestMatchResourceAttr("data.aws_ami.windows_ami", "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr("data.aws_ami.windows_ami", "image_location", regexp.MustCompile("^amazon/")),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "image_owner_alias", "amazon"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "image_type", "machine"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "most_recent", "true"),
					resource.TestMatchResourceAttr("data.aws_ami.windows_ami", "name", regexp.MustCompile("^Windows_Server-2012-R2")),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "owner_id", "801119661308"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "platform", "windows"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "public", "true"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "product_codes.#", "0"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "root_device_type", "ebs"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "state", "available"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "tags.#", "0"),
					resource.TestCheckResourceAttr("data.aws_ami.windows_ami", "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccAWSAmiDataSource_instanceStore(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsAmiDataSourceInstanceStoreConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami.instance_store_ami"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "block_device_mappings.#", "0"),
					resource.TestMatchResourceAttr("data.aws_ami.instance_store_ami", "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "hypervisor", "xen"),
					resource.TestMatchResourceAttr("data.aws_ami.instance_store_ami", "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr("data.aws_ami.instance_store_ami", "image_location", regexp.MustCompile("images/hvm-instance/ubuntu-trusty-14.04-amd64-server")),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "image_type", "machine"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "most_recent", "true"),
					resource.TestMatchResourceAttr("data.aws_ami.instance_store_ami", "name", regexp.MustCompile("^ubuntu/images/hvm-instance/ubuntu-trusty-14.04-amd64-server")),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "owner_id", "099720109477"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "public", "true"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "product_codes.#", "0"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "root_device_type", "instance-store"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "state", "available"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "tags.#", "0"),
					resource.TestCheckResourceAttr("data.aws_ami.instance_store_ami", "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccAWSAmiDataSource_owners(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsAmiDataSourceOwnersConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami.amazon_ami"),
				),
			},
		},
	})
}

func testAccCheckAwsAmiDataSourceDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckAwsAmiDataSourceID(n string) resource.TestCheckFunc {
	// Wait for IAM role
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AMI data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AMI data source ID not set")
		}
		return nil
	}
}

// Using NAT AMIs for testing - I would expect with NAT gateways now a thing,
// that this will possibly be deprecated at some point in time. Other candidates
// for testing this after that may be Ubuntu's AMI's, or Amazon's regular
// Amazon Linux AMIs.
const testAccCheckAwsAmiDataSourceConfig = `
data "aws_ami" "nat_ami" {
	most_recent = true
	filter {
		name = "owner-alias"
		values = ["amazon"]
	}
	filter {
		name = "name"
		values = ["amzn-ami-vpc-nat*"]
	}
	filter {
		name = "virtualization-type"
		values = ["hvm"]
	}
	filter {
		name = "root-device-type"
		values = ["ebs"]
	}
	filter {
		name = "block-device-mapping.volume-type"
		values = ["standard"]
	}
}
`

// Windows image test.
const testAccCheckAwsAmiDataSourceWindowsConfig = `
data "aws_ami" "windows_ami" {
	most_recent = true
	filter {
		name = "owner-alias"
		values = ["amazon"]
	}
	filter {
		name = "name"
		values = ["Windows_Server-2012-R2*"]
	}
	filter {
		name = "virtualization-type"
		values = ["hvm"]
	}
	filter {
		name = "root-device-type"
		values = ["ebs"]
	}
	filter {
		name = "block-device-mapping.volume-type"
		values = ["gp2"]
	}
}
`

// Instance store test - using Ubuntu images
const testAccCheckAwsAmiDataSourceInstanceStoreConfig = `
data "aws_ami" "instance_store_ami" {
	most_recent = true
	filter {
		name = "owner-id"
		values = ["099720109477"]
	}
	filter {
		name = "name"
		values = ["ubuntu/images/hvm-instance/ubuntu-trusty-14.04-amd64-server*"]
	}
	filter {
		name = "virtualization-type"
		values = ["hvm"]
	}
	filter {
		name = "root-device-type"
		values = ["instance-store"]
	}
}
`

// Testing owner parameter
const testAccCheckAwsAmiDataSourceOwnersConfig = `
data "aws_ami" "amazon_ami" {
	most_recent = true
	owners = ["amazon"]
}
`

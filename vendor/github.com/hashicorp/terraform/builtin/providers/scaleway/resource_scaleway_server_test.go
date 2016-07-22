package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccScalewayServer_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayServerConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerAttributes("scaleway_server.base"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "type", "C1"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "name", "test"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "tags.0", "terraform-test"),
				),
			},
		},
	})
}

func testAccCheckScalewayServerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).scaleway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		_, err := client.GetServer(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Server still exists")
		}
	}

	return nil
}

func testAccCheckScalewayServerAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Unknown resource: %s", n)
		}

		client := testAccProvider.Meta().(*Client).scaleway
		server, err := client.GetServer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if server.Name != "test" {
			return fmt.Errorf("Server has wrong name")
		}
		if server.Image.Identifier != armImageIdentifier {
			return fmt.Errorf("Wrong server image")
		}
		if server.CommercialType != "C1" {
			return fmt.Errorf("Wrong server type")
		}

		return nil
	}
}

func testAccCheckScalewayServerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server ID is set")
		}

		client := testAccProvider.Meta().(*Client).scaleway
		server, err := client.GetServer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if server.Identifier != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		return nil
	}
}

var armImageIdentifier = "5faef9cd-ea9b-4a63-9171-9e26bec03dbc"

var testAccCheckScalewayServerConfig = fmt.Sprintf(`
resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]
}`, armImageIdentifier)

package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSBeanstalkEnv_basic(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBeanstalkEnvConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_tier(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription
	beanstalkQueuesNameRegexp := regexp.MustCompile("https://sqs.+?awseb[^,]+")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBeanstalkWorkerEnvConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvTier("aws_elastic_beanstalk_environment.tfenvtest", &app),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "queues.0", beanstalkQueuesNameRegexp),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_outputs(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription
	beanstalkAsgNameRegexp := regexp.MustCompile("awseb.+?AutoScalingGroup[^,]+")
	beanstalkElbNameRegexp := regexp.MustCompile("awseb.+?EBLoa[^,]+")
	beanstalkInstancesNameRegexp := regexp.MustCompile("i-([0-9a-fA-F]{8}|[0-9a-fA-F]{17})")
	beanstalkLcNameRegexp := regexp.MustCompile("awseb.+?AutoScalingLaunch[^,]+")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBeanstalkEnvConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "autoscaling_groups.0", beanstalkAsgNameRegexp),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "load_balancers.0", beanstalkElbNameRegexp),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "instances.0", beanstalkInstancesNameRegexp),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "launch_configurations.0", beanstalkLcNameRegexp),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_cname_prefix(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription
	cnamePrefix := acctest.RandString(8)
	beanstalkCnameRegexp := regexp.MustCompile("^" + cnamePrefix + ".+?elasticbeanstalk.com$")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBeanstalkEnvCnamePrefixConfig(cnamePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "cname", beanstalkCnameRegexp),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_config(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBeanstalkConfigTemplate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tftest", &app),
					testAccCheckBeanstalkEnvConfigValue("aws_elastic_beanstalk_environment.tftest", "1"),
				),
			},

			resource.TestStep{
				Config: testAccBeanstalkConfigTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tftest", &app),
					testAccCheckBeanstalkEnvConfigValue("aws_elastic_beanstalk_environment.tftest", "2"),
				),
			},

			resource.TestStep{
				Config: testAccBeanstalkConfigTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tftest", &app),
					testAccCheckBeanstalkEnvConfigValue("aws_elastic_beanstalk_environment.tftest", "3"),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_resource(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBeanstalkResourceOptionSetting,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_vpc(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBeanstalkEnv_VPC(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.default", &app),
				),
			},
		},
	})
}

func testAccCheckBeanstalkEnvDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastic_beanstalk_environment" {
			continue
		}

		// Try to find the environment
		describeBeanstalkEnvOpts := &elasticbeanstalk.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeEnvironments(describeBeanstalkEnvOpts)
		if err == nil {
			switch {
			case len(resp.Environments) > 1:
				return fmt.Errorf("Error %d environments match, expected 1", len(resp.Environments))
			case len(resp.Environments) == 1:
				if *resp.Environments[0].Status == "Terminated" {
					return nil
				}
				return fmt.Errorf("Elastic Beanstalk ENV still exists.")
			default:
				return nil
			}
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidBeanstalkEnvID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckBeanstalkEnvExists(n string, app *elasticbeanstalk.EnvironmentDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		env, err := describeBeanstalkEnv(testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn, aws.String(rs.Primary.ID))
		if err != nil {
			return err
		}

		*app = *env

		return nil
	}
}

func testAccCheckBeanstalkEnvTier(n string, app *elasticbeanstalk.EnvironmentDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		env, err := describeBeanstalkEnv(testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn, aws.String(rs.Primary.ID))
		if err != nil {
			return err
		}
		if *env.Tier.Name != "Worker" {
			return fmt.Errorf("Beanstalk Environment tier is %s, expected Worker", *env.Tier.Name)
		}

		*app = *env

		return nil
	}
}

func testAccCheckBeanstalkEnvConfigValue(n string, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		resp, err := conn.DescribeConfigurationOptions(&elasticbeanstalk.DescribeConfigurationOptionsInput{
			ApplicationName: aws.String(rs.Primary.Attributes["application"]),
			EnvironmentName: aws.String(rs.Primary.Attributes["name"]),
			Options: []*elasticbeanstalk.OptionSpecification{
				{
					Namespace:  aws.String("aws:elasticbeanstalk:application:environment"),
					OptionName: aws.String("TEMPLATE"),
				},
			},
		})
		if err != nil {
			return err
		}

		if len(resp.Options) != 1 {
			return fmt.Errorf("Found %d options, expected 1.", len(resp.Options))
		}

		log.Printf("[DEBUG] %d Elastic Beanstalk Option values returned.", len(resp.Options[0].ValueOptions))

		for _, value := range resp.Options[0].ValueOptions {
			if *value != expectedValue {
				return fmt.Errorf("Option setting value: %s. Expected %s", *value, expectedValue)
			}
		}

		return nil
	}
}

func describeBeanstalkEnv(conn *elasticbeanstalk.ElasticBeanstalk,
	envID *string) (*elasticbeanstalk.EnvironmentDescription, error) {
	describeBeanstalkEnvOpts := &elasticbeanstalk.DescribeEnvironmentsInput{
		EnvironmentIds: []*string{envID},
	}

	log.Printf("[DEBUG] Elastic Beanstalk Environment TEST describe opts: %s", describeBeanstalkEnvOpts)

	resp, err := conn.DescribeEnvironments(describeBeanstalkEnvOpts)
	if err != nil {
		return &elasticbeanstalk.EnvironmentDescription{}, err
	}
	if len(resp.Environments) == 0 {
		return &elasticbeanstalk.EnvironmentDescription{}, fmt.Errorf("Elastic Beanstalk ENV not found.")
	}
	if len(resp.Environments) > 1 {
		return &elasticbeanstalk.EnvironmentDescription{}, fmt.Errorf("Found %d environments, expected 1.", len(resp.Environments))
	}
	return resp.Environments[0], nil
}

const testAccBeanstalkEnvConfig = `
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name = "tf-test-name"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"
}
`

const testAccBeanstalkWorkerEnvConfig = `
resource "aws_iam_instance_profile" "tftest" {
  name = "tftest_profile"
  roles = ["${aws_iam_role.tftest.name}"]
}

resource "aws_iam_role" "tftest" {
  name = "tftest_role"
  path = "/"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"ec2.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
}

resource "aws_iam_role_policy" "tftest" {
  name = "tftest_policy"
  role = "${aws_iam_role.tftest.id}"
  policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"QueueAccess\",\"Action\":[\"sqs:ChangeMessageVisibility\",\"sqs:DeleteMessage\",\"sqs:ReceiveMessage\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"
}

resource "aws_elastic_beanstalk_application" "tftest" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name = "tf-test-name"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  tier = "Worker"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = "${aws_iam_instance_profile.tftest.name}"
  }
}
`

func testAccBeanstalkEnvCnamePrefixConfig(randString string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
name = "tf-test-name"
description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
name = "tf-test-name"
application = "${aws_elastic_beanstalk_application.tftest.name}"
cname_prefix = "%s"
solution_stack_name = "64bit Amazon Linux running Python"
}
`, randString)
}

const testAccBeanstalkConfigTemplate = `
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tftest" {
  name = "tf-test-name"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  template_name = "${aws_elastic_beanstalk_configuration_template.tftest.name}"
}

resource "aws_elastic_beanstalk_configuration_template" "tftest" {
  name        = "tf-test-original"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "TEMPLATE"
    value     = "1"
 }
}
`

const testAccBeanstalkConfigTemplateUpdate = `
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tftest" {
  name = "tf-test-name"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  template_name = "${aws_elastic_beanstalk_configuration_template.tftest.name}"
}

resource "aws_elastic_beanstalk_configuration_template" "tftest" {
  name        = "tf-test-updated"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "TEMPLATE"
    value     = "2"
  }
}
`

const testAccBeanstalkConfigTemplateOverride = `
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tftest" {
  name = "tf-test-name"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  template_name = "${aws_elastic_beanstalk_configuration_template.tftest.name}"

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "TEMPLATE"
    value     = "3"
  }
}

resource "aws_elastic_beanstalk_configuration_template" "tftest" {
  name        = "tf-test-updated"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "TEMPLATE"
    value     = "2"
  }
}
`
const testAccBeanstalkResourceOptionSetting = `
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name = "tf-test-name"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource = "ScheduledAction01"
    name = "MinSize"
    value = "2"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource = "ScheduledAction01"
    name = "MaxSize"
    value = "6"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource = "ScheduledAction01"
    name = "Recurrence"
    value = "0 8 * * *"
  }
}
`

func testAccBeanstalkEnv_VPC(name string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "tf_b_test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "tf_b_test" {
  vpc_id = "${aws_vpc.tf_b_test.id}"
}

resource "aws_route" "r" {
  route_table_id = "${aws_vpc.tf_b_test.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.tf_b_test.id}"
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.tf_b_test.id}"
  cidr_block = "10.0.0.0/24"
}

resource "aws_security_group" "default" {
  name = "tf-b-test-%s"
  vpc_id = "${aws_vpc.tf_b_test.id}"
}

resource "aws_elastic_beanstalk_application" "default" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "default" {
  name = "tf-test-name"
  application = "${aws_elastic_beanstalk_application.default.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = "${aws_vpc.tf_b_test.id}"
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = "${aws_subnet.main.id}"
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = "${aws_security_group.default.id}"
  }
}
`, name)
}

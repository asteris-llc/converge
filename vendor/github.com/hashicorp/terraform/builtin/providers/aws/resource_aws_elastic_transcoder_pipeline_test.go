package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSElasticTranscoderPipeline_basic(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elastictranscoder_pipeline.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: awsElasticTranscoderPipelineConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists("aws_elastictranscoder_pipeline.bar", pipeline),
				),
			},
		},
	})
}

func TestAccAWSElasticTranscoderPipeline_withContentConfig(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elastictranscoder_pipeline.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: awsElasticTranscoderPipelineWithContentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists("aws_elastictranscoder_pipeline.bar", pipeline),
				),
			},
			resource.TestStep{
				Config: awsElasticTranscoderPipelineWithContentConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists("aws_elastictranscoder_pipeline.bar", pipeline),
				),
			},
		},
	})
}

func TestAccAWSElasticTranscoderPipeline_withPermissions(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elastictranscoder_pipeline.baz",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: awsElasticTranscoderPipelineWithPerms,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists("aws_elastictranscoder_pipeline.baz", pipeline),
				),
			},
		},
	})
}

func testAccCheckAWSElasticTranscoderPipelineExists(n string, res *elastictranscoder.Pipeline) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pipeline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elastictranscoderconn

		out, err := conn.ReadPipeline(&elastictranscoder.ReadPipelineInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*res = *out.Pipeline

		return nil
	}
}

func testAccCheckElasticTranscoderPipelineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elastictranscoderconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastictranscoder_pipline" {
			continue
		}

		out, err := conn.ReadPipeline(&elastictranscoder.ReadPipelineInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if out.Pipeline != nil && *out.Pipeline.Id == rs.Primary.ID {
				return fmt.Errorf("Elastic Transcoder Pipeline still exists")
			}
		}

		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		if awsErr.Code() != "ResourceNotFoundException" {
			return fmt.Errorf("unexpected error: %s", awsErr)
		}

	}
	return nil
}

const awsElasticTranscoderPipelineConfigBasic = `
resource "aws_elastictranscoder_pipeline" "bar" {
  input_bucket  = "${aws_s3_bucket.test_bucket.bucket}"
  output_bucket = "${aws_s3_bucket.test_bucket.bucket}"
  name          = "aws_elastictranscoder_pipeline_tf_test_"
  role          = "${aws_iam_role.test_role.arn}"
}

resource "aws_iam_role" "test_role" {
  name = "aws_elastictranscoder_pipeline_tf_test_role_"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test_bucket" {
  bucket = "aws-elasticencoder-pipeline-tf-test-bucket"
  acl    = "private"
}
`

const awsElasticTranscoderPipelineWithContentConfig = `
resource "aws_elastictranscoder_pipeline" "bar" {
  input_bucket = "${aws_s3_bucket.content_bucket.bucket}"
  name         = "aws_elastictranscoder_pipeline_tf_test_"
  role         = "${aws_iam_role.test_role.arn}"

  content_config {
    bucket        = "${aws_s3_bucket.content_bucket.bucket}"
    storage_class = "Standard"
  }

  thumbnail_config {
    bucket        = "${aws_s3_bucket.content_bucket.bucket}"
    storage_class = "Standard"
  }
}

resource "aws_iam_role" "test_role" {
  name = "aws_elastictranscoder_pipeline_tf_test_role_"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "content_bucket" {
  bucket = "aws-elasticencoder-pipeline-tf-content-bucket"
  acl    = "private"
}

resource "aws_s3_bucket" "input_bucket" {
  bucket = "aws-elasticencoder-pipeline-tf-input-bucket"
  acl    = "private"
}

resource "aws_s3_bucket" "thumb_bucket" {
  bucket = "aws-elasticencoder-pipeline-tf-thumb-bucket"
  acl    = "private"
}
`

const awsElasticTranscoderPipelineWithContentConfigUpdate = `
resource "aws_elastictranscoder_pipeline" "bar" {
  input_bucket = "${aws_s3_bucket.input_bucket.bucket}"
  name         = "aws_elastictranscoder_pipeline_tf_test_"
  role         = "${aws_iam_role.test_role.arn}"

  content_config {
    bucket        = "${aws_s3_bucket.content_bucket.bucket}"
    storage_class = "Standard"
  }

  thumbnail_config {
    bucket        = "${aws_s3_bucket.thumb_bucket.bucket}"
    storage_class = "Standard"
  }
}

resource "aws_iam_role" "test_role" {
  name = "aws_elastictranscoder_pipeline_tf_test_role_"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "content_bucket" {
  bucket = "aws-elasticencoder-pipeline-tf-content-bucket"
  acl    = "private"
}

resource "aws_s3_bucket" "input_bucket" {
  bucket = "aws-elasticencoder-pipeline-tf-input-bucket"
  acl    = "private"
}

resource "aws_s3_bucket" "thumb_bucket" {
  bucket = "aws-elasticencoder-pipeline-tf-thumb-bucket"
  acl    = "private"
}
`

const awsElasticTranscoderPipelineWithPerms = `
resource "aws_elastictranscoder_pipeline" "baz" {
  input_bucket = "${aws_s3_bucket.content_bucket.bucket}"
  name         = "aws_elastictranscoder_pipeline_tf_test_"
  role         = "${aws_iam_role.test_role.arn}"

  content_config {
    bucket        = "${aws_s3_bucket.content_bucket.bucket}"
    storage_class = "Standard"
  }

  content_config_permissions = {
    grantee_type = "Group"
    grantee      = "AuthenticatedUsers"
    access       = ["FullControl"]
  }

  thumbnail_config {
    bucket        = "${aws_s3_bucket.content_bucket.bucket}"
    storage_class = "Standard"
  }

  thumbnail_config_permissions = {
    grantee_type = "Group"
    grantee      = "AuthenticatedUsers"
    access       = ["FullControl"]
  }
}

resource "aws_iam_role" "test_role" {
  name = "aws_elastictranscoder_pipeline_tf_test_role_"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "content_bucket" {
  bucket = "aws-elasticencoder-pipeline-tf-content-bucket"
  acl    = "private"
}
`

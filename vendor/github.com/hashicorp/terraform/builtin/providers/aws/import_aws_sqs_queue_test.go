package aws

import (
	"testing"

	"fmt"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSSQSQueue_importBasic(t *testing.T) {
	resourceName := "aws_sqs_queue.queue-with-defaults"
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				//The name is never returned after the initial create of the queue.
				//It is part of the URL and can be split down if needed
				//ImportStateVerifyIgnore: []string{"name"},
			},
		},
	})
}

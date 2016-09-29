param "converge-bin-dir" {}
module "docker-test.hcl" "basic" {
	params {
		test-case = "/samples/basic.hcl"
		converge-bin-dir = "{{param `converge-bin-dir`}}"
	}
}

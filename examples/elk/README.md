# converge-elk

A Converge example that sets up a single node docker-based [ELK](https://www.elastic.co/webinars/introduction-elk-stack) stack.

[Filebeat](https://www.elastic.co/products/beats/filebeat) is used instead of [Logstash](https://www.elastic.co/products/logstash) for the log collection component.

## Usage

### Vagrant

In the [Vagrantfile](./Vagrantfile), change the file provisioner source to point to a version of the converge binary built with the `linux/amd64` OS architecture.

After running `vagrant up`, you should have a working [Kibana](https://www.elastic.co/products/kibana) instance backed by [Elasticsearch](https://www.elastic.co/products/elasticsearch). Filebeat is installed on the Vagrant host and is configured to send logs to Elasticsearch.

After Vagrant provisioning is complete, you should be able to access the Kibana web interface at [http://localhost:5601](http://localhost:5601).

### Terraform (AWS)

You must have a version of the [Converge Terraform provisioner](https://github.com/ChrisAubuchon/terraform-provisioner-converge) built and configured as a plugin for Terraform:

```shell
$ cat ~/.terraformrc
provisioners {
  converge = "/path/to/terraform-provisioner-converge"
}
```

You must have also set valid [AWS credentials](https://www.terraform.io/docs/providers/aws/index.html) (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`) in your environment. Then you can run:

```
terraform apply
```

After provisioning completes, you should be able to access the url for the Kibana interface by running:

```shell
echo "http://$(terraform output ip):5601/"
```

## Graphs

![elk graph](./graphs/elk.png)

## Warning

When deploying via Terraform, Kibana will be publicly accessible on port 5601 without authentication. You can adjust the security group in `main.tf` to change this behavior.

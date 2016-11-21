provider "aws" {
  region = "${var.region}"
}

data "aws_ami" "ami" {
  most_recent      = true
  executable_users = ["all", "self"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04*"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "owner-id"
    values = ["099720109477"] # ubuntu
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "hypervisor"
    values = ["xen"]
  }

  filter {
    name   = "state"
    values = ["available"]
  }
}

resource "aws_vpc" "default" {
  cidr_block = "${var.vpc_cidr_block}"

  tags {
    Name = "${var.name}"
  }
}

resource "aws_internet_gateway" "default" {
  vpc_id = "${aws_vpc.default.id}"

  tags {
    Name = "${var.name}"
  }
}

resource "aws_route" "internet_access" {
  route_table_id         = "${aws_vpc.default.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.default.id}"
}

resource "aws_subnet" "default" {
  vpc_id                  = "${aws_vpc.default.id}"
  cidr_block              = "${var.vpc_cidr_block}"
  map_public_ip_on_launch = true

  tags {
    Name = "${var.name}"
  }
}

resource "aws_security_group" "kube_apiserver" {
  name        = "${var.name}-kube-apiserver"
  description = "default security group for kubernetes api server load balancer"
  vpc_id      = "${aws_vpc.default.id}"

  tags {
    Name = "${var.name}-kube-apiserver"
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "controller" {
  name        = "${var.name}-controller"
  description = "default security group for kubernetes controllers"
  vpc_id      = "${aws_vpc.default.id}"

  tags {
    Name = "${var.name}-controller"
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 6443          # kubernetes api server
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "node" {
  name        = "${var.name}-node"
  description = "security group for kubernetes nodes"
  vpc_id      = "${aws_vpc.default.id}"

  tags {
    Name = "${var.name}-node"
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_key_pair" "auth" {
  key_name   = "${var.name}"
  public_key = "${file(var.public_key_path)}"
}

data "template_file" "controller_config" {
  template = "${file("${path.module}/user_data.tpl")}"
  count    = "${var.controller_count}"

  vars {
    hostname = "controller-${count.index}"
  }
}

resource "aws_instance" "controller" {
  count                  = "${var.controller_count}"
  instance_type          = "${var.controller_instance_type}"
  ami                    = "${coalesce(var.controller_ami_id, data.aws_ami.ami.id)}"
  vpc_security_group_ids = ["${aws_security_group.controller.id}"]
  subnet_id              = "${aws_subnet.default.id}"
  key_name               = "${aws_key_pair.auth.id}"
  user_data              = "${element(data.template_file.controller_config.*.rendered, count.index)}"

  tags {
    Name            = "${var.name}"
    Role            = "controller"
    ControllerIndex = "${count.index}"
  }

  root_block_device {
    delete_on_termination = true
  }

  connection {
    user = "${var.ssh_user_name}"
  }

  provisioner "file" {
    source      = "../../converge"
    destination = "/home/${var.ssh_user_name}/converge"
  }

  provisioner "converge" {
    hcl = [
      "converge/cfssl.hcl",
      "converge/docker.hcl",
    ]

    download_binary = true
    version         = "${var.converge_version}"
    prevent_sudo    = false
  }
}

data "template_file" "node_config" {
  template = "${file("${path.module}/user_data.tpl")}"
  count    = "${var.node_count}"

  vars {
    hostname = "node-${count.index}"
  }
}

resource "aws_instance" "node" {
  count                  = "${var.node_count}"
  instance_type          = "${var.node_instance_type}"
  ami                    = "${coalesce(var.node_ami_id, data.aws_ami.ami.id)}"
  vpc_security_group_ids = ["${aws_security_group.node.id}"]
  subnet_id              = "${aws_subnet.default.id}"
  key_name               = "${aws_key_pair.auth.id}"
  user_data              = "${element(data.template_file.node_config.*.rendered, count.index)}"

  tags {
    Name      = "${var.name}"
    Role      = "node"
    NodeIndex = "${count.index}"
  }

  root_block_device {
    delete_on_termination = true
  }

  connection {
    user = "${var.ssh_user_name}"
  }

  provisioner "file" {
    source      = "../../converge"
    destination = "/home/${var.ssh_user_name}/converge"
  }

  provisioner "converge" {
    hcl = [
      "converge/cfssl.hcl",
      "converge/docker.hcl",
    ]

    download_binary = true
    version         = "${var.converge_version}"
    prevent_sudo    = false
  }
}

resource "null_resource" "controller_bootstrap" {
  triggers {
    controller_bootstrap = "${aws_instance.controller.0.id}"
  }

  connection {
    user = "${var.ssh_user_name}"
    host = "${aws_instance.controller.0.public_ip}"
  }

  provisioner "converge" {
    params = {
      internal-ip = "${aws_instance.controller.0.private_ip}"
    }

    hcl = [
      "converge/generate-ca.hcl",
    ]

    download_binary = true
    version         = "${var.converge_version}"
    prevent_sudo    = false
  }
}

resource "null_resource" "controllers" {
  count      = "${var.controller_count}"
  depends_on = ["null_resource.controller_bootstrap"]

  triggers {
    controller_ids = "${join(",", aws_instance.controller.*.id)}"
  }

  connection {
    user = "${var.ssh_user_name}"
    host = "${element(aws_instance.controller.*.public_ip, count.index)}"
  }

  provisioner "converge" {
    params = {
      internal-ip          = "${element(aws_instance.controller.*.private_ip, count.index)}"
      ca-url               = "http://${aws_instance.controller.0.private_ip}:9090/ca.tar.gz"
      etcd-node-name       = "${format("etcd-%d", count.index)}"
      etcd-initial-cluster = "${join(",", formatlist("etcd-%s=https://%s:2380", aws_instance.controller.*.tags.ControllerIndex, aws_instance.controller.*.private_ip))}"
      etcd-servers         = "${join(",", formatlist("https://%s:2379", aws_instance.controller.*.private_ip))}"
      kubelet-token        = "${var.kubelet_token}"
      admin-token          = "${var.admin_token}"
      scheduler-token      = "${var.scheduler_token}"
      hosts                = "127.0.0.1,localhost,${format("controller-%d", count.index)},${element(aws_instance.controller.*.private_ip, count.index)},${element(aws_instance.controller.*.public_ip, count.index)}"
    }

    hcl = [
      "converge/generate-cert.hcl",
      "converge/etcd.hcl",
      "converge/kubernetes-controller.hcl",
    ]

    download_binary = true
    version         = "${var.converge_version}"
    prevent_sudo    = false
  }
}

resource "null_resource" "nodes" {
  count      = "${var.node_count}"
  depends_on = ["null_resource.controller_bootstrap"]

  triggers {
    node_ids = "${join(",", aws_instance.node.*.id)}"
  }

  connection {
    user = "${var.ssh_user_name}"
    host = "${element(aws_instance.node.*.public_ip, count.index)}"
  }

  provisioner "converge" {
    params = {
      internal-ip   = "${element(aws_instance.node.*.private_ip, count.index)}"
      ca-url        = "http://${aws_instance.controller.0.private_ip}:9090/ca.tar.gz"
      controller-ip = "${aws_instance.controller.0.private_ip}"
      api-servers   = "${join(",", formatlist("https://%s:6443", aws_instance.controller.*.private_ip))}"

      # peers         = "${join(" ", aws_instance.node.*.private_ip)}"
      peers         = "${replace(join(" ", aws_instance.node.*.private_ip), "/${element(aws_instance.node.*.private_ip, count.index)}\\s+/", "")}"
      kubelet-token = "${var.kubelet_token}"
    }

    hcl = [
      "converge/generate-cert.hcl",
      "converge/cni.hcl",
      "converge/weave.hcl",
      "converge/kubernetes-node.hcl",
    ]

    download_binary = true
    version         = "${var.converge_version}"
    prevent_sudo    = false
  }
}

output "controller_ips" {
  value = ["${aws_instance.controller.*.public_ip}"]
}

output "node_ips" {
  value = ["${aws_instance.node.*.public_ip}"]
}

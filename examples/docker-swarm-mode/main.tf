variable "name" {
  default = "docker-swarm"
}

variable "region" {
  default = "us-east-1"
}

variable "public_key_path" {
  default = "~/.ssh/id_rsa.pub"
}

variable "amis" {
  # centos 7
  default = {
    ap-south-1     = "ami-95cda6fa"
    ap-southeast-1 = "ami-f068a193"
    ap-southeast-2 = "ami-fedafc9d"
    ap-northeast-1 = "ami-eec1c380"
    ap-northeast-2 = "ami-c74789a9"
    eu-central-1   = "ami-9bf712f4"
    eu-west-1      = "ami-7abd0209"
    sa-east-1      = "ami-26b93b4a"
    us-east-1      = "ami-6d1c2007"
    us-west-1      = "ami-af4333cf"
    us-west-2      = "ami-d2c924b2"
  }
}

variable "ssh_user_name" {
  default = "centos"
}

variable "vpc_cidr_block" {
  default = "10.0.0.0/16"
}

variable "bucket-prefix" {
  default = "asteris-llc"
}

variable "worker-count" {
  default = 2
}

provider "aws" {
  region = "${var.region}"
}

resource "aws_s3_bucket" "token" {
  bucket = "${var.bucket-prefix}-${var.name}"
  acl = "private"
  force_destroy = true

  tags {
    Name = "${var.bucket-prefix}-${var.name}"
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
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  tags {
    Name = "${var.name}"
  }
}

resource "aws_security_group" "default" {
  name = "${var.name}-default"
  description = "default security group for elk-test"
  vpc_id      = "${aws_vpc.default.id}"
  tags {
    Name = "${var.name}-default"
  }

  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
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


resource "aws_iam_role" "node-role" {
  name = "${var.name}-node-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "ec2.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "node-policy" {
  name = "${var.name}-node-policy"
  role = "${aws_iam_role.node-role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:ListBucket"],
      "Resource": ["arn:aws:s3:::${var.bucket-prefix}-${var.name}"]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:DeleteObject"
      ],
      "Resource": ["arn:aws:s3:::${var.bucket-prefix}-${var.name}/*"]
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "node-profile" {
  name = "${var.name}-node-profile"
  roles = ["${aws_iam_role.node-role.name}"]
}

resource "aws_instance" "manager" {
  depends_on = ["aws_s3_bucket.token"]
  ami = "${lookup(var.amis, var.region)}"
  instance_type = "m3.medium"
  vpc_security_group_ids = ["${aws_security_group.default.id}"]
  subnet_id = "${aws_subnet.default.id}"
  key_name = "${aws_key_pair.auth.id}"
  iam_instance_profile = "${aws_iam_instance_profile.node-profile.name}"
  tags {
    Name = "${var.name}"
    Role = "manager"
  }

  root_block_device {
    delete_on_termination = true
  }

  connection {
    user = "${var.ssh_user_name}"
  }

  provisioner "file" {
    source = "./converge-hcl"
    destination = "/home/${var.ssh_user_name}/converge-hcl"
  }

  provisioner "converge" {
    params = {
      docker-group-user-name = "${var.ssh_user_name}"
      swarm-manager-ip = "${aws_instance.manager.private_ip}"
      swarm-token-bucket = "${var.bucket-prefix}-${var.name}"
    }
    modules = [
      "converge-hcl/main.hcl",
      "converge-hcl/manager.hcl"
    ]
    download_binary = true
    prevent_sudo = false
  }
}

resource "aws_instance" "worker" {
  count = "${var.worker-count}"
  ami = "${lookup(var.amis, var.region)}"
  instance_type = "m3.medium"
  vpc_security_group_ids = ["${aws_security_group.default.id}"]
  subnet_id = "${aws_subnet.default.id}"
  key_name = "${aws_key_pair.auth.id}"
  iam_instance_profile = "${aws_iam_instance_profile.node-profile.name}"
  tags {
    Name = "${var.name}"
    Role = "worker"
  }

  root_block_device {
    delete_on_termination = true
  }

  connection {
    user = "${var.ssh_user_name}"
  }

  provisioner "file" {
    source = "./converge-hcl"
    destination = "/home/${var.ssh_user_name}/converge-hcl"
  }

  provisioner "converge" {
    params = {
      docker-group-user-name = "${var.ssh_user_name}",
    }
    modules = [
      "converge-hcl/main.hcl"
    ]
    download_binary = true
    prevent_sudo = false
  }
}

resource "null_resource" "worker-join" {
  count = "${var.worker-count}"
  depends_on = ["aws_instance.manager"]
  triggers {
    workers = "${join(",", aws_instance.worker.*.id)}"
  }

  connection {
    user = "${var.ssh_user_name}"
    host = "${element(aws_instance.worker.*.public_ip, count.index)}"
  }

  provisioner "converge" {
    params = {
      swarm-manager-ip = "${aws_instance.manager.private_ip}"
      swarm-token-bucket = "${var.bucket-prefix}-${var.name}"
    }
    modules = [
      "converge-hcl/worker.hcl"
    ]
    download_binary = true
    prevent_sudo = false
  }
}

output "manager-ip" {
  value = "${aws_instance.manager.public_ip}"
}

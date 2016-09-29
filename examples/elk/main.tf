variable "name" {
  default = "elk-test"
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

provider "aws" {
  region = "${var.region}"
}

resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"
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
  name = "default-elk-test"
  description = "default security group for elk-test"
  vpc_id      = "${aws_vpc.default.id}"
  tags {
    Name = "${var.name}"
  }

  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 5601
    to_port = 5601
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
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

resource "aws_instance" "elk" {
  ami = "${lookup(var.amis, var.region)}"
  instance_type = "m3.medium"
  vpc_security_group_ids = ["${aws_security_group.default.id}"]
  subnet_id = "${aws_subnet.default.id}"
  key_name = "${aws_key_pair.auth.id}"
  tags {
    Name = "${var.name}"
  }

  root_block_device {
    delete_on_termination = true
  }

  connection {
    user = "${var.ssh_user_name}"
  }

  provisioner "file" {
    source = "./converge"
    destination = "/home/${var.ssh_user_name}/converge"
  }

  provisioner "converge" {
    params = {
      user-name = "${var.ssh_user_name}"
    }
    hcl = [
      "converge/elk.hcl"
    ]
    download_binary = true
    prevent_sudo = false
  }
}

output "ip" {
  value = "${aws_instance.elk.public_ip}"
}

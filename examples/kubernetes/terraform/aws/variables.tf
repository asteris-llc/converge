variable "region" {
  default = "us-east-1"
}

variable "name" {
  default = "converge-kubernetes"
}

variable "ssh_user_name" {
  default = "ubuntu"
}

variable "controller_ami_id" {
  default     = ""
  description = "the ami id for controller instances. If left blank, it will use the most recent version of the official ubuntu 16.04 (xenial) ami"
}

variable "node_ami_id" {
  default     = ""
  description = "the ami id for node instances. If left blank, it will use the most recent version of the official ubuntu 16.04 (xenial) ami"
}

variable "public_key_path" {
  default = "~/.ssh/id_rsa.pub"
}

variable "controller_instance_type" {
  default     = "m3.medium"
  description = "the ec2 instance type for kubernetes controller nodes"
}

variable "node_instance_type" {
  default     = "m3.medium"
  description = "the ec2 instance type for kubernetes nodes"
}

variable "controller_count" {
  default     = 3
  description = "the number of kubernetes controllers"
}

variable "node_count" {
  default     = 2
  description = "the number of kubernetes nodes"
}

variable "vpc_cidr_block" {
  default = "10.0.0.0/16"
}

variable "converge_version" {
  default = "0.4.0-rc1"
}

variable "kubelet_token" {
  default = "chAng3m3"
}

variable "admin_token" {
  default = "chAng3m3"
}

variable "scheduler_token" {
  default = "chAng3m3"
}

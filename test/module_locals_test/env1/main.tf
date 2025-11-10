# Module記述のテスト
module "vpc" {
  source = "./modules/vpc"

  vpc_cidr = "10.0.0.0/16"
  environment = "dev"
}

# Locals記述のテスト
locals {
  common_tags = {
    Environment = "dev"
    Project     = "test"
  }

  vpc_cidr = "10.0.0.0/16"
}

# Variable記述のテスト
variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

# Output記述のテスト
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

# Data記述のテスト
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}

# 通常のresource記述（比較用）
resource "aws_instance" "test" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  tags = local.common_tags
}
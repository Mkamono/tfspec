# Module記述のテスト（環境別設定）
module "vpc" {
  source = "./modules/vpc"

  vpc_cidr = "10.1.0.0/16"  # 異なるCIDR
  environment = "prod"      # 異なる環境
}

# Locals記述のテスト（環境別設定）
locals {
  common_tags = {
    Environment = "prod"    # 異なる環境
    Project     = "test"
  }

  vpc_cidr = "10.1.0.0/16" # 異なるCIDR

  # Boolean locals のテスト（env1と異なる値）
  enable_monitoring = false  # env1はtrue
  enable_backup = true       # env1はfalse
}

# Variable記述のテスト（追加属性）
variable "instance_type" {
  description = "EC2 instance type for production"  # 異なるdescription
  type        = string
  default     = "t3.small"  # 異なるdefault値
}

# 追加variable（env1にはない）
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

# Output記述のテスト（同じ）
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

# 追加output（env1にはない）
output "vpc_cidr" {
  description = "VPC CIDR Block"
  value       = local.vpc_cidr
}

# Data記述のテスト（異なるfilter）
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]  # 異なるUbuntu版
  }
}

# 通常のresource記述（比較用）
resource "aws_instance" "test" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  tags = local.common_tags
}
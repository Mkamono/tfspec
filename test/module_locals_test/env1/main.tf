# Module記述のテスト
module "vpc" {
  source = "./modules/vpc"

  vpc_cidr    = "10.0.0.0/16"
  environment = "dev"
}

# Locals記述のテスト
locals {
  common_tags = {
    Environment = "dev"
    Project     = "test"
  }

  vpc_cidr = "10.0.0.0/16"

  # Boolean locals のテスト
  enable_monitoring = true
  enable_backup     = false

  # Object locals のテスト
  database_config = {
    engine                  = "mysql"
    engine_version          = "8.0"
    multi_az                = true
    backup_retention_period = 7
  }

  # List locals のテスト
  allowed_cidr_blocks = [
    "10.0.0.0/8",
    "172.16.0.0/12"
  ]

  # HCL関数を含むローカル変数
  name_with_length = length(var.instance_type)
  merged_tags      = merge(local.common_tags, { "AdditionalTag" = "value" })
  name_prefix      = "app-${var.instance_type}"
  file_content     = file("${path.module}/config.txt")

  # env1のみに存在するオブジェクト
  dev_only_config = {
    debug_mode = true
    log_level  = "debug"
  }

  long_object = {
    level1 = {
      level2 = {
        level3_1 = {
          key             = "deep_value"
          another_key     = "another_value"
          yet_another_key = "yet_another_value"
          deep_nested_key = "deep_nested_value"
        }
        level3_2 = {
          key             = "deep_value"
          another_key     = "another_value"
          yet_another_key = "yet_another_value"
          deep_nested_key = "deep_nested_value"
        }
        level3_3 = {
          key             = "deep_value"
          another_key     = "another_value"
          yet_another_key = "yet_another_value"
          deep_nested_key = "deep_nested_value"
        }
      }
    }
  }

  concat_test = concat([
    "a", # a
    "b", # b
    ], [
    "c",
    "d",
  ])

  object_test = {
    # comment inside object
    name = "test_object"
    # another comment
    numbers = [1, 2, 3, 4, 5]
    # where is the comment?
    nested = {
      key1 = "value1"
      key2 = "value2"
    }
  }
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

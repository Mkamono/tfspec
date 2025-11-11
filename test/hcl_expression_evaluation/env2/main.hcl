# 変数定義
variable "instance_type" {
  default = "t3.medium"  # env2は異なるインスタンスタイプ
}

variable "environment" {
  default = "env2"
}

# 変数参照を使ったリソース
resource "aws_instance" "web" {
  instance_type = var.instance_type
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name        = "web-${var.environment}"
    Environment = var.environment
  }

  # HCL関数を使った属性（env2はfile）
  user_data = file("${path.module}/user-data.sh")
}

# dataリソースでのHCL式評価
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-${var.environment}-*"]
  }

  owners = ["099720109477"]
}

# 複雑な式評価
resource "aws_security_group" "web" {
  name        = "web-sg-${var.environment}"
  description = templatefile("${path.module}/sg-description.tpl", {
    env = var.environment
  })

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# locals での式評価
locals {
  common_tags = {
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "tfspec-test"
  }

  instance_name = "instance-${var.environment}-${formatdate("YYYY-MM-DD", timestamp())}"
}

# output での式評価
output "instance_id" {
  value       = aws_instance.web.id
  description = "The ID of the web instance in ${var.environment}"
}

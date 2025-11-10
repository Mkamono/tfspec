resource "aws_security_group" "complex" {
  name = "complex-sg-dev"

  # 多数のingressルール
  ingress {
    from_port = 80
    to_port   = 80
    protocol  = "tcp"
    cidr_blocks = ["10.0.1.0/24"]
  }

  ingress {
    from_port = 443
    to_port   = 443
    protocol  = "tcp"
    cidr_blocks = ["10.0.1.0/24"]
  }

  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port = 8080
    to_port   = 8080
    protocol  = "tcp"
    cidr_blocks = ["10.0.2.0/24"]
  }

  ingress {
    from_port = 3306
    to_port   = 3306
    protocol  = "tcp"
    cidr_blocks = ["10.0.3.0/24"]
  }

  ingress {
    from_port = 6379
    to_port   = 6379
    protocol  = "tcp"
    cidr_blocks = ["10.0.4.0/24"]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Environment = "dev"
  }
}

resource "aws_launch_configuration" "complex" {
  name          = "complex-lc-dev"
  image_id      = "ami-12345678"
  instance_type = "t3.small"

  # 多数のEBSブロックデバイス
  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 10
    volume_type = "gp2"
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 20
    volume_type = "gp3"
  }

  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 30
    volume_type = "io1"
    iops = 100
  }

  ebs_block_device {
    device_name = "/dev/sde"
    volume_size = 40
    volume_type = "gp2"
  }

  lifecycle {
    create_before_destroy = true
  }
}
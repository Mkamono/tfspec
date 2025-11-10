resource "aws_security_group" "complex" {
  name = "complex-sg-staging"

  # staging環境では異なるingressルール
  ingress {
    from_port = 80
    to_port   = 80
    protocol  = "tcp"
    cidr_blocks = ["10.0.1.0/24", "10.0.5.0/24"]  # 追加のCIDR
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

  # staging環境では8080ポートが異なる設定
  ingress {
    from_port = 8080
    to_port   = 8080
    protocol  = "tcp"
    cidr_blocks = ["10.0.2.0/24", "10.0.6.0/24"]  # 追加のCIDR
  }

  ingress {
    from_port = 3306
    to_port   = 3306
    protocol  = "tcp"
    cidr_blocks = ["10.0.3.0/24"]
  }

  # staging環境ではRedisポートが異なる
  ingress {
    from_port = 6379
    to_port   = 6379
    protocol  = "tcp"
    cidr_blocks = ["10.0.4.0/24", "10.0.7.0/24"]  # 追加のCIDR
  }

  # 追加のポート
  ingress {
    from_port = 9200
    to_port   = 9200
    protocol  = "tcp"
    cidr_blocks = ["10.0.8.0/24"]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Environment = "staging"
  }
}

resource "aws_launch_configuration" "complex" {
  name          = "complex-lc-staging"
  image_id      = "ami-87654321"
  instance_type = "t3.medium"

  # staging環境では異なるEBSブロックデバイス
  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 15  # 異なるサイズ
    volume_type = "gp3"
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 25  # 異なるサイズ
    volume_type = "gp3"
  }

  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 35  # 異なるサイズ
    volume_type = "io1"
    iops = 150
  }

  ebs_block_device {
    device_name = "/dev/sde"
    volume_size = 45  # 異なるサイズ
    volume_type = "gp3"
  }

  # staging環境では追加のブロックデバイス
  ebs_block_device {
    device_name = "/dev/sdf"
    volume_size = 50
    volume_type = "gp3"
  }

  lifecycle {
    create_before_destroy = true
  }
}
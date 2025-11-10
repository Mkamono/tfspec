resource "aws_security_group" "complex" {
  name = "complex-sg-production"

  # 本番環境では最も厳格なingressルール
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

  # 本番環境ではSSHアクセスがより制限的
  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"
    cidr_blocks = ["10.0.0.0/24"]  # より狭い範囲
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

  # 本番環境では追加の監視ポート
  ingress {
    from_port = 9100
    to_port   = 9100
    protocol  = "tcp"
    cidr_blocks = ["10.0.9.0/24"]
  }

  ingress {
    from_port = 3000
    to_port   = 3000
    protocol  = "tcp"
    cidr_blocks = ["10.0.10.0/24"]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Environment = "production"
  }
}

resource "aws_launch_configuration" "complex" {
  name          = "complex-lc-production"
  image_id      = "ami-production"
  instance_type = "t3.large"

  # 本番環境では高性能なEBSブロックデバイス
  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 50
    volume_type = "gp3"
    throughput = 250
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 100
    volume_type = "gp3"
    throughput = 500
  }

  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 200
    volume_type = "io1"
    iops = 1000
  }

  ebs_block_device {
    device_name = "/dev/sde"
    volume_size = 100
    volume_type = "gp3"
    throughput = 500
  }

  ebs_block_device {
    device_name = "/dev/sdf"
    volume_size = 500
    volume_type = "gp3"
    throughput = 1000
  }

  # 本番環境では追加のログ用ボリューム
  ebs_block_device {
    device_name = "/dev/sdg"
    volume_size = 1000
    volume_type = "gp3"
    throughput = 1000
  }

  lifecycle {
    create_before_destroy = true
  }
}
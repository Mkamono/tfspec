resource "aws_instance" "web" {
  ami           = "ami-0abcdef1234567890"
  instance_type = "t3.medium"  # 本番環境はより大きなインスタンス

  tags = {
    Name        = "web-server"
    Environment = "production"
  }
}

resource "aws_instance" "worker" {
  ami           = "ami-0abcdef1234567890"
  instance_type = "t3.small"   # 本番環境のworkerも大きめ

  tags = {
    Name = "worker"
    Type = "background"
  }
}
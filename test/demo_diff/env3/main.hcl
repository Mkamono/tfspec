resource "aws_instance" "demo" {
  instance_type = "t3.large"  # env3では大きなインスタンス
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "demo-server"
    Environment = "production"  # 本番環境では異なるタグ
  }
}
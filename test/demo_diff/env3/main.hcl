resource "aws_instance" "web" {
  instance_type = "t3.large"
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "web-server"
    Environment = "production"  # 意図的にenv3ではなくproductionに設定
  }
}
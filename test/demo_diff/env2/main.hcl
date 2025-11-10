resource "aws_instance" "demo" {
  instance_type = "t3.medium"  # env2では中程度のインスタンス
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "demo-server"
    Environment = "env2"
  }
}
resource "aws_instance" "web" {
  instance_type = "t3.small"
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "web-server"
    Environment = "env1"
  }
}
resource "aws_instance" "demo" {
  instance_type = "t3.micro"
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "demo-instance-env1"
    Environment = "env1"
  }
}
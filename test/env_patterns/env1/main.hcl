resource "aws_instance" "web" {
  instance_type = "t3.small"
  ami           = "ami-0abcdef1234567890"
  monitoring    = false

  tags = {
    Name = "web-server"
    Environment = "env1"
    Backup = "false"
  }
}
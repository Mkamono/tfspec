resource "aws_instance" "web" {
  instance_type = "t3.large"
  ami           = "ami-0abcdef1234567890"
  monitoring    = true

  tags = {
    Name = "web-server"
    Environment = "env3"
    Backup = "true"
  }
}
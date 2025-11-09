resource "aws_instance" "web" {
  instance_type = "t3.medium"
  ami           = "ami-0abcdef1234567890"

  root_block_device {
    volume_size = 50
    volume_type = "gp3"
  }

  tags = {
    Name        = "web-server"
    Environment = "env2"
    Backup      = "true"
  }
}
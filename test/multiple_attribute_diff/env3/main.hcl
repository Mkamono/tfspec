resource "aws_instance" "web" {
  instance_type = "t3.large"
  ami           = "ami-0abcdef1234567890"

  root_block_device {
    volume_size = 100
    volume_type = "gp3"
  }

  tags = {
    Name        = "web-server"
    Environment = "env3"
    Backup      = "true"
  }
}
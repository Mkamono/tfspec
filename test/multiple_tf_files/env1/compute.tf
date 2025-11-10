resource "aws_instance" "web" {
  ami           = "ami-0abcdef1234567890"
  instance_type = "t3.small"

  tags = {
    Name        = "web-server"
    Environment = "development"
  }
}

resource "aws_instance" "worker" {
  ami           = "ami-0abcdef1234567890"
  instance_type = "t3.micro"

  tags = {
    Name = "worker"
    Type = "background"
  }
}
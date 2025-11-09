resource "aws_instance" "web" {
  instance_type = "t3.small"
  tags = {
    Environment = "dev"
  }
}

resource "aws_instance" "db" {
  instance_type = "t3.micro"
}

resource "aws_instance" "cache" {
  instance_type = "t3.nano"
}

resource "aws_security_group" "web" {
  ingress {
    from_port = 80
    to_port   = 80
  }
}
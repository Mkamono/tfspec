resource "aws_instance" "web" {
  instance_type = "t3.medium"
  tags = {
    Environment = "staging"
  }
}

resource "aws_instance" "db" {
  instance_type = "t3.small"
}

resource "aws_instance" "cache" {
  instance_type = "t3.micro"
}

resource "aws_security_group" "web" {
  ingress {
    from_port = 80
    to_port   = 80
  }
  ingress {
    from_port = 443
    to_port   = 443
  }
}
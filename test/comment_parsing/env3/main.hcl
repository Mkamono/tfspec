resource "aws_instance" "web" {
  instance_type = "t3.large"
  tags = {
    Environment = "production"
  }
}

resource "aws_instance" "db" {
  instance_type = "t3.medium"
}

resource "aws_instance" "cache" {
  instance_type = "t3.small"
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
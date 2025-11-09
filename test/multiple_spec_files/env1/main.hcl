resource "aws_instance" "web" {
  instance_type = "t3.small"
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "web-server"
    Environment = "env1"
  }
}

resource "aws_security_group" "web" {
  name_prefix = "web-sg"
  description = "Security group for web servers"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "web-security-group"
    Environment = "env1"
  }
}

resource "aws_cloudwatch_log_group" "app" {
  name              = "/aws/app/web"
  retention_in_days = 7

  tags = {
    Environment = "env1"
  }
}
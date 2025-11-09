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
    AllowedPorts = "80"
  }
}
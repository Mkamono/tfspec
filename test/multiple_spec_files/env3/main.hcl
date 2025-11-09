resource "aws_instance" "web" {
  instance_type = "t3.large"
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "web-server"
    Environment = "env3"
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

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "web-security-group"
    Environment = "env3"
  }
}

resource "aws_cloudwatch_log_group" "app" {
  name              = "/aws/app/web"
  retention_in_days = 365

  tags = {
    Environment = "env3"
  }
}

resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  alarm_name          = "high-cpu-usage"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "85"
  alarm_description   = "This metric monitors ec2 cpu utilization"

  dimensions = {
    InstanceId = aws_instance.web.id
  }
}
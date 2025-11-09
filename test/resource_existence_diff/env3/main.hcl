resource "aws_instance" "web" {
  instance_type = "t3.small"
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "web-server"
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
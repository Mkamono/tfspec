resource "aws_instance" "demo" {
  instance_type = "t3.large"
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "demo-instance-env3"
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
}
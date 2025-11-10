resource "aws_instance" "web" {
  instance_type = "t3.large"
  tags = {
    Environment = "staging"
    Owner = "team1"
  }
}

# 異なる重複パターン
resource "aws_instance" "web" {
  instance_type = "t3.xlarge"
  tags = {
    Environment = "staging_second"
    Owner = "team3"
  }
}
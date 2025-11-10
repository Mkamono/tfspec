# env3では重複しない正常な定義
resource "aws_instance" "web" {
  instance_type = "t3.2xlarge"
  tags = {
    Environment = "production"
    Owner = "team1"
  }
}

resource "aws_instance" "web_backup" {
  instance_type = "t3.large"
  tags = {
    Environment = "production"
    Owner = "team1"
    Purpose = "backup"
  }
}
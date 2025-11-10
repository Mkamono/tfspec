resource "aws_instance" "web" {
  instance_type = "t3.small"
  tags = {
    Environment = "dev"
    Owner = "team1"
  }
}

# 同じ名前のリソースを重複定義（HCL的にはエラーだが、パーサーの動作をテスト）
resource "aws_instance" "web" {
  instance_type = "t3.medium"
  tags = {
    Environment = "dev_duplicate"
    Owner = "team2"
  }
}
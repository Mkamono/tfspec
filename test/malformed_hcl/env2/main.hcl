# 構文エラー: 括弧が閉じられていない
resource "aws_instance" "web" {
  instance_type = "t3.medium"
  tags = {
    Environment = "staging"
  # この括弧が閉じられていない
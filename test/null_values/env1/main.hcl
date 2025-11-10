resource "aws_instance" "web" {
  instance_type = "t3.small"

  # 明示的にnullに設定
  user_data = null
  key_name  = null

  tags = {
    Environment = "dev"
    # null値のタグ
    NullTag = null
  }
}

resource "aws_instance" "db" {
  instance_type = "t3.micro"
  # user_dataやkey_nameを定義しない（undefined）

  tags = {
    Environment = "dev"
    Role = "database"
  }
}
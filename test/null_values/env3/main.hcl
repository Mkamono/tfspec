resource "aws_instance" "web" {
  instance_type = "t3.large"

  # 本番では値を設定
  user_data = "#!/bin/bash\necho 'production'"
  # key_nameは未定義のまま

  tags = {
    Environment = "production"
    NullTag = "actually_has_value"  # env1ではnullだった
  }
}

resource "aws_instance" "db" {
  instance_type = "t3.medium"

  # nullを明示的に設定
  user_data = null
  # key_nameは未定義

  tags = {
    Environment = "production"
    Role = "database"
  }
}
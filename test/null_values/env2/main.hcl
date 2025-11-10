resource "aws_instance" "web" {
  instance_type = "t3.medium"

  # env1とは異なる値を設定
  user_data = "#!/bin/bash\necho 'staging'"
  key_name  = "staging-key"

  tags = {
    Environment = "staging"
    # NullTagは定義しない（env1ではnullだった）
  }
}

resource "aws_instance" "db" {
  instance_type = "t3.small"

  # こちらでは値を設定
  user_data = "#!/bin/bash\necho 'db staging'"
  key_name  = "db-staging-key"

  tags = {
    Environment = "staging"
    Role = "database"
  }
}
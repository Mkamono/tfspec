resource "aws_instance" "web" {
  instance_type = "t3.large"  # tfspec(env="nonexistent_env", reason="存在しない環境名でのテスト")
  instance_type = "t3.medium" # tfspec(env="env2", reason="") # 空のreason
  instance_type = "t3.small"  # tfspec(env=["env1"]) # reasonパラメータなし

  tags = {
    Environment = "env3" # tfspec(invalid_syntax_here)
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}
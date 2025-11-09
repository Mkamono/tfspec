resource "aws_instance" "web" {
  instance_type = "t3.large"  # tfspec(env="env3", reason="本番環境の高負荷要件")
  instance_type = "t3.medium" # tfspec(env="env2", reason="ステージング環境の中間性能")
  instance_type = "t3.small"  # tfspec(env="env1", reason="開発環境のコスト最適化")

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}
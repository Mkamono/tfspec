resource "aws_instance" "web" {
  instance_type = "t3.large"  # tfspec(env="env3", reason="本番環境の高負荷要件")
  instance_type = "t3.medium" # tfspec(env=["env2"], reason="ステージング環境の性能要件")
  instance_type = "t3.small"  # tfspec(env="env1", reason="開発環境のコスト最適化")
  monitoring    = true  # tfspec(env=["env2", "env3"], reason="本番・ステージング環境での監視必須")
  monitoring    = false # tfspec(env="env1", reason="開発環境では監視不要")

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
    Backup = "true"  # tfspec(env=["env2", "env3"], reason="本番・ステージング環境でのデータ保護")
    Backup = "false" # tfspec(env="env1", reason="開発環境はバックアップ不要")
  }
}
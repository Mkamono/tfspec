resource "aws_instance" "web" {
  instance_type = "t3.large"  # tfspec(env="env3", reason="本番環境の高負荷要件")
  instance_type = "t3.medium" # tfspec(env="env2", reason="ステージング環境の中間性能")
  instance_type = "t3.small"  # tfspec(env="env1", reason="開発環境のコスト最適化")

  root_block_device {
    volume_size = 100 # tfspec(env="env3", reason="本番データ容量要件")
    volume_size = 50  # tfspec(env="env2", reason="ステージング用データ容量")
    volume_size = 20  # tfspec(env="env1", reason="開発環境のコスト最適化")
  }

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
    Backup      = "true"  # tfspec(env=["env2", "env3"], reason="本番・ステージング環境のデータ保護")
    Backup      = "false" # tfspec(env="env1", reason="開発環境はバックアップ不要")
  }
}
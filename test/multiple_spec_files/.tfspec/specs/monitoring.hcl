resource "aws_cloudwatch_log_group" "app" {
  retention_in_days = 365 # tfspec(env="env3", reason="本番環境でのコンプライアンス要件")
  retention_in_days = 30  # tfspec(env="env2", reason="ステージング環境の標準保持期間")
  retention_in_days = 7   # tfspec(env="env1", reason="開発環境のコスト最適化")

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}

# CloudWatchアラームはenv3にのみ存在
resource "aws_cloudwatch_metric_alarm" "high_cpu" {} # tfspec(env="env3", reason="本番環境でのSLA保証のための必須監視")
resource "aws_instance" "web" {
  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}

# このアラームはenv3環境にのみ存在する
resource "aws_cloudwatch_metric_alarm" "high_cpu" {} # tfspec(env="env3", reason="本番環境でのSLA保証のための必須監視")
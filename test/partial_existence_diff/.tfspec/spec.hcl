resource "aws_instance" "web" {
  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}

resource "aws_security_group" "web" {
  # HTTPS はenv2とenv3にのみ存在し、両環境間でcidr_blocksに差分あり
  ingress {
    cidr_blocks = ["172.16.0.0/12"] # tfspec(env="env3", reason="本番環境では限定されたネットワークからのみHTTPSアクセス")
    cidr_blocks = ["0.0.0.0/0"]     # tfspec(env="env2", reason="ステージング環境では全体からHTTPSアクセス可")
  }

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}

# CloudWatchアラームはenv2とenv3にのみ存在し、しきい値と評価期間に差分あり
resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  evaluation_periods = "2"  # tfspec(env="env3", reason="本番環境では迅速な検知のため短い評価期間")
  evaluation_periods = "3"  # tfspec(env="env2", reason="ステージング環境では誤検知を避けるため長めの評価期間")
  threshold          = "85" # tfspec(env="env3", reason="本番環境では厳格なCPU使用率監視")
  threshold          = "75" # tfspec(env="env2", reason="ステージング環境では早期警告のための低いしきい値")
}
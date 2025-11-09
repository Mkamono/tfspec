resource "aws_security_group" "web" {
  # HTTPS はenv2とenv3にのみ存在
  ingress {} # tfspec(env=["env2", "env3"], reason="SSL終端のため本番・ステージング環境でHTTPS必須")

  ingress {
    cidr_blocks = ["172.16.0.0/12"] # tfspec(env="env3", reason="本番環境では限定されたネットワークからのみSSH接続")
    cidr_blocks = ["10.0.0.0/8"]    # tfspec(env=["env1", "env2"], reason="開発・ステージング環境では社内ネットワークからSSH接続")
  }

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}
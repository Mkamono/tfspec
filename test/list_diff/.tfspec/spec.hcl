resource "aws_security_group" "web" {
  # HTTPS はenv2とenv3にのみ存在
  ingress {} # tfspec(env=["env2", "env3"], reason="SSL/TLS通信のためHTTPS必須")

  # 管理用ポートはenv3にのみ存在
  ingress {} # tfspec(env="env3", reason="本番環境での管理インターフェースアクセス")

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
    AllowedPorts = "80,443,8080" # tfspec(env="env3", reason="全ポート許可")
    AllowedPorts = "80,443"      # tfspec(env="env2", reason="HTTP/HTTPS許可")
    AllowedPorts = "80"          # tfspec(env="env1", reason="HTTPのみ許可")
  }
}
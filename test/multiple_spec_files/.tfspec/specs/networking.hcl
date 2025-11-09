resource "aws_security_group" "web" {
  # HTTPS はenv2とenv3にのみ存在
  ingress {} # tfspec(env=["env2", "env3"], reason="SSL/TLS通信のためHTTPS必須")

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}
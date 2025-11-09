resource "aws_instance" "web" {
  instance_type = "t3.large" # tfspec(env="env3", reason="本番環境のパフォーマンス要件")

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}
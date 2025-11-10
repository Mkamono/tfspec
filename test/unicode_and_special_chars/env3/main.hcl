resource "aws_instance" "web_日本語" {
  instance_type = "t3.large"
  tags = {
    "日本語キー" = "本番環境"
    "emoji_🌟" = "✨"
    "special-chars_$" = "different_value!@#"
    Environment = "production"
  }
}

resource "aws_instance" "web-special_$chars" {
  instance_type = "t3.large"
  tags = {
    "日本語キー" = "本番用設定"
    "emoji_🌟" = "💎"
  }
}
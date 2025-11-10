resource "aws_instance" "web_日本語" {
  instance_type = "t3.medium"
  tags = {
    "日本語キー" = "ステージング環境"
    "emoji_🌟" = "🌙"
    "special-chars_$" = "test@#$%^&*()"
    Environment = "staging"
  }
}

resource "aws_instance" "web-special_$chars" {
  instance_type = "t3.small"
  tags = {
    "日本語キー" = "異なる値"
    "emoji_🌟" = "⚡"
  }
}
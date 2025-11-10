resource "aws_instance" "web_日本語" {
  instance_type = "t3.small"
  tags = {
    "日本語キー" = "日本語値"
    "emoji_🌟" = "⭐"
    "special-chars_$" = "test@#$%^&*()"
    Environment = "dev"
  }
}

resource "aws_instance" "web-special_$chars" {
  instance_type = "t3.micro"
  tags = {
    "日本語キー" = "開発環境"
    "emoji_🌟" = "🚀"
  }
}
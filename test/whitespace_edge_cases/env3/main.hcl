resource "aws_instance" "web" {
  instance_type = "t3.large"

  # 本番環境用の整理された空白
  user_data = <<-EOF
#!/bin/bash
echo "Clean production formatting"
echo "Consistent indentation"
echo "No mixed whitespace"
  EOF

  tags = {
    Environment = "production"
    # 整理されたキーと値
    "Key With Spaces" = "Production Value"
    "Leading Tab Key" = "Clean Production Value"
    "Mixed	Spaces　And　Full-Width　Spaces" = "本番環境用値"
  }
}
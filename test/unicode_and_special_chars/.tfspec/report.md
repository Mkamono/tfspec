# Tfspec Check Results

## 🚨 意図されていない差分

| 該当箇所 | env1 | env2 | env3 |
|----------|-------|-------|-------|
| aws_instance.web-special_$chars.instance_type | t3.micro | t3.small | t3.large |
| aws_instance.web-special_$chars.tags.emoji_🌟 | 🚀 | ⚡ | 💎 |
| aws_instance.web_日本語.tags.Environment | dev | staging | production |
| aws_instance.web_日本語.tags.emoji_🌟 | ⭐ | 🌙 | ✨ |
| aws_instance.web_日本語.tags.special-chars_$ | test@#$%^&*() | - | different_value!@# |
| aws_instance.web_日本語.tags.日本語キー | 日本語値 | ステージング環境 | 本番環境 |

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_instance.web-special_$chars.tags.日本語キー | 開発環境 | 異なる値 | 本番用設定 | 特殊文字のテスト |
| aws_instance.web_日本語.instance_type | t3.small | t3.medium | t3.large | Unicode文字のテスト：日本語コメント |


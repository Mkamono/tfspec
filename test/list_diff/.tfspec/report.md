# Tfspec Check Results

## ✅ 意図されていない差分

意図されていない差分は検出されませんでした。

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_security_group.web.ingress[1] | (存在しない) | block_exists | block_exists | - |
| aws_security_group.web.ingress[2] | (存在しない) | (存在しない) | block_exists | - |
| aws_security_group.web.tags.AllowedPorts | 80 | 80,443 | 80,443,8080 | 許可ポート設定の環境別要件 |
| aws_security_group.web.tags.Environment | env1 | env2 | env3 | 環境識別タグの意図的差分 |


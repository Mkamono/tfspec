# Tfspec Check Results

## ✅ 意図されていない差分

意図されていない差分は検出されませんでした。

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_instance.cache.instance_type | t3.nano | t3.micro | t3.small | これも行末コメント |
| aws_instance.db.instance_type | t3.micro | t3.small | t3.medium | 行末コメント |
| aws_instance.web.instance_type | t3.small | t3.medium | t3.large | 複数行コメントのテスト これは2行目のコメント これは3行目のコメント |
| aws_instance.web.tags.Environment | dev | staging | production | 単一行コメント |
| aws_security_group.web.ingress[1] | (存在しない) | block_exists | block_exists | - |


# Tfspec Check Results

## ✅ 意図されていない差分

意図されていない差分は検出されませんでした。

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_cloudwatch_metric_alarm.high_cpu | - | - | true | 本番環境でのSLA保証のための必須監視（他環境では不要） |
| aws_instance.web.tags.Environment | env1 | env2 | env3 | 環境識別タグの意図的差分 |


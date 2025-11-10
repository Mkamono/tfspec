# Tfspec Check Results

## 🚨 意図されていない差分

| 該当箇所 | env1 | env2 | env3 |
|----------|-------|-------|-------|
| aws_cloudwatch_log_group.app.retention_in_days | 7 | 30 | 365 |
| aws_cloudwatch_log_group.app.tags.Environment | env1 | env2 | env3 |
| aws_security_group.web.tags.Environment | env1 | env2 | env3 |

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_cloudwatch_metric_alarm.high_cpu | - | - | true | 本番環境での監視要件（他環境では不要） |
| aws_instance.web.instance_type | t3.small | t3.medium | t3.large | 環境別パフォーマンス要件 |
| aws_instance.web.tags.Environment | env1 | env2 | env3 | 環境識別タグ |
| aws_security_group.web.ingress[1] | - | {<br>&nbsp;&nbsp;cidr_blocks: [["0.0.0.0/0"]],<br>&nbsp;&nbsp;from_port: 443,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 443<br>} | {<br>&nbsp;&nbsp;cidr_blocks: [["0.0.0.0/0"]],<br>&nbsp;&nbsp;from_port: 443,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 443<br>} | SSL/TLS通信要件による意図的差分（インデックス1は2番目のingressブロック） |


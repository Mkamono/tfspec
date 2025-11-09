# Tfspec Check Results

## 🚨 意図されていない差分

| 該当箇所 | env1 | env2 | env3 |
|----------|-------|-------|-------|
| aws_cloudwatch_metric_alarm.high_cpu | false | true | true |
| aws_security_group.web.ingress[1] | - | {<br>&nbsp;&nbsp;cidr_blocks: [["0.0.0.0/0"]],<br>&nbsp;&nbsp;from_port: 443,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 443<br>} | {<br>&nbsp;&nbsp;cidr_blocks: [["172.16.0.0/12"]],<br>&nbsp;&nbsp;from_port: 443,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 443<br>} |
| aws_security_group.web.tags.Environment | env1 | env2 | env3 |

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_instance.web.tags.Environment | env1 | env2 | env3 | Environment tag differences |


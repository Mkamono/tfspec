# Tfspec Check Results

## 🚨 意図されていない差分

| 該当箇所 | env1 | env2 | env3 |
|----------|-------|-------|-------|
| aws_instance.web.instance_type | t3.small | t3.medium | t3.large |
| aws_instance.web.tags.Environment | env1 | env2 | env3 |


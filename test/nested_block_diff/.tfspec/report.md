# Tfspec Check Results

## ✅ 意図されていない差分

意図されていない差分は検出されませんでした。

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_security_group.web.ingress[1].cidr_blocks | [10.0.0.0/8] | [0.0.0.0/0] | [0.0.0.0/0] | - |
| aws_security_group.web.ingress[1].from_port | 22 | 443 | 443 | - |
| aws_security_group.web.ingress[1].to_port | 22 | 443 | 443 | - |
| aws_security_group.web.ingress[2] | - | {<br>&nbsp;&nbsp;cidr_blocks: [["10.0.0.0/8"]],<br>&nbsp;&nbsp;from_port: 22,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 22<br>} | {<br>&nbsp;&nbsp;cidr_blocks: [["172.16.0.0/12"]],<br>&nbsp;&nbsp;from_port: 22,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 22<br>} | 3番目のingress ブロック存在差分（本番環境でのSSH設定の再配置） |
| aws_security_group.web.tags.Environment | env1 | env2 | env3 | 環境識別タグの意図的差分 |


# Tfspec Check Results

## 🚨 意図されていない差分

| リソースタイプ |    リソース名    | 属性パス  |                   ENV 1                   |                ENV 2                |                   ENV 3                   |
|:--------------:|:----------------:|:---------:|:-----------------------------------------:|:-----------------------------------:|:-----------------------------------------:|
|    resource    | aws_instance.web | user_data | filebase64("${path.module}/user-data.sh") | file("${path.module}/user-data.sh") | filebase64("${path.module}/user-data.sh") |
|    variable    |   environment    |  default  |                   env1                    |                env2                 |                   env3                    |
|                |  instance_type   |  default  |                 t3.small                  |              t3.medium              |                 t3.large                  |

## 📝 無視された差分（意図的）

| リソースタイプ |       リソース名       |  属性パス  | ENV 1 | ENV 2 |                                                                     ENV 3                                                                     |             理由              |
|:--------------:|:----------------------:|:----------:|:-----:|:-----:|:---------------------------------------------------------------------------------------------------------------------------------------------:|:-----------------------------:|
|    resource    | aws_security_group.web | ingress[1] |   -   |   -   | {<br>&nbsp;&nbsp;cidr_blocks: [["10.0.0.0/8"]],<br>&nbsp;&nbsp;from_port: 22,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 22<br>} | env3のみの追加ingressは意図的 |


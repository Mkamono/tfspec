# Tfspec Check Results

## 🚨 意図されていない差分

|            該当箇所            |  ENV 1   |   ENV 2   |  ENV 3   |
|:------------------------------:|:--------:|:---------:|:--------:|
| aws_instance.web.instance_type | t3.small | t3.medium | t3.large |

## 📝 無視された差分（意図的）

|             該当箇所              | ENV 1 | ENV 2 | ENV 3 |            理由             |
|:---------------------------------:|:-----:|:-----:|:-----:|:---------------------------:|
| aws_instance.web.tags.Environment | env1  | env2  | env3  | Environment tag differences |


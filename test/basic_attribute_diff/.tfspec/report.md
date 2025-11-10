# Tfspec Check Results

## ✅ 意図されていない差分

意図されていない差分は検出されませんでした。

## 📝 無視された差分（意図的）

|             該当箇所              |  ENV 1   |  ENV 2   |  ENV 3   |                     理由                     |
|:---------------------------------:|:--------:|:--------:|:--------:|:--------------------------------------------:|
|  aws_instance.web.instance_type   | t3.small | t3.small | t3.large | 本番環境のパフォーマンス要件による意図的差分 |
| aws_instance.web.tags.Environment |   env1   |   env2   |   env3   |           環境識別タグの意図的差分           |


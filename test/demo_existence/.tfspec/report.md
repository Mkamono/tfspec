# Tfspec Check Results

## 意図されていない差分

意図されていない差分は検出されませんでした。

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:--------------:|:------------------------------------:|:----------------:|:------------------:|:-----:|:------------------:|:------------------------------------:|
|resource|aws_cloudwatch_metric_alarm.high_cpu||❌|✅|✅|監視設定の環境別要件による意図的差分|
||aws_instance.demo||✅|❌|✅|デモインスタンスの環境別配置要件|
|||instance_type|t3.micro|-|t3.large|-|
|||tags.Environment|env1|-|env3|-|
|||tags.Name|demo-instance-env1|-|demo-instance-env3|-|


# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:--------------:|:----------------:|:-------------:|:--------|:---------|:--------|
|resource|aws_instance.web|instance_type|t3.small|t3.medium|t3.large|
|||monitoring|false|true|true|
|||tags.Backup|false|true|true|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:--------------:|:----------------:|:----------------:|:-----|:-----|:-----|:---------------------------:|
|resource|aws_instance.web|tags.Environment|env1|env2|env3|Environment tag differences|


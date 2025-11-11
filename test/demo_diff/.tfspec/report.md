# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:--------------:|:-----------------:|:----------------:|:-----|:-----|:----------|
|resource|aws_instance.demo|tags.Environment|env1|env2|production|
|||tags.Project|demo|-|-|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:--------------:|:-----------------:|:-------------:|:--------|:---------|:--------|:------------------------------:|
|resource|aws_instance.demo|instance_type|t3.micro|t3.medium|t3.large|Demo configuration differences|


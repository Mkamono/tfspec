# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:--------------:|:----------------:|:----------------:|:--------|:--------------------------------------------|:------------------|
|resource|aws_instance.db|instance_type|t3.micro|t3.small|t3.medium|
|||key_name|-|db-staging-key|-|
|||tags.Environment|dev|staging|production|
|||user_data|-|#!/bin/bash<br>&nbsp;&nbsp;echo 'db staging'|-|
||aws_instance.web|instance_type|t3.small|t3.medium|t3.large|
|||tags.Environment|dev|staging|production|
|||tags.NullTag|-|-|actually_has_value|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:--------------:|:----------------:|:---------:|:-----|:-----------------------------------------|:--------------------------------------------|:------------------------:|
|resource|aws_instance.web|key_name|-|staging-key|-|オプショナル属性のテスト|
|||user_data|-|#!/bin/bash<br>&nbsp;&nbsp;echo 'staging'|#!/bin/bash<br>&nbsp;&nbsp;echo 'production'|null値のテスト用|


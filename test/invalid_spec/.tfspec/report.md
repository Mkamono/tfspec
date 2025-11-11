# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:-:|:-:|:-:|:-|:-|:-|
|resource|aws_instance.web|instance_type|t3.small|t3.medium|t3.large|
|||tags.Environment|env1|env2|env3|


# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:--------------:|:----------------:|:--------------------------------:|:-----|:-----|:-----|
|resource|aws_instance.web|root_block_device[0].volume_size|20|50|100|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:--------------:|:----------------:|:----------------:|:--------|:---------|:--------|:----------------------------------------:|
|resource|aws_instance.web|instance_type|t3.small|t3.medium|t3.large|環境別パフォーマンス要件による意図的差分|
|||tags.Backup|false|true|true|バックアップポリシーの環境別要件|
|||tags.Environment|env1|env2|env3|環境識別タグの意図的差分|


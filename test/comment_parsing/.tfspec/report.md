# Tfspec Check Results

## 意図されていない差分

意図されていない差分は検出されませんでした。

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:-:|:-:|:-:|:-|:-|:-|:-:|
|resource|aws_instance.cache|instance_type|t3.nano|t3.micro|t3.small|これも行末コメント|
||aws_instance.db|instance_type|t3.micro|t3.small|t3.medium|行末コメント|
||aws_instance.web|instance_type|t3.small|t3.medium|t3.large|複数行コメントのテスト<br>これは2行目のコメント<br>これは3行目のコメント|
|||tags.Environment|dev|staging|production|単一行コメント|
||aws_security_group.web|ingress[1]|-|{<br>&nbsp;&nbsp;from_port: 443,<br>&nbsp;&nbsp;to_port: 443<br>}|{<br>&nbsp;&nbsp;from_port: 443,<br>&nbsp;&nbsp;to_port: 443<br>}|新しいセクション（インデックス1は2番目のingressブロック）|


# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|
|:--------------:|:-------------------:|:-----------------------:|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|:-------------------------------------------------------------------------------------------------------------------------:|
|data|aws_ami|ubuntu.filter[0].values|[ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*]|[ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*]|
|local|allowed_cidr_blocks||[10.0.0.0/8, 172.16.0.0/12]|[10.0.0.0/8<br>172.16.0.0/12<br>192.168.0.0/16]|
||concat_test||concat([<br>    "a", # a<br>    "b", # b<br>    ], [<br>    "c",<br>    "d",<br>  ])|-|
||database_config||{backup_retention_period: 7<br>engine: mysql<br>engine_version: 8.0<br>multi_az: true}|{backup_retention_period: 30<br>engine: postgresql<br>engine_version: 14.0<br>multi_az: false<br>storage_encrypted: true}|
||dev_only_config||{debug_mode: true, log_level: debug}|-|
||enable_backup||false|true|
||enable_monitoring||true|false|
||long_object||{level1: {level2: {level3_1: {another_key: another_value<br>deep_nested_key: deep_nested_value<br>key: deep_value<br>yet_another_key: yet_another_value}<br>level3_2: {another_key: another_value<br>dee...|-|
||object_test||{name: test_object<br>nested: {key1: value1, key2: value2}<br>numbers: [1<br>2<br>3<br>4<br>5]}|-|
||prod_only_config||-|{alert_endpoints: [ops@example.com]<br>monitoring_level: production<br>ssl_enabled: true}|
|variable|instance_type|default|t3.micro|t3.small|
|||description|EC2 instance type|EC2 instance type for production|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|理由|
|:--------------:|:-----------------:|:-----------:|:-------------------------------------------------------:|:----------------------------------------------------:|:-------------------------------------------:|
|local|common_tags||{Environment: dev, Project: test}|{Environment: prod, Project: test}|環境別のlocal変数は意図的な差分|
||file_content||file("${path.module}/config.txt")|file("${path.module}/prod-config.txt")|-|
||merged_tags||merge(local.common_tags, { "AdditionalTag" = "value" })|merge(local.common_tags, { "Environment" = "prod" })|-|
||name_prefix||"app-${var.instance_type}"|"prod-${var.db_instance_class}"|-|
||name_with_length||length(var.instance_type)|length(var.db_instance_class)|HCL関数は環境によって異なることが予想される|
||vpc_cidr||10.0.0.0/16|10.1.0.0/16|-|
|output|vpc_cidr||false|true|本番環境では追加のoutputが必要|
|resource|module.vpc|environment|dev|prod|環境別のmodule設定は意図的な差分|
|||vpc_cidr|10.0.0.0/16|10.1.0.0/16|-|
|variable|db_instance_class||-|db.t3.micro|本番環境では追加のvariableが必要|


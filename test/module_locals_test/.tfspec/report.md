# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|
|:--------------:|:-------------------------------------------:|:----------------:|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|data|aws_ami.ubuntu|filter[0].values|[ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*]|[ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*]|
||google_certificate_manager_certificate.test||✅|❌|
|local|allowed_cidr_blocks||[10.0.0.0/8, 172.16.0.0/12]|[<br>&nbsp;&nbsp;10.0.0.0/8<br>&nbsp;&nbsp;172.16.0.0/12<br>&nbsp;&nbsp;192.168.0.0/16<br>]|
||concat_test||concat([<br>&nbsp;&nbsp;    "a", # a<br>&nbsp;&nbsp;    "b", # b<br>&nbsp;&nbsp;    ], [<br>&nbsp;&nbsp;    "c",<br>&nbsp;&nbsp;    "d",<br>&nbsp;&nbsp;  ])|-|
||database_config||{<br>&nbsp;&nbsp;backup_retention_period: 7<br>&nbsp;&nbsp;engine: mysql<br>&nbsp;&nbsp;engine_version: 8.0<br>&nbsp;&nbsp;multi_az: true<br>}|{<br>&nbsp;&nbsp;backup_retention_period: 30<br>&nbsp;&nbsp;engine: postgresql<br>&nbsp;&nbsp;engine_version: 14.0<br>&nbsp;&nbsp;multi_az: false<br>&nbsp;&nbsp;storage_encrypted: true<br>}|
||dev_only_config||{debug_mode: true, log_level: debug}|-|
||enable_backup||false|true|
||enable_monitoring||true|false|
||long_object||{level1: {level2: {<br>&nbsp;&nbsp;level3_1: {<br>&nbsp;&nbsp;another_key: another_value<br>&nbsp;&nbsp;deep_nested_key: deep_nested_value<br>&nbsp;&nbsp;key: deep_value<br>&nbsp;&nbsp;yet_another_key...|-|
||object_test||{<br>&nbsp;&nbsp;name: test_object<br>&nbsp;&nbsp;nested: {key1: value1, key2: value2}<br>&nbsp;&nbsp;numbers: [<br>&nbsp;&nbsp;1<br>&nbsp;&nbsp;2<br>&nbsp;&nbsp;3<br>&nbsp;&nbsp;4<br>&nbsp;&nbsp;5<br...|-|
||prod_only_config||-|{<br>&nbsp;&nbsp;alert_endpoints: [ops@example.com]<br>&nbsp;&nbsp;monitoring_level: production<br>&nbsp;&nbsp;ssl_enabled: true<br>}|
|variable|instance_type|default|t3.micro|t3.small|
|||description|EC2 instance type|EC2 instance type for production|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|理由|
|:--------------:|:-----------------:|:-----------:|:-------------------------------------------------------|:----------------------------------------------------|:-------------------------------------------:|
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


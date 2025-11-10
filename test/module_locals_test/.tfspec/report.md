# Tfspec Check Results

## 🚨 意図されていない差分

| リソースタイプ |     リソース名      |  属性パス   |                                      ENV 1                                       |                                                       ENV 2                                                       |
|:--------------:|:-------------------:|:-----------:|:--------------------------------------------------------------------------------:|:-----------------------------------------------------------------------------------------------------------------:|
|     local      | allowed_cidr_blocks |             |                           [10.0.0.0/8, 172.16.0.0/12]                            |                                    [10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16]                                    |
|                |   database_config   |             | {backup_retention_period: 7, engine: mysql, engine_version: 8.0, multi_az: true} | {backup_retention_period: 30, engine: postgresql, engine_version: 14.0, multi_az: false, storage_encrypted: true} |
|                |   dev_only_config   |             |                       {debug_mode: true, log_level: debug}                       |                                                         -                                                         |
|                |    enable_backup    |             |                                      false                                       |                                                       true                                                        |
|                |  enable_monitoring  |             |                                       true                                       |                                                       false                                                       |
|                |  prod_only_config   |             |                                        -                                         |               {alert_endpoints: [ops@example.com], monitoring_level: production, ssl_enabled: true}               |
|    variable    |    instance_type    |   default   |                                     t3.micro                                     |                                                     t3.small                                                      |
|                |                     | description |                                EC2 instance type                                 |                                         EC2 instance type for production                                          |

## 📝 無視された差分（意図的）

| リソースタイプ |    リソース名     |  属性パス   |               ENV 1               |               ENV 2                |               理由               |
|:--------------:|:-----------------:|:-----------:|:---------------------------------:|:----------------------------------:|:--------------------------------:|
|     local      |    common_tags    |             | {Environment: dev, Project: test} | {Environment: prod, Project: test} | 環境別のlocal変数は意図的な差分  |
|                |     vpc_cidr      |             |            10.0.0.0/16            |            10.1.0.0/16             |                -                 |
|     output     |     vpc_cidr      |             |               false               |                true                |  本番環境では追加のoutputが必要  |
|    resource    |    module.vpc     | environment |                dev                |                prod                | 環境別のmodule設定は意図的な差分 |
|                |                   |  vpc_cidr   |            10.0.0.0/16            |            10.1.0.0/16             |                -                 |
|    variable    | db_instance_class |             |                 -                 |            db.t3.micro             | 本番環境では追加のvariableが必要 |


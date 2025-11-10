# Tfspec Check Results

## 🚨 意図されていない差分

|           該当箇所            |                                      ENV 1                                       |                                                       ENV 2                                                       |
|:-----------------------------:|:--------------------------------------------------------------------------------:|:-----------------------------------------------------------------------------------------------------------------:|
|   local.allowed_cidr_blocks   |                           [10.0.0.0/8, 172.16.0.0/12]                            |                                    [10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16]                                    |
|     local.database_config     | {backup_retention_period: 7, engine: mysql, engine_version: 8.0, multi_az: true} | {backup_retention_period: 30, engine: postgresql, engine_version: 14.0, multi_az: false, storage_encrypted: true} |
|     local.dev_only_config     |                       {debug_mode: true, log_level: debug}                       |                                                         -                                                         |
|      local.enable_backup      |                                      false                                       |                                                       true                                                        |
|    local.enable_monitoring    |                                       true                                       |                                                       false                                                       |
|    local.prod_only_config     |                                        -                                         |               {alert_endpoints: [ops@example.com], monitoring_level: production, ssl_enabled: true}               |
|   var.instance_type.default   |                                     t3.micro                                     |                                                     t3.small                                                      |
| var.instance_type.description |                                EC2 instance type                                 |                                         EC2 instance type for production                                          |

## 📝 無視された差分（意図的）

|        該当箇所        |               ENV 1               |               ENV 2                |               理由               |
|:----------------------:|:---------------------------------:|:----------------------------------:|:--------------------------------:|
|   local.common_tags    | {Environment: dev, Project: test} | {Environment: prod, Project: test} | 環境別のlocal変数は意図的な差分  |
|     local.vpc_cidr     |            10.0.0.0/16            |            10.1.0.0/16             |                -                 |
| module.vpc.environment |                dev                |                prod                | 環境別のmodule設定は意図的な差分 |
|  module.vpc.vpc_cidr   |            10.0.0.0/16            |            10.1.0.0/16             |                -                 |
|    output.vpc_cidr     |               false               |                true                |  本番環境では追加のoutputが必要  |
| var.db_instance_class  |               false               |                true                | 本番環境では追加のvariableが必要 |


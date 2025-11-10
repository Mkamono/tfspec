# Tfspec Check Results

## 🚨 意図されていない差分

| リソースタイプ |      リソース名       |     属性パス      |    ENV 1    |   ENV 2    |
|:--------------:|:---------------------:|:-----------------:|:-----------:|:----------:|
|    resource    |   aws_instance.web    | tags.Environment  | development | production |
|                |  aws_instance.worker  |   instance_type   |  t3.micro   |  t3.small  |
|                | aws_rds_instance.main | allocated_storage |     20      |    100     |
|                |                       |   storage_type    |     gp2     |    gp3     |
|                |                       | tags.Environment  | development | production |
|                |  aws_subnet.private   |                   |     ❌      |     ✅     |
|                |     aws_vpc.main      | tags.Environment  | development | production |

## 📝 無視された差分（意図的）

| リソースタイプ |      リソース名       |     属性パス      |    ENV 1    |    ENV 2    |                                              理由                                              |
|:--------------:|:---------------------:|:-----------------:|:-----------:|:-----------:|:----------------------------------------------------------------------------------------------:|
|    resource    |   aws_instance.web    |   instance_type   |  t3.small   |  t3.medium  | 複数ファイル読み込みテスト用の無視ルール<br>dev環境はt3.smallだが本番はt3.medium（意図的差分） |
|                | aws_rds_instance.main | db_instance_class | db.t3.micro | db.t3.small |                              環境別データベース設定（意図的差分）                              |


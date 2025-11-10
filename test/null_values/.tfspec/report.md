# Tfspec Check Results

## 🚨 意図されていない差分

|             該当箇所              |  ENV 1   |       ENV 2       |       ENV 3        |
|:---------------------------------:|:--------:|:-----------------:|:------------------:|
|   aws_instance.db.instance_type   | t3.micro |     t3.small      |     t3.medium      |
|     aws_instance.db.key_name      |    -     |  db-staging-key   |         -          |
| aws_instance.db.tags.Environment  |   dev    |      staging      |     production     |
|     aws_instance.db.user_data     |    -     |    #!/bin/bash    |         -          |
|                                   |          | echo 'db staging' |                    |
|  aws_instance.web.instance_type   | t3.small |     t3.medium     |      t3.large      |
| aws_instance.web.tags.Environment |   dev    |      staging      |     production     |
|   aws_instance.web.tags.NullTag   |    -     |         -         | actually_has_value |

## 📝 無視された差分（意図的）

|          該当箇所          | ENV 1 |     ENV 2      |       ENV 3       |           理由           |
|:--------------------------:|:-----:|:--------------:|:-----------------:|:------------------------:|
| aws_instance.web.key_name  |   -   |  staging-key   |         -         | オプショナル属性のテスト |
| aws_instance.web.user_data |   -   |  #!/bin/bash   |    #!/bin/bash    |     null値のテスト用     |
|                            |       | echo 'staging' | echo 'production' |                          |


# Tfspec Check Results

## 🚨 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:--------------:|:----------------:|:--------------------------------:|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|:----------------------------------------------------------------------------------------------------------------------------------------------------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|resource|aws_instance.web|instance_type|t3.small|t3.medium|t3.large|
|||root_block_device[0].volume_size|999999999999|888888888888|777777777777|
|||root_block_device[0].volume_type|gp3|gp2|-|
|||tags.Environment|dev|staging|production|
|||tags.VeryLongTagKey|This is a very long tag value that might cause display issues in the report generation. It contains many characters and should test the limits of string handling in the diff detection and reporting sy...|This is a different very long tag value that also might cause display issues. It has different content but similar length to test various scenarios.|This is the production very long tag value that definitely will cause display issues if not handled properly. It contains the most characters and should thoroughly test the string handling limits.|

## 📝 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:--------------:|:----------------:|:---------------:|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|:----------------------:|
|resource|aws_instance.web|security_groups|[sg-12345678901234567<br>sg-23456789012345678<br>sg-34567890123456789<br>sg-45678901234567890<br>sg-56789012345678901<br>sg-67890123456789012<br>sg-78901234567890123<br>sg-89012345678901234<br>sg-9012...|[sg-11111111111111111<br>sg-22222222222222222<br>sg-33333333333333333<br>sg-44444444444444444<br>sg-55555555555555555<br>sg-66666666666666666<br>sg-77777777777777777]|[sg-prod-111111111111<br>sg-prod-222222222222<br>sg-prod-333333333333<br>sg-prod-444444444444<br>sg-prod-555555555555<br>sg-prod-666666666666<br>sg-prod-777777777777<br>sg-prod-888888888888<br>sg-prod...|長いリストのテスト|
|||user_data|#!/bin/bash<br># This is a very long user data script that contains many lines<br># and might cause issues with parsing or display<br>echo "Starting very long script..."<br>for i in {1..1000}; do<br> ...|#!/bin/bash<br># This is a different very long user data script<br>echo "Starting different long script..."<br>for i in {1..500}; do<br>  echo "Different processing item $i"<br>  echo "This is a diffe...|#!/bin/bash<br># Production very long user data script<br>echo "Starting production long script..."<br>for i in {1..2000}; do<br>  echo "Production processing item $i"<br>  echo "This is production li...|巨大な値の差分テスト用|


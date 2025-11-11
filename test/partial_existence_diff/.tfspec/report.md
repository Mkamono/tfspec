# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:--------------:|:------------------------------------:|:----------------:|:-----:|:----------------------------------------------------------------------------------------------------------------------------------------------:|:--------------------------------------------------------------------------------------------------------------------------------------------------:|
|resource|aws_cloudwatch_metric_alarm.high_cpu||❌|✅|✅|
||aws_security_group.web|ingress[1]|-|{<br>&nbsp;&nbsp;cidr_blocks: [["0.0.0.0/0"]],<br>&nbsp;&nbsp;from_port: 443,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 443<br>}|{<br>&nbsp;&nbsp;cidr_blocks: [["172.16.0.0/12"]],<br>&nbsp;&nbsp;from_port: 443,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 443<br>}|
|||tags.Environment|env1|env2|env3|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:--------------:|:----------------:|:----------------:|:-----:|:-----:|:-----:|:---------------------------:|
|resource|aws_instance.web|tags.Environment|env1|env2|env3|Environment tag differences|


# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:--------------:|:--------------------------------:|:-------------------------------:|:--------------|:--------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------|
|resource|aws_launch_configuration.complex|ebs_block_device[0].throughput|-|-|250|
|||ebs_block_device[0].volume_size|10|15|50|
|||ebs_block_device[0].volume_type|gp2|gp3|gp3|
|||ebs_block_device[1].throughput|-|-|500|
|||ebs_block_device[1].volume_size|20|25|100|
|||ebs_block_device[2].iops|100|150|1000|
|||ebs_block_device[2].volume_size|30|35|200|
|||ebs_block_device[4]|-|{<br>&nbsp;&nbsp;device_name: "/dev/sdf",<br>&nbsp;&nbsp;volume_size: 50,<br>&nbsp;&nbsp;volume_type: "gp3"<br>}|{<br>&nbsp;&nbsp;device_name: "/dev/sdf",<br>&nbsp;&nbsp;throughput: 1000,<br>&nbsp;&nbsp;volume_size: 500,<br>&nbsp;&nbsp;volume_type: "gp3"<br>}|
|||ebs_block_device[5]|-|-|{<br>&nbsp;&nbsp;device_name: "/dev/sdg",<br>&nbsp;&nbsp;throughput: 1000,<br>&nbsp;&nbsp;volume_size: 1000,<br>&nbsp;&nbsp;volume_type: "gp3"<br>}|
|||image_id|ami-12345678|ami-87654321|ami-production|
|||instance_type|t3.small|t3.medium|t3.large|
|||name|complex-lc-dev|complex-lc-staging|complex-lc-production|
||aws_security_group.complex|ingress[0].cidr_blocks|[10.0.1.0/24]|[10.0.1.0/24, 10.0.5.0/24]|-|
|||ingress[3].cidr_blocks|[10.0.2.0/24]|[10.0.2.0/24, 10.0.6.0/24]|-|
|||ingress[6]|-|{<br>&nbsp;&nbsp;cidr_blocks: [["10.0.8.0/24"]],<br>&nbsp;&nbsp;from_port: 9200,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 9200<br>}|{<br>&nbsp;&nbsp;cidr_blocks: [["10.0.9.0/24"]],<br>&nbsp;&nbsp;from_port: 9100,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 9100<br>}|
|||ingress[7]|-|-|{<br>&nbsp;&nbsp;cidr_blocks: [["10.0.10.0/24"]],<br>&nbsp;&nbsp;from_port: 3000,<br>&nbsp;&nbsp;protocol: "tcp",<br>&nbsp;&nbsp;to_port: 3000<br>}|
|||name|complex-sg-dev|complex-sg-staging|complex-sg-production|
|||tags.Environment|dev|staging|production|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:--------------:|:--------------------------------:|:-------------------------------:|:-------------|:--------------------------|:-------------|:----:|
|resource|aws_launch_configuration.complex|ebs_block_device[3].throughput|-|-|500|-|
|||ebs_block_device[3].volume_size|40|45|100|-|
|||ebs_block_device[3].volume_type|gp2|gp3|gp3|-|
||aws_security_group.complex|ingress[2].cidr_blocks|[10.0.0.0/16]|-|[10.0.0.0/24]|-|
|||ingress[5].cidr_blocks|[10.0.4.0/24]|[10.0.4.0/24, 10.0.7.0/24]|-|-|


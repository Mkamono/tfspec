# Tfspec Check Results

## 🚨 意図されていない差分

|               該当箇所               | ENV 1 | ENV 2 | ENV 3 |
|:------------------------------------:|:-----:|:-----:|:-----:|
| aws_cloudwatch_metric_alarm.high_cpu | false | true  | true  |

## 📝 無視された差分（意図的）

|              該当箇所              |       ENV 1        | ENV 2 |       ENV 3        |                理由                 |
|:----------------------------------:|:------------------:|:-----:|:------------------:|:-----------------------------------:|
|         aws_instance.demo          |        true        |   -   |        true        | Demo resource existence differences |
|  aws_instance.demo.instance_type   |      t3.micro      |   -   |      t3.large      |                  -                  |
| aws_instance.demo.tags.Environment |        env1        |   -   |        env3        |                  -                  |
|    aws_instance.demo.tags.Name     | demo-instance-env1 |   -   | demo-instance-env3 |                  -                  |


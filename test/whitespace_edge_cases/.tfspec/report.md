# Tfspec Check Results

## 🚨 意図されていない差分

| 該当箇所 | env1 | env2 | env3 |
|----------|-------|-------|-------|
| aws_instance.web.tags.	Leading Tab Key | Trailing Space Value  | - | - |
| aws_instance.web.tags.Key With Spaces | Value With	Tabs | Different Value With Spaces | Production Value |
| aws_instance.web.tags.Leading Tab Key | - | Different Trailing Value | Clean Production Value |
| aws_instance.web.tags.Mixed	Spaces　And　Full-Width　Spaces | 　全角空白を含む値　 | 異なる　全角空白　値 | 本番環境用値 |

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_instance.web.instance_type | t3.small | t3.medium | t3.large | タブと空白が混在するルール名 |
| aws_instance.web.tags.Environment | dev | staging | production | 全角空白を含むコメント |
| aws_instance.web.user_data | #!/bin/bash
 echo "Mixed tabs and spaces"
	echo "More mixed indentation"
 	echo "Different indentation"
 |  #!/bin/bash
 echo "Different spacing"
echo "Different tab usage"
   echo "Different indentation"
 | #!/bin/bash
echo "Clean production formatting"
echo "Consistent indentation"
echo "No mixed whitespace"
 | 行末空白あり |


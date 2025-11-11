# Tfspec Check Results

## 意図されていない差分

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|
|:-:|:-:|:-:|:-|:-|:-|
|resource|aws_instance.web|tags.	Leading Tab Key|Trailing Space Value|-|-|
|||tags.Key With Spaces|Value With	Tabs|Different Value With Spaces|Production Value|
|||tags.Leading Tab Key|-|Different Trailing Value|Clean Production Value|
|||tags.Mixed	Spaces　And　Full-Width　Spaces|全角空白を含む値|異なる　全角空白　値|本番環境用値|

## 無視された差分（意図的）

|リソースタイプ|リソース名|属性パス|ENV 1|ENV 2|ENV 3|理由|
|:-:|:-:|:-:|:-|:-|:-|:-:|
|resource|aws_instance.web|instance_type|t3.small|t3.medium|t3.large|タブと空白が混在するルール名|
|||tags.Environment|dev|staging|production|全角空白を含むコメント|
|||user_data|#!/bin/bash<br>&nbsp;&nbsp; echo "Mixed tabs and spaces"<br>&nbsp;&nbsp;	echo "More mixed indentation"<br>&nbsp;&nbsp; 	echo "Different indentation"<br>&nbsp;&nbsp;|#!/bin/bash<br>&nbsp;&nbsp; echo "Different spacing"<br>&nbsp;&nbsp;echo "Different tab usage"<br>&nbsp;&nbsp;   echo "Different indentation"<br>&nbsp;&nbsp;|#!/bin/bash<br>&nbsp;&nbsp;echo "Clean production formatting"<br>&nbsp;&nbsp;echo "Consistent indentation"<br>&nbsp;&nbsp;echo "No mixed whitespace"<br>&nbsp;&nbsp;|行末空白あり|


resource "aws_instance" "web"	{
  instance_type = "t3.small"

  # タブとスペースが混在した値
  user_data = <<-EOF
  	#!/bin/bash
    echo "Mixed tabs and spaces"
  		echo "More mixed indentation"
    	echo "Different indentation"
  EOF

  tags	=	{
    Environment = "dev"
    # 空白文字を含む値
    "Key With Spaces" = "Value With	Tabs"
    "	Leading Tab Key" = "Trailing Space Value "
    "Mixed	Spaces　And　Full-Width　Spaces" = "　全角空白を含む値　"
  }
}



# 空行が多い構成
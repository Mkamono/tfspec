resource "aws_instance"    "web" {
  instance_type    =    "t3.medium"

  # 異なる空白文字パターン
  user_data    =    <<-EOF
    #!/bin/bash
    echo "Different spacing"
  	echo "Different tab usage"
      echo "Different indentation"
  EOF

  tags = {
    Environment    =    "staging"
    # 異なる空白パターン
    "Key With Spaces"    =    "Different Value With Spaces"
    "Leading Tab Key" = "Different Trailing Value"
    "Mixed	Spaces　And　Full-Width　Spaces" = "異なる　全角空白　値"
  }
}

# スペースとタブの使い分け
resource "aws_instance" "web" {
  instance_type = "t3.medium"

  # 異なる巨大なユーザーデータ
  user_data = <<-EOF
    #!/bin/bash
    # This is a different very long user data script
    echo "Starting different long script..."
    for i in {1..500}; do
      echo "Different processing item $i"
      echo "This is a different line $i of the script"
      echo "Different content to make this really long..."
      sleep 0.05
    done

    # Different configuration
    apt update -y
    apt install -y nginx
    systemctl start nginx
    systemctl enable nginx

    # Create a different large configuration file
    cat > /tmp/different_large_config.conf << 'INNER_EOF'
    # Different configuration file with many options
    different_option1=different_value1
    different_option2=different_value2
    different_option3=different_value3
    INNER_EOF

    echo "Different script completed successfully"
  EOF

  # 異なる非常に大きな数値
  root_block_device {
    volume_size = 888888888888
    volume_type = "gp2"
  }

  # 異なる長いリスト
  security_groups = [
    "sg-11111111111111111",
    "sg-22222222222222222",
    "sg-33333333333333333",
    "sg-44444444444444444",
    "sg-55555555555555555",
    "sg-66666666666666666",
    "sg-77777777777777777"
  ]

  tags = {
    Environment = "staging"
    VeryLongTagKey = "This is a different very long tag value that also might cause display issues. It has different content but similar length to test various scenarios."
  }
}
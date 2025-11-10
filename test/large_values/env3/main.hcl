resource "aws_instance" "web" {
  instance_type = "t3.large"

  # 本番用の巨大なユーザーデータ
  user_data = <<-EOF
    #!/bin/bash
    # Production very long user data script
    echo "Starting production long script..."
    for i in {1..2000}; do
      echo "Production processing item $i"
      echo "This is production line $i of the script"
      echo "Production content to make this really long..."
      if [ $((i % 100)) -eq 0 ]; then
        echo "Checkpoint at item $i"
      fi
    done

    # Production configuration
    yum update -y
    yum install -y httpd mysql-server php
    systemctl start httpd
    systemctl enable httpd
    systemctl start mysqld
    systemctl enable mysqld

    # Production large configuration file
    cat > /tmp/production_large_config.conf << 'INNER_EOF'
    # Production configuration file with many options
    prod_option1=prod_value1
    prod_option2=prod_value2
    prod_option3=prod_value3
    prod_option4=prod_value4
    prod_option5=prod_value5
    prod_option6=prod_value6
    prod_option7=prod_value7
    prod_option8=prod_value8
    INNER_EOF

    echo "Production script completed successfully"
  EOF

  # 本番用の非常に大きな数値
  root_block_device {
    volume_size = 777777777777
    volume_type = "gp3"
  }

  # 本番用の長いリスト
  security_groups = [
    "sg-prod-111111111111",
    "sg-prod-222222222222",
    "sg-prod-333333333333",
    "sg-prod-444444444444",
    "sg-prod-555555555555",
    "sg-prod-666666666666",
    "sg-prod-777777777777",
    "sg-prod-888888888888",
    "sg-prod-999999999999",
    "sg-prod-000000000000",
    "sg-prod-aaaaaaaaaaaa"
  ]

  tags = {
    Environment = "production"
    VeryLongTagKey = "This is the production very long tag value that definitely will cause display issues if not handled properly. It contains the most characters and should thoroughly test the string handling limits."
  }
}
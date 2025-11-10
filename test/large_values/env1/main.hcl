resource "aws_instance" "web" {
  instance_type = "t3.small"

  # 巨大なユーザーデータ
  user_data = <<-EOF
    #!/bin/bash
    # This is a very long user data script that contains many lines
    # and might cause issues with parsing or display
    echo "Starting very long script..."
    for i in {1..1000}; do
      echo "Processing item $i"
      echo "This is line $i of the script"
      echo "Adding more content to make this really long..."
      sleep 0.1
    done

    # More configuration
    yum update -y
    yum install -y docker
    systemctl start docker
    systemctl enable docker

    # Create a large configuration file
    cat > /tmp/large_config.conf << 'INNER_EOF'
    # Configuration file with many options
    option1=value1
    option2=value2
    option3=value3
    option4=value4
    option5=value5
    INNER_EOF

    echo "Script completed successfully"
  EOF

  # 非常に大きな数値
  root_block_device {
    volume_size = 999999999999
    volume_type = "gp3"
  }

  # 長いリスト
  security_groups = [
    "sg-12345678901234567",
    "sg-23456789012345678",
    "sg-34567890123456789",
    "sg-45678901234567890",
    "sg-56789012345678901",
    "sg-67890123456789012",
    "sg-78901234567890123",
    "sg-89012345678901234",
    "sg-90123456789012345"
  ]

  tags = {
    Environment = "dev"
    VeryLongTagKey = "This is a very long tag value that might cause display issues in the report generation. It contains many characters and should test the limits of string handling in the diff detection and reporting system."
  }
}
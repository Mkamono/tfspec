# Tfspec Check Results

## 🚨 意図されていない差分

| 該当箇所 | env1 | env2 | env3 |
|----------|-------|-------|-------|
| aws_instance.web.instance_type | t3.small | t3.medium | t3.large |
| aws_instance.web.root_block_device[0].volume_size | 999999999999 | 888888888888 | 777777777777 |
| aws_instance.web.root_block_device[0].volume_type | gp3 | gp2 | - |
| aws_instance.web.tags.Environment | dev | staging | production |
| aws_instance.web.tags.VeryLongTagKey | This is a very long tag value that might cause display issues in the report generation. It contains many characters and should test the limits of string handling in the diff detection and reporting system. | This is a different very long tag value that also might cause display issues. It has different content but similar length to test various scenarios. | This is the production very long tag value that definitely will cause display issues if not handled properly. It contains the most characters and should thoroughly test the string handling limits. |

## 📝 無視された差分（意図的）

| 該当箇所 | env1 | env2 | env3 | 理由 |
|----------|-------|-------|-------|------|
| aws_instance.web.security_groups | [sg-12345678901234567, sg-23456789012345678, sg-34567890123456789, sg-45678901234567890, sg-56789012345678901, sg-67890123456789012, sg-78901234567890123, sg-89012345678901234, sg-90123456789012345] | [sg-11111111111111111, sg-22222222222222222, sg-33333333333333333, sg-44444444444444444, sg-55555555555555555, sg-66666666666666666, sg-77777777777777777] | [sg-prod-111111111111, sg-prod-222222222222, sg-prod-333333333333, sg-prod-444444444444, sg-prod-555555555555, sg-prod-666666666666, sg-prod-777777777777, sg-prod-888888888888, sg-prod-999999999999, sg-prod-000000000000, sg-prod-aaaaaaaaaaaa] | 長いリストのテスト |
| aws_instance.web.user_data | #!/bin/bash
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
 | #!/bin/bash
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
 | #!/bin/bash
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
 | 巨大な値の差分テスト用 |


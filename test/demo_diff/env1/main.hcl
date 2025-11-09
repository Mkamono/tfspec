resource "aws_instance" "web" {
  instance_type = "t3.micro"  # 意図的にt3.smallではなくt3.microに設定
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "web-server"
    Environment = "env1"
    Project = "test"  # 仕様書で宣言されていない属性
  }
}
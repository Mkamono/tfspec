resource "aws_instance" "web" {
  instance_type = "t3.medium" # この差分は仕様書に宣言されていない（意図的）
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "web-server"
    Environment = "env2"
  }
}
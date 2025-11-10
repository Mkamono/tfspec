resource "aws_instance" "demo" {
  instance_type = "t3.micro"  # env1では小さなインスタンス
  ami           = "ami-0abcdef1234567890"

  tags = {
    Name = "demo-server"
    Environment = "env1"
    Project = "demo"  # demoプロジェクト用の追加属性
  }
}
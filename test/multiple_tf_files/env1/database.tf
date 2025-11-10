resource "aws_rds_instance" "main" {
  allocated_storage    = 20
  storage_type        = "gp2"
  engine              = "mysql"
  engine_version      = "8.0"
  db_instance_class   = "db.t3.micro"
  identifier          = "main-db"
  username            = "admin"
  password            = "password"

  tags = {
    Name = "main-database"
    Environment = "development"
  }
}
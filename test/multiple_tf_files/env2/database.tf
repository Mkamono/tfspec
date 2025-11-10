resource "aws_rds_instance" "main" {
  allocated_storage    = 100     # 本番環境はより大きなストレージ
  storage_type        = "gp3"    # より高速なストレージ
  engine              = "mysql"
  engine_version      = "8.0"
  db_instance_class   = "db.t3.small"  # より大きなインスタンスクラス
  identifier          = "main-db"
  username            = "admin"
  password            = "password"

  tags = {
    Name = "main-database"
    Environment = "production"
  }
}
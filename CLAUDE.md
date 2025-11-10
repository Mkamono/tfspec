# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

会話、ドキュメント記述は日本語で行うこと

# プロダクト概要
tfspecは、Terraformの環境間構成差分を自動検出し、「意図的な差分」として宣言されたもの以外を「構成ドリフト」として報告するツールです。

# プロジェクト構造
- `.tfspec/`ディレクトリにすべての設定が集約される
- **意図的な差分**は`.tfspec/.tfspecignore`（単一ファイル）または`.tfspec/.tfspecignore/`（分割ファイル）で管理
- シンプルなリソース名/属性名のリスト形式で記述

## .tfspecignore形式
- **超シンプル**: リソース名・属性名の単純なリスト
- **柔軟性**: 単一ファイルまたはカテゴリ別分割ファイル
- **直感的**: 「これは意図的な差分だから無視」という明確な意図

# アーキテクチャ
## 動作フロー
1. **差分検出**: 全環境のTerraformファイル（.tf/.hclファイル）を解析・比較
2. **フィルタリング**: `.tfspecignore`に記述されたリソース・属性は「意図的な差分」として除外
3. **レポート**: 残った差分のみを「構成ドリフト」として報告

## HCLパッケージの使用
標準的なTerraform HCL構文の解析のみで、カスタム関数は不要

### 主要なHCLパッケージ
- `hclparse` - ファイルベースの解析
- `gohcl` - Go構造体ベースのデコード（シンプルな構造体マッピング）

# ディレクトリ構造
```
tfspec/
├── main.go          # エントリーポイント
├── go.mod           # Goモジュール定義（Go 1.25.1）
├── app/             # アプリケーションコード
│   ├── cmd.go          # コマンドライン処理・UI
│   ├── differ.go       # 差分検出ロジック
│   ├── parser.go       # HCL解析・.tfspecignore読み込み
│   ├── types.go        # データ構造定義
│   ├── reporter.go     # Markdownレポート生成
│   └── ignore_matcher.go  # 無視ルール判定
├── docs/            # HCLライブラリの技術文書
│   └── hcl_deepwiki.md  # HCLライブラリの詳細仕様
└── test/            # テストケース群
    ├── basic_attribute_diff/      # 基本的な属性差分
    ├── comment_parsing/          # コメント解析
    ├── demo_diff/               # デモ差分
    ├── demo_existence/          # デモ存在差分
    ├── env_patterns/            # 環境パターン
    ├── invalid_spec/            # 無効な仕様書
    ├── list_diff/               # リスト差分
    ├── multiple_attribute_diff/ # 複数属性差分
    ├── multiple_spec_files/     # 複数仕様書ファイル
    ├── multiple_tf_files/       # 複数.tf/.hclファイル読み込み
    ├── nested_block_diff/       # ネストブロック差分
    ├── partial_existence_diff/  # 部分存在差分
    ├── resource_existence_diff/ # リソース存在差分
    ├── single_spec_file/        # 単一仕様書ファイル
    └── undeclared_diff/         # 未宣言差分
```

# テストケース例

## basic_attribute_diff の例
```
test/basic_attribute_diff/
├── .tfspec/
│   └── .tfspecignore     # 意図的な差分を宣言
├── env1/
│   └── main.hcl         # t3.small を使用
├── env2/
│   └── main.hcl         # t3.small を使用
└── env3/
    └── main.hcl         # t3.large を使用（本番環境）
```

意図的差分の宣言（.tfspec/.tfspecignore）:
```
# 本番環境のパフォーマンス要件による意図的差分
aws_instance.web.instance_type

# 環境識別タグの意図的差分
aws_instance.web.tags.Environment
```

## comment_parsing の例
コメント付き.tfspecignoreファイルの解析をテストするケース。
```
# 複数行コメントのテスト
# これは2行目のコメント
# これは3行目のコメント
aws_instance.web.instance_type

# 単一行コメント
aws_instance.web.tags.Environment

aws_instance.db.instance_type  # 行末コメント
aws_instance.cache.instance_type # これも行末コメント

# インデックス1は2番目のingressブロック
aws_security_group.web.ingress[1]
```

## リソース存在差分の例
```
# 本番環境でのSLA保証のための必須監視（他環境では不要）
aws_cloudwatch_metric_alarm.high_cpu
```

## ブロック存在差分の例
```
# SSL/TLS通信要件による意図的差分（開発環境はHTTPのみ、インデックス1は2番目のingressブロック）
aws_security_group.web.ingress[1]
```

## カテゴリ別分割の例
```
test/multiple_categories/
├── .tfspec/
│   └── .tfspecignore/
│       ├── security.txt      # セキュリティ関連差分
│       ├── performance.txt   # パフォーマンス関連差分
│       └── monitoring.txt    # 監視関連差分
├── env1/
├── env2/
└── env3/
```

security.txt:
```
# SSL/TLS設定の環境別要件（インデックス1は2番目のingressブロック）
aws_security_group.web.ingress[1]
aws_load_balancer.web.ssl_policy
```

performance.txt:
```
# 環境別パフォーマンス要件
aws_instance.web.instance_type
aws_rds_instance.main.db_instance_class
```

## 複数ファイル読み込みの例（multiple_tf_files）
各環境ディレクトリ内の全ての.tf/.hclファイルを自動検出・読み込み：
```
test/multiple_tf_files/
├── .tfspec/
│   └── .tfspecignore     # 意図的な差分を宣言
├── env1/
│   ├── compute.tf        # EC2インスタンス定義
│   ├── database.tf       # RDSインスタンス定義
│   └── network.hcl       # VPC、サブネット定義
└── env2/
    ├── compute.tf        # EC2インスタンス定義
    ├── database.tf       # RDSインスタンス定義
    └── network.hcl       # VPC、サブネット定義（+ 追加リソース）
```

**ファイル読み込み仕様：**
- 環境ディレクトリ内の全.tf/.hclファイルを自動検出
- ファイルを順次解析してリソースを結合
- `main.hcl`/`main.tf`がない環境でも動作
- ファイル順序は安定化（ソート済み）

# アーキテクチャ詳細

## 主要コンポーネント

### cmd.go
- コマンドライン引数の処理
- checkコマンドの実装
- 各処理ステップの調整
- エラーハンドリング

### differ.go
- 環境間の差分検出ロジック
- IgnoreMatcherを使用した無視ルール適用
- リソース存在差分、属性差分、ブロック差分の検出

### parser.go
- HCLファイルの解析（単一ファイル・複数ファイル対応）
- .tfspecignoreファイルの読み込み
- 単一ファイル・分割ファイル両方に対応

### types.go
- データ構造の定義
- EnvResource, EnvBlock, DiffResult等

### reporter.go
- Markdownテーブル形式でのレポート生成
- 差分結果の整理・ソート
- 値の補完処理

### ignore_matcher.go
- 無視ルールのマッチング判定
- 階層的パスマッチング
- インデックスベースの判定

# 開発関連コマンド
このプロジェクトはGoモジュールです（Go 1.25.1）
- `go build` - ビルド
- `go test` - テスト実行
- `go mod tidy` - 依存関係の整理

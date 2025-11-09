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
1. **差分検出**: 全環境のTerraformファイル（main.hcl）を解析・比較
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
├── main.go          # エントリーポイント（現在は空の実装）
├── go.mod           # Goモジュール定義（Go 1.25.1）
├── app/             # アプリケーションコード（将来の実装場所）
├── docs/            # HCLライブラリの技術文書
│   └── hcl_deepwiki.md  # HCLライブラリの詳細仕様
└── test/            # テストケース群
    ├── basic_attribute_diff/      # 基本的な属性差分
    ├── env_patterns/             # 環境パターン
    ├── invalid_spec/             # 無効な仕様書
    ├── list_diff/                # リスト差分
    ├── multiple_attribute_diff/  # 複数属性差分
    ├── multiple_spec_files/      # 複数仕様書ファイル
    ├── nested_block_diff/        # ネストブロック差分
    ├── partial_existence_diff/   # 部分存在差分
    ├── resource_existence_diff/  # リソース存在差分
    ├── single_spec_file/         # 単一仕様書ファイル
    └── undeclared_diff/          # 未宣言差分
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

## リソース存在差分の例
```
# 本番環境でのSLA保証のための必須監視（他環境では不要）
aws_cloudwatch_metric_alarm.high_cpu
```

## ブロック存在差分の例
```
# SSL/TLS通信要件による意図的差分（開発環境はHTTPのみ）
aws_security_group.web.ingress[443]
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
# SSL/TLS設定の環境別要件
aws_security_group.web.ingress[443]
aws_load_balancer.web.ssl_policy
```

performance.txt:
```
# 環境別パフォーマンス要件
aws_instance.web.instance_type
aws_rds_instance.main.db_instance_class
```

# 開発関連コマンド
このプロジェクトはGoモジュールです（Go 1.25.1）
- `go build` - ビルド
- `go test` - テスト実行
- `go mod tidy` - 依存関係の整理

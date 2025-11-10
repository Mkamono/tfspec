# tfspec

**Terraform環境間構成差分検出ツール**
(読み: ティーエフスペック / 意味: Terraform Specification)

## 概要

tfspecは、Terraformの環境間構成差分を自動検出し、「意図的な差分」として宣言されたもの以外を「構成ドリフト」として報告するツールです。

`.tfspec/`ディレクトリに設定が集約され、意図的な差分は`.tfspec/.tfspecignore`（単一ファイル）または`.tfspec/.tfspecignore/`（分割ファイル）で管理されます。シンプルなリソース名・属性名のリスト形式で記述するため、習得コストが低く既存プロジェクトに容易に導入できます。

## 特徴

- **シンプルな記法**: リソース名・属性名の単純なリスト形式
- **柔軟な管理方式**: 単一ファイルまたはカテゴリ別分割ファイル
- **非侵襲的**: 既存のプロジェクト構造を変更せず導入可能
- **直感的**: 「これは意図的な差分だから無視」という明確な意図で記述

## 基本的な使い方

### 1. 構成差分をチェック

```bash
tfspec check
```

### 2. 特定の環境のみチェック

```bash
tfspec check env1 env2 env3
```

### 3. 結果をMarkdownファイルに出力

```bash
tfspec check -o report.md
# または .tfspec/report.md に出力
tfspec check -o
```

## .tfspecignore形式

### 単一ファイル（`.tfspec/.tfspecignore`）

```
# 本番環境のパフォーマンス要件による意図的差分
aws_instance.web.instance_type

# 環境識別タグの意図的差分
aws_instance.web.tags.Environment

# SSL/TLS通信要件による意図的差分（インデックス1は2番目のingressブロック）
aws_security_group.web.ingress[1]
```

### 分割ファイル（`.tfspec/.tfspecignore/`）

```
.tfspec/
└── .tfspecignore/
    ├── security.txt      # セキュリティ関連差分
    ├── performance.txt   # パフォーマンス関連差分
    └── monitoring.txt    # 監視関連差分
```

**security.txt例:**
```
# SSL/TLS設定の環境別要件
aws_security_group.web.ingress[1]
aws_load_balancer.web.ssl_policy
```

**performance.txt例:**
```
# 環境別パフォーマンス要件
aws_instance.web.instance_type
aws_rds_instance.main.db_instance_class
```

## アーキテクチャ

### 動作フロー

1. **差分検出**: 全環境のTerraformファイル（.tf/.hclファイル）を解析・比較
2. **フィルタリング**: `.tfspecignore`に記述されたリソース・属性は「意図的な差分」として除外
3. **レポート**: 残った差分のみを「構成ドリフト」として報告

### ファイル構成

```
tfspec/
├── main.go              # エントリーポイント
├── app/
│   ├── cmd.go          # コマンドライン処理
│   ├── differ.go       # 差分検出ロジック
│   ├── parser.go       # HCL解析
│   ├── types.go        # データ構造定義
│   ├── reporter.go     # レポート生成
│   └── ignore_matcher.go  # 無視ルール判定
└── test/               # テストケース群
```

## 開発

### ビルド

```bash
go build
```

### テスト

プロジェクトはGoモジュールです（Go 1.25.1）

```bash
go test ./...
```

## プロジェクト構造例

```
your-terraform-project/
├── .tfspec/
│   ├── .tfspecignore     # 意図的な差分の宣言
│   └── report.md         # 生成される差分レポート
├── env1/
│   ├── main.tf           # または複数の.tf/.hclファイル
│   ├── compute.tf        # （全ての.tf/.hclファイルを自動検出）
│   └── network.hcl
├── env2/
│   ├── main.tf
│   ├── compute.tf
│   └── network.hcl
└── env3/
    ├── main.tf
    ├── compute.tf
    └── network.hcl
```

## ファイル読み込み仕様

- **自動検出**: 環境ディレクトリ内の全ての.tf/.hclファイルを自動検出
- **ファイル結合**: 複数ファイルのリソースを結合して解析
- **柔軟性**: `main.hcl`/`main.tf`がない環境でも動作
- **後方互換性**: 従来の単一ファイル構成も引き続きサポート

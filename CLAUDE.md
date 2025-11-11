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
├── go.sum           # 依存関係ロック
├── app/             # アプリケーションコード
│   ├── cmd/
│   │   └── cmd.go              # コマンドライン処理・UI（Cobra）
│   ├── config/
│   │   └── config.go           # 設定管理・環境ディレクトリ検出
│   ├── differ/
│   │   ├── differ.go           # 差分検出ロジック（リソース、属性、ブロック）
│   │   └── ignore_matcher.go   # 無視ルール判定・検証
│   ├── interfaces/
│   │   └── interfaces.go       # サービスインターフェース定義（DI用）
│   ├── parser/
│   │   ├── parser.go           # HCL解析・.tfspecignore読み込み
│   │   └── formatter.go        # cty.Value値のフォーマット（Markdown対応）
│   ├── reporter/
│   │   └── reporter.go         # Markdownテーブルレポート生成
│   ├── service/
│   │   ├── analyzer.go         # HCL解析・差分検出の統合
│   │   ├── output.go           # レポート生成・出力
│   │   └── service.go          # コマンド実行の統合（AppService）
│   └── types/
│       └── types.go            # データ構造定義（EnvResource, DiffResult等）
├── docs/            # 技術文書
│   └── hcl_deepwiki.md  # HCLライブラリの詳細仕様
└── test/            # テストケース群（26種類）
    ├── basic_attribute_diff/        # 基本的な属性差分
    ├── comment_parsing/             # .tfspecignoreコメント解析
    ├── deeply_nested_blocks/        # 深くネストされたブロック
    ├── demo_diff/ & demo_existence/ # デモンストレーション
    ├── duplicate_resources/         # 同名リソースの処理
    ├── empty_files/                 # 空HCLファイル処理
    ├── env_patterns/                # 異なる環境パターン
    ├── invalid_spec/                # 無効な.tfspecignore
    ├── large_values/                # 大型データ処理
    ├── list_diff/                   # リスト属性差分
    ├── malformed_hcl/               # 不正なHCL構文
    ├── module_locals_test/          # モジュール・ローカル変数
    ├── multiple_attribute_diff/     # 複数属性差分
    ├── multiple_spec_files/         # 複数.tfspecignoreファイル
    ├── multiple_tf_files/           # 環境内の複数.tf/.hclファイル
    ├── nested_block_diff/           # ネストブロック差分
    ├── null_values/                 # null・ゼロ値処理
    ├── partial_existence_diff/      # リソース部分存在差分
    ├── resource_existence_diff/     # リソース存在差分
    ├── single_spec_file/            # 単一.tfspecignoreファイル
    ├── undeclared_diff/             # 未宣言差分
    ├── unicode_and_special_chars/   # Unicode・特殊文字
    └── whitespace_edge_cases/       # 空白文字エッジケース
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

## 処理フロー

```
main.go
  ↓
TfspecApp.CreateRootCommand() (app/cmd/cmd.go)
  ↓
checkコマンド実行
  ↓
AppService.RunCheck() (app/service/service.go)
  ├─ ConfigService.LoadConfig() (app/config/config.go)
  │   └─ 環境ディレクトリの検出・確認
  │
  ├─ AnalyzerService.Analyze() (app/service/analyzer.go)
  │   ├─ parser.LoadIgnoreRules()
  │   ├─ parser.LoadIgnoreRulesWithComments() (コメント付き)
  │   ├─ parseEnvironments() → parser.ParseMultipleFiles() (各環境)
  │   │   └─ parser.ParseEnvFile() (各.tf/.hclファイル)
  │   ├─ differ.Compare() (app/differ/differ.go)
  │   │   ├─ ignoreMatcher.ValidateRules()
  │   │   ├─ compareResourceExistence()
  │   │   ├─ compareAttributes()
  │   │   ├─ compareBlocks()
  │   │   ├─ compareModules()
  │   │   ├─ compareLocals()
  │   │   ├─ compareVariables()
  │   │   ├─ compareOutputs()
  │   │   └─ compareDataSources()
  │   └─ ignoreMatcher.GetWarnings()
  │
  └─ OutputService.OutputResults() (app/service/output.go)
      ├─ reporter.GenerateMarkdown() (app/reporter/reporter.go)
      │   ├─ buildTables()
      │   ├─ fillMissingValues()
      │   ├─ convertToGroupedRows()
      │   └─ buildGroupedMarkdownTable()
      ├─ コンソール出力
      ├─ ファイル出力 (.tfspec/report.md等)
      └─ PrintSummary()
```

## 主要コンポーネント

### cmd/cmd.go
- **TfspecApp** - メインアプリケーションクラス
- Cobraフレームワークによるコマンドライン処理
- checkコマンドの実装
- フラグ:
  - `-v, --verbose` - 詳細出力
  - `-o, --output [FILE]` - 出力ファイル指定
  - `--no-fail` - ドリフト検出時にエラー終了しない
  - `-e, --exclude-dirs` - 除外ディレクトリ
  - `--max-value-length` - テーブル値の最大文字数
  - `--trim-cell` - テーブルセルの前後余白削除

### config/config.go
- **ConfigService** - 設定管理
- `.tfspec` ディレクトリ検出
- 環境ディレクトリの自動検出・解決
- Terraformファイル（.tf/.hcl）の存在確認
- 除外ディレクトリの処理

### parser/parser.go
- **HCLParser** - HCL解析エンジン
- 単一ファイル・複数ファイル対応
- リソース、モジュール、ローカル変数、入力変数、出力値、データソースを対応
- 未解決の変数参照や関数呼び出しをソーステキストとして保存
- **.tfspecignore読み込み機能:**
  - 単一ファイル形式（`.tfspec/.tfspecignore`）
  - 分割ファイル形式（`.tfspec/.tfspecignore/*.txt`）
  - コメント解析（行頭・行末・複数行対応）

### parser/formatter.go
- **ValueFormatter** - cty.Value値のフォーマット
- Markdown形式への変換（改行を`<br>`に）
- 最大文字数制限機能
- リスト・マップ・ネストされた値の整形

### differ/differ.go
- **HCLDiffer** - 環境間差分検出エンジン
- IgnoreMatcherを使用した無視ルール適用
- **比較対象:**
  - リソース存在差分
  - 属性差分（tags含むネスト属性）
  - ネストブロック差分（ingress[0]等）
  - モジュール差分
  - ローカル変数差分
  - 入力変数差分
  - 出力値差分
  - データソース差分
- **属性比較の統一:**
  - ComparisonCallback型によるコールバック関数
  - compareMapAttributes()ヘルパー関数で重複排除

### differ/ignore_matcher.go
- **IgnoreMatcher** - 無視ルール判定エンジン
- IsIgnored() - ルールマッチング（完全一致・プレフィックスマッチ）
- ValidateRules() - ルール検証（警告報告）
- 互換性エイリアス:
  - IsIgnoredWithBlock()
  - IsIgnoredWithBlockAttribute()

### reporter/reporter.go
- **ResultReporter** - Markdownレポート生成
- 階層化テーブル形式（リソースタイプ → リソース名 → 属性）
- ルールコメント付与
- 欠落値の補填（環境別の値補完）
- テーブル値の最大文字数制限
- セル余白削除機能

### types/types.go
- **データ構造定義:**
  - **EnvResource** - Terraformリソース（属性・ネストブロック）
  - **EnvBlock** - ネストブロック（labels・属性）
  - **EnvResources** - 環境全体のリソース集合（リソース・モジュール・ローカル・変数・出力・データソース）
  - **DiffResult** - 差分検出結果（リソース・環境・パス・期待値・実値・無視フラグ）
  - **TableRow / GroupedTableRow** - テーブル表示用

### service/analyzer.go
- **AnalyzerService** - HCL解析・差分検出の統合
- Analyze() - .tfspecignore読み込み → 環境解析 → 差分検出 → 検証

### service/output.go
- **OutputService** - レポート生成・出力
- OutputResults() - Markdown生成 → コンソール出力 → ファイル出力
- PrintSummary() - カウント結果表示

### service/service.go
- **AppService** - コマンド実行統合
- RunCheck() - 設定 → 分析 → 出力 → エラー処理

### interfaces/interfaces.go
- サービスインターフェース定義（DI対応）
- AnalysisResult型定義

## 最近の改善（属性比較リファクタリング）

**コミット**: 3060ce8 (2025年11月11日)

### 追加機能:
1. **ComparisonCallback型** - 属性比較のコールバック関数化
2. **compareMapAttributes()ヘルパー** - 汎用属性比較ロジック一元化
3. **互換性エイリアス** - IsIgnoredWithBlock()等による後方互換性

### 効果:
- コード重複削減（differ.go: -118行）
- 新ブロック型追加時の拡張性向上
- メンテナンス性向上

# 開発関連コマンド
このプロジェクトはGoモジュールです（Go 1.25.1）
- `go build` - ビルド
- `go test` - テスト実行
- `go mod tidy` - 依存関係の整理
- `./tfspec check` - ツール実行

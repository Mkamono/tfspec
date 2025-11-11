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

### 4. コマンドラインフラグ一覧

| フラグ | 説明 | 例 |
|--------|------|-----|
| `-v, --verbose` | 詳細な差分情報を表示 | `tfspec check -v` |
| `-o, --output [FILE]` | 結果をMarkdownファイルに出力（省略時: .tfspec/report.md） | `tfspec check -o custom.md` |
| `--no-fail` | 構成ドリフト検出時もエラー終了しない | `tfspec check --no-fail` |
| `-e, --exclude-dirs` | 除外するディレクトリ（複数指定可） | `tfspec check -e node_modules -e .git` |
| `--max-value-length N` | テーブル値の最大文字数（デフォルト: 200） | `tfspec check --max-value-length 500` |
| `--trim-cell` | テーブルセルの前後余白を削除 | `tfspec check --trim-cell` |

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
├── main.go                    # エントリーポイント
├── go.mod / go.sum           # 依存関係管理
├── app/
│   ├── cmd/cmd.go            # コマンドライン処理（Cobra CLI）
│   ├── config/config.go       # 設定管理・環境ディレクトリ検出
│   ├── differ/
│   │   ├── differ.go         # 差分検出ロジック
│   │   └── ignore_matcher.go # 無視ルール判定
│   ├── interfaces/
│   │   └── interfaces.go     # サービスインターフェース（DI）
│   ├── parser/
│   │   ├── parser.go         # HCL解析・.tfspecignore読み込み
│   │   └── formatter.go      # 値フォーマッティング
│   ├── reporter/
│   │   └── reporter.go       # Markdownレポート生成
│   ├── service/
│   │   ├── analyzer.go       # 解析の統合
│   │   ├── output.go         # 出力処理
│   │   └── service.go        # コマンド実行の統合
│   └── types/
│       └── types.go          # データ構造定義
└── test/                      # テストケース群（26種類）
```

## アーキテクチャの詳細

tfspecの実装はモジュール化されたレイヤーアーキテクチャで構成されています：

### クリーンアーキテクチャレイヤー

1. **コマンド層** (`app/cmd/cmd.go`)
   - Cobraフレームワークによるコマンドライン処理
   - ユーザーインターフェース

2. **サービス層** (`app/service/`)
   - `AppService` - コマンド実行の統合
   - `AnalyzerService` - 解析ロジックの統合
   - `OutputService` - 出力処理の統合

3. **ドメイン層**
   - `Config` - 設定情報
   - `HCLParser` - HCL解析エンジン
   - `HCLDiffer` - 差分検出エンジン
   - `IgnoreMatcher` - ルール判定エンジン
   - `ResultReporter` - レポート生成エンジン

4. **データ層** (`app/types/`)
   - `EnvResource` - リソース
   - `DiffResult` - 差分結果
   - その他のデータ構造

### 設計パターン

- **依存性注入（DI）**: インターフェース駆動設計でテスト可能性を向上
- **コールバック関数**: 属性比較ロジックを統一化
- **ヘルパー関数**: `compareMapAttributes()` でコード重複を削減

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

### コード構成の特徴

- **モジュール化**: 各処理を独立したパッケージに分離
- **インターフェース駆動**: サービス間の結合度を低減
- **テストカバレッジ**: 26種類の包括的なテストケースで動作検証

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

## ドキュメント

詳細な技術ドキュメントは `docs/` ディレクトリを参照してください：

- **[docs/INDEX.md](docs/INDEX.md)** - ドキュメントインデックス（全ドキュメントの概要と読順ガイド）
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - アーキテクチャ詳細（クリーンアーキテクチャ、設計パターン、拡張性）
- **[docs/hcl_deepwiki.md](docs/hcl_deepwiki.md)** - HCL v2ライブラリの設計と仕様
- **[docs/hcl_expression_source_extraction.md](docs/hcl_expression_source_extraction.md)** - Expression からのソーステキスト取得方法
- **[docs/INVESTIGATION_REPORT.md](docs/INVESTIGATION_REPORT.md)** - HCL Expressionの詳細調査報告
- **[docs/parser_improvement_proposal.md](docs/parser_improvement_proposal.md)** - パーサー改善提案（既実装）

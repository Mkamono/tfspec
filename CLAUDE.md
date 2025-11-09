# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

会話、ドキュメント記述は日本語で行うこと

# プロダクト概要
tfspecは、Terraformの環境間構成差分を仕様書として宣言的に管理し、構成ドリフトを自動検出するツールです。

# プロジェクト構造
- `.tfspec/`ディレクトリにすべての設定が集約される
- 仕様書は`.tfspec/spec.hcl`（単一ファイル）または`.tfspec/specs/`（分割ファイル）で管理
- Terraformの構文を再利用して、`# tfspec(...)`構造化コメントで差分の仕様を宣言

## 仕様書の記法
- **簡潔な記法**: 仕様書には差分のある属性・ブロック・リソースのみを記述
- **共通部分の省略**: 全環境で共通の値（ami、name、descriptionなど）は仕様書から省略可能
- **差分のみ焦点**: `# tfspec(...)`コメントが付いた行のみがツールの検証対象
- **存在差分の空ブロック**: 特定環境にのみ存在するブロックは `{}` で表現し、パラメータ記載不要

# アーキテクチャ
## HCLパッケージの使用
tfspecは[HashiCorp Configuration Language (HCL)](https://github.com/hashicorp/hcl)のパッケージを使用してTerraformの構文解析と処理を行います。

### 主要なHCLパッケージ
- `hclsyntax` - HCLネイティブ構文の解析と評価
- `hclparse` - ファイルベースの解析とキャッシュ
- `hcldec` - 仕様ベースのデコード
- `gohcl` - Go構造体ベースのデコード
- `hclwrite` - プログラマティックなHCL生成

### HCL情報モデル
- **属性 (Attributes)**: `key = value` 形式
- **ブロック (Blocks)**: 構造化された設定セクション
- **式評価 (Expression)**: 動的設定値の計算
- go-ctyライブラリとの統合による型安全性

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
│   └── spec.hcl     # 仕様書: instance_type差分を宣言
├── env1/
│   └── main.hcl     # t3.small を使用
├── env2/
│   └── main.hcl     # t3.small を使用
└── env3/
    └── main.hcl     # t3.large を使用（本番環境）
```

仕様書（.tfspec/spec.hcl）:
```hcl
resource "aws_instance" "web" {
  instance_type = "t3.large" # tfspec(env="env3", reason="本番環境のパフォーマンス要件")

  tags = {
    Environment = "env3" # tfspec(env="env3", reason="環境識別用タグ")
    Environment = "env2" # tfspec(env="env2", reason="環境識別用タグ")
    Environment = "env1" # tfspec(env="env1", reason="環境識別用タグ")
  }
}
```

# 開発関連コマンド
このプロジェクトはGoモジュールです（Go 1.25.1）
- `go build` - ビルド
- `go test` - テスト実行
- `go mod tidy` - 依存関係の整理

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

# 開発関連コマンド
このプロジェクトはGoモジュールです（Go 1.25.1）
- `go build` - ビルド
- `go test` - テスト実行
- `go mod tidy` - 依存関係の整理

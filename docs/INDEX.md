# tfspec ドキュメントインデックス

## 概要

このディレクトリには、tfspecプロジェクトに関する技術ドキュメントが集約されています。

## ドキュメント一覧

### 1. ARCHITECTURE.md ★推奨

**対象者**: 開発者、コードリーダー
**目的**: プロジェクト全体のアーキテクチャ理解
**内容:**
- クリーンアーキテクチャレイヤー構成図
- 主要コンポーネント詳細（各パッケージの役割）
- 処理フロー全体
- 設計パターン解説
- テストカバレッジ一覧
- 拡張性ガイド

**読むべきタイミング:**
- プロジェクト初期理解時
- 新しい機能追加前の設計検討時
- コードレビュー時

---

### 2. hcl_deepwiki.md

**対象者**: HCLライブラリの詳細を知りたい開発者
**目的**: HashiCorp HCL v2の設計哲学と仕様
**内容:**
- HCLの設計哲学
- 情報モデル（属性とブロック）
- パッケージアーキテクチャ
- go-ctyとの統合
- 主要ユースケース

**読むべきタイミング:**
- HCL解析処理を変更する時
- 新しいブロック型を追加する時
- ライブラリのupgrade検討時

**出典**: https://deepwiki.com/hashicorp/hcl

---

### 3. hcl_expression_source_extraction.md

**対象者**: HCLファイル解析に携わる開発者
**目的**: Expressionからソーステキスト取得方法の解説
**内容:**
- hclsyntax.Expressionインターフェース
- Range情報の構造
- SliceBytes()による元のソーステキスト取得
- 実装例とサンプルコード

**読むべきタイミング:**
- パーサーの改善・デバッグ時
- 未解決の変数参照の処理方法検討時
- Expression処理を拡張する時

**関連コード**:
- `app/parser/parser.go` の `sourceCache` 機構

---

### 4. INVESTIGATION_REPORT.md

**対象者**: パーサーの実装詳細に興味のある開発者
**目的**: HCL Expressionソーステキスト取得の調査結果報告
**内容:**
- 調査目的と結論
- hclsyntax.Expressionの詳細メソッド一覧
- Range情報の活用方法
- SliceBytes()の実装パターン
- 実装上の注意点

**読むべきタイミング:**
- パーサーの動作仕様を詳しく知りたい時
- Expression処理のトラブルシューティング時

---

### 5. parser_improvement_proposal.md

**対象者**: パーサーの改善を検討している開発者
**目的**: ソーステキスト保持による正確な差分検出の提案
**内容:**
- 現在の問題点（固定文字列での値保存）
- 改善提案（元のソーステキスト保持）
- 実装方法詳細
- メリット・デメリット分析
- 互換性への配慮

**ステータス**: ✅ 既に実装済み
**実装内容**: `sourceCache` 機構と `Range.SliceBytes()` の組み合わせ

---

## ドキュメント選択ガイド

### 「どのドキュメントを読むべき？」フローチャート

```
┌─ tfspecのアーキテクチャ全体を理解したい？
│  └─ YES → ARCHITECTURE.md を読む
│
├─ HCLライブラリそのものについて知りたい？
│  └─ YES → hcl_deepwiki.md を読む
│
├─ Expression から元のテキストを取得する方法は？
│  └─ YES → hcl_expression_source_extraction.md を読む
│
├─ パーサーの詳細な動作を理解したい？
│  └─ YES → INVESTIGATION_REPORT.md を読む
│
└─ パーサーの改善提案について知りたい？
   └─ YES → parser_improvement_proposal.md を読む
```

## ドキュメント読了順序（推奨）

**新規開発者向け:**
1. ARCHITECTURE.md（全体像把握）
2. hcl_deepwiki.md（基礎知識）
3. パッケージ別のコード読み込み

**パーサー開発者向け:**
1. ARCHITECTURE.md
2. hcl_deepwiki.md
3. hcl_expression_source_extraction.md
4. INVESTIGATION_REPORT.md
5. parser_improvement_proposal.md

**機能追加検討者向け:**
1. ARCHITECTURE.md（該当レイヤー部分）
2. 関連ドキュメント
3. テストケース確認

## クイックリファレンス

| 質問 | ドキュメント | 章番号 |
|------|----------|--------|
| パッケージ構成は？ | ARCHITECTURE.md | 3. ドメイン層 |
| ConfigServiceの責務は？ | ARCHITECTURE.md | 3.1 ConfigService |
| HCLファイルはどう解析される？ | ARCHITECTURE.md | 3.2 HCLParser |
| 差分検出ロジックは？ | ARCHITECTURE.md | 3.4 HCLDiffer |
| Expressionとは何か？ | hcl_deepwiki.md | Core Information Model |
| ソーステキスト取得方法は？ | hcl_expression_source_extraction.md | 実装例 |
| パーサーの問題点は？ | parser_improvement_proposal.md | 現状の問題 |

## 関連ファイル

### CLAUDEガイド
- `/CLAUDE.md` - Claude Code用プロジェクトガイド

### ソースコード
- `app/cmd/cmd.go` - コマンドライン処理
- `app/config/config.go` - 設定管理
- `app/parser/parser.go` - HCL解析
- `app/differ/differ.go` - 差分検出
- `app/reporter/reporter.go` - レポート生成

### テスト
- `test/` - 26種類のテストケース

## ドキュメント更新履歴

| 更新日 | ドキュメント | 内容 |
|--------|-----------|------|
| 2025-11-11 | ARCHITECTURE.md | 新規作成、実装に基づく詳細記述 |
| 2025-11-11 | INDEX.md | 新規作成、全ドキュメント索引化 |
| 2025-11-11 | hcl_deepwiki.md | 既存ドキュメント |
| 2025-11-11 | hcl_expression_source_extraction.md | 既存ドキュメント |
| 2025-11-11 | INVESTIGATION_REPORT.md | 既存ドキュメント |
| 2025-11-11 | parser_improvement_proposal.md | 既存ドキュメント（実装済み） |

## フィードバック・改善

ドキュメントの改善提案・質問は以下のいずれかの方法でお願いします：

1. GitHubのIssue（新しい情報追加など）
2. Pull Request（誤りや不明確な点の修正）
3. GitHubのDiscussions（質問や意見交換）

---

**最終更新**: 2025-11-11
**メンテナー**: Claude Code

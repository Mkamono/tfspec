# HCL Expressionからソーステキスト取得の調査レポート

## 調査目的

HCL v2ライブラリ（`hashicorp/hcl/v2`）で、`hclsyntax.Expression`オブジェクトから元のソースコードテキストを取得する方法を調査しました。

## 調査結果サマリー

✅ **結論**: `Expression.Range().SliceBytes(sourceBytes)`で元のソーステキストを取得できる

## 詳細調査結果

### 1. hclsyntax.Expressionインターフェース

```go
type Expression interface {
    Node
    Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics)
    Variables() []hcl.Traversal
    StartRange() hcl.Range
}
```

#### 主要メソッド

| メソッド | 説明 |
|---------|------|
| `Range()` | 式全体のソースコード範囲を返す（Nodeから継承） |
| `Value()` | 式を評価してcty.Valueを返す |
| `Variables()` | 式内で使用されている変数の参照を返す |
| `StartRange()` | 式の開始位置のRangeを返す |

### 2. Range構造体

```go
type Range struct {
    Filename string
    Start, End Pos
}

type Pos struct {
    Line   int  // 行番号（1始まり）
    Column int  // 列番号（1始まり）
    Byte   int  // バイトオフセット（0始まり）
}
```

#### 重要なメソッド

| メソッド | 説明 |
|---------|------|
| `SliceBytes(b []byte) []byte` | バイト列から範囲の部分を切り出す |
| `CanSliceBytes(b []byte) bool` | 安全に切り出せるかチェック |
| `ContainsPos(pos Pos) bool` | 位置が範囲内にあるかチェック |

### 3. ソースバイト列の取得方法

#### 方法A: hcl.File.Bytesフィールド

```go
file, _ := parser.ParseHCLFile("example.hcl")
sourceBytes := file.Bytes
```

#### 方法B: parser.Sources()メソッド

```go
parser := hclparse.NewParser()
file, _ := parser.ParseHCL(src, "example.hcl")
sources := parser.Sources()
sourceBytes := sources["example.hcl"]
```

**注**: `file.Bytes`と`parser.Sources()[filename]`は同一の内容を返します。

### 4. 実装例

#### 基本的な使い方

```go
package main

import (
    "fmt"
    "github.com/hashicorp/hcl/v2/hclparse"
    "github.com/hashicorp/hcl/v2/hclsyntax"
)

func main() {
    parser := hclparse.NewParser()

    src := []byte(`
resource "aws_instance" "web" {
  instance_type = "t3.small"
  ami           = var.ami_id
  count         = length(var.subnets)
}
`)

    file, _ := parser.ParseHCL(src, "example.hcl")
    sourceBytes := file.Bytes

    if syntaxBody, ok := file.Body.(*hclsyntax.Body); ok {
        for _, block := range syntaxBody.Blocks {
            if block.Type == "resource" {
                for name, attr := range block.Body.Attributes {
                    // Rangeを取得
                    exprRange := attr.Expr.Range()

                    // 元のソーステキストを抽出
                    sourceText := string(exprRange.SliceBytes(sourceBytes))

                    fmt.Printf("%s = %s\n", name, sourceText)
                }
            }
        }
    }
}
```

#### 出力例

```
instance_type = "t3.small"
ami = var.ami_id
count = length(var.subnets)
```

### 5. データフロー図

```
┌─────────────────┐
│  HCLソース      │
│  ([]byte)       │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────┐
│ parser.ParseHCL()           │
│ parser.ParseHCLFile()       │
└────────┬────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ hcl.File {                  │
│   Body: hcl.Body            │
│   Bytes: []byte ←────────┐  │
│ }                        │  │
└────────┬─────────────────┼──┘
         │                 │
         ▼                 │
┌─────────────────────┐    │
│ hclsyntax.Body {    │    │
│   Attributes        │    │
│   Blocks            │    │
│ }                   │    │
└────────┬────────────┘    │
         │                 │
         ▼                 │
┌─────────────────────┐    │
│ Attribute {         │    │
│   Name: string      │    │
│   Expr: Expression  │    │
│ }                   │    │
└────────┬────────────┘    │
         │                 │
         ▼                 │
┌─────────────────────┐    │
│ Expression.Range()  │    │
│   → hcl.Range       │    │
└────────┬────────────┘    │
         │                 │
         ▼                 │
┌──────────────────────────┼──┐
│ Range.SliceBytes(Bytes) ←┘  │
│   → 元のソーステキスト       │
└─────────────────────────────┘
```

### 6. 実装の検証

サンプルコード（`examples/expression_extraction.go`）を実行した結果：

```
=== Expression Source Text Extraction Demo ===

【リソース】 aws_instance.web
------------------------------------------------------------
属性: instance_type
  元のソース: "t3.small"
  評価結果: cty.StringVal("t3.small") (型: string)

属性: ami
  元のソース: var.ami_id
  評価結果: [評価不可] Variables not allowed

属性: count
  元のソース: length(var.subnets)
  評価結果: [評価不可] Function calls not allowed
```

✅ **確認事項**:
- リテラル値（`"t3.small"`）は評価可能
- 変数参照（`var.ami_id`）は評価不可だが、元のソーステキストは取得可能
- 関数呼び出し（`length(var.subnets)`）も元のソーステキストは取得可能

## tfspecプロジェクトへの適用

### 現状の問題

現在のparser.goでは、評価できない式を固定文字列で保存：

```go
if diags.HasErrors() {
    resource.Attrs[name] = cty.StringVal("${unresolved_reference}")
}
```

### 問題点

異なる変数参照や関数呼び出しがすべて同じ値として扱われるため、差分検出ができません：

```
env1: ami = var.ami_id        → "${unresolved_reference}"
env2: ami = var.ami_id_v2     → "${unresolved_reference}"
結果: 差分なし（検出漏れ）
```

### 改善案

元のソーステキストを保存することで、正確な差分検出が可能に：

```go
if diags.HasErrors() {
    exprRange := attr.Expr.Range()
    sourceText := string(exprRange.SliceBytes(sourceBytes))
    resource.Attrs[name] = cty.StringVal(sourceText)
}
```

改善後の動作：

```
env1: ami = var.ami_id        → "var.ami_id"
env2: ami = var.ami_id_v2     → "var.ami_id_v2"
結果: 差分あり（正確に検出）
```

## 作成したドキュメント

1. **`docs/hcl_expression_source_extraction.md`**
   - Expressionからソーステキストを取得する詳細な技術ドキュメント
   - インターフェース定義、実装例、注意点などを網羅

2. **`examples/expression_extraction.go`**
   - 実際に動作するサンプルコード
   - 実行して動作確認済み

3. **`docs/parser_improvement_proposal.md`**
   - parser.goの具体的な改善提案
   - 修正箇所、期待される効果、テスト戦略を含む

## 推奨される次のステップ

1. ✅ **調査完了** - この調査レポート
2. ⏭️ **実装** - parser_improvement_proposal.mdの内容を実装
3. ⏭️ **テスト** - 新しいテストケースの追加
4. ⏭️ **検証** - 既存のテストケースで動作確認

## 参考リンク

- [hashicorp/hcl - GitHub](https://github.com/hashicorp/hcl)
- [hclsyntax package - Go Packages](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hclsyntax)
- [hclparse package - Go Packages](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hclparse)
- [hcl package - Go Packages](https://pkg.go.dev/github.com/hashicorp/hcl/v2)

## まとめ

HCL v2ライブラリでは、`Expression.Range().SliceBytes(sourceBytes)`という標準的な方法で、元のソーステキストを簡単に取得できることが確認できました。この方法をtfspecプロジェクトに適用することで、変数参照や関数呼び出しの差分を正確に検出できるようになります。

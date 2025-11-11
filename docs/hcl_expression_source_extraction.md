# HCL v2でExpressionから元のソースコードテキストを取得する方法

## 概要

HCL v2ライブラリ（`github.com/hashicorp/hcl/v2`）では、`hclsyntax.Expression`オブジェクトから元のソースコードテキストを取得する方法が提供されています。これは、Range情報とパース時の元のバイト列を組み合わせることで実現できます。

## 1. hclsyntax.Expressionのインターフェース定義

```go
type Expression interface {
    Node
    Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics)
    Variables() []hcl.Traversal
    StartRange() hcl.Range
}
```

`Expression`は`Node`インターフェースを埋め込んでおり、以下のメソッドを提供します：

### Nodeインターフェース
```go
type Node interface {
    Range() hcl.Range
    // 非公開メソッド
}
```

### 主要メソッド

- **`Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics)`**
  式を評価してcty.Valueを返す

- **`Variables() []hcl.Traversal`**
  式内で使用されている変数の参照を返す

- **`StartRange() hcl.Range`**
  式の開始位置のRangeを返す（`Node.Range()`との競合を避けるため別名）

- **`Range() hcl.Range`** (Nodeから継承)
  式全体のソースコード範囲を返す

## 2. Range情報の構造

```go
type Range struct {
    Filename string
    Start, End Pos
}

type Pos struct {
    Line   int  // 行番号（1始まり）
    Column int  // 列番号（1始まり、Unicodeグラフィームクラスタ単位）
    Byte   int  // バイトオフセット（0始まり）
}
```

### 重要なメソッド

- **`Range.SliceBytes(b []byte) []byte`**
  与えられたバイト列から、Rangeが示す範囲の部分を切り出して返す

- **`Range.CanSliceBytes(b []byte) bool`**
  指定されたバイト列からSliceBytesで安全に切り出せるかチェック

## 3. 元のソーステキストを取得する方法

### 方法1: hclparse.Parserを使う場合

```go
package main

import (
    "fmt"
    "log"

    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclparse"
    "github.com/hashicorp/hcl/v2/hclsyntax"
)

func main() {
    parser := hclparse.NewParser()

    // ファイルをパース
    file, diags := parser.ParseHCLFile("example.hcl")
    if diags.HasErrors() {
        log.Fatal(diags)
    }

    // Parserから元のソースバイト列を取得
    sources := parser.Sources()
    sourceBytes := sources["example.hcl"]

    // Bodyから属性を取得
    if syntaxBody, ok := file.Body.(*hclsyntax.Body); ok {
        for name, attr := range syntaxBody.Attributes {
            // Expressionから元のソーステキストを取得
            exprRange := attr.Expr.Range()
            exprSourceText := string(exprRange.SliceBytes(sourceBytes))

            fmt.Printf("属性 %s の式: %s\n", name, exprSourceText)
        }
    }
}
```

### 方法2: ParseHCL（バイト列から直接パース）を使う場合

```go
package main

import (
    "fmt"
    "log"

    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclparse"
    "github.com/hashicorp/hcl/v2/hclsyntax"
)

func main() {
    parser := hclparse.NewParser()

    // HCLソースコード
    src := []byte(`
resource "aws_instance" "web" {
  instance_type = "t3.small"
  ami           = var.ami_id
  count         = length(var.subnets)

  tags = {
    Name = "web-server"
  }
}
`)

    // バイト列から直接パース
    file, diags := parser.ParseHCL(src, "example.hcl")
    if diags.HasErrors() {
        log.Fatal(diags)
    }

    // file.Bytesフィールドに元のソースが保存されている
    sourceBytes := file.Bytes

    // または parser.Sources() でも取得可能
    // sourceBytes := parser.Sources()["example.hcl"]

    if syntaxBody, ok := file.Body.(*hclsyntax.Body); ok {
        for _, block := range syntaxBody.Blocks {
            if block.Type == "resource" {
                fmt.Printf("リソース: %s.%s\n", block.Labels[0], block.Labels[1])

                for name, attr := range block.Body.Attributes {
                    // Expressionの範囲を取得
                    exprRange := attr.Expr.Range()

                    // 元のソーステキストを抽出
                    exprSourceText := string(exprRange.SliceBytes(sourceBytes))

                    fmt.Printf("  属性 %s の式: %s\n", name, exprSourceText)
                }
            }
        }
    }
}
```

### 出力例

```
リソース: aws_instance.web
  属性 instance_type の式: "t3.small"
  属性 ami の式: var.ami_id
  属性 count の式: length(var.subnets)
  属性 tags の式: {
    Name = "web-server"
  }
```

## 4. 実装例：変数参照や関数呼び出しの元のテキストを保持する

現在のparser.goでは、式を評価できない場合に`"${unresolved_reference}"`という固定文字列を保存していますが、元のソーステキストを保存することで、より正確な差分検出が可能になります。

### 改善例

```go
// parseResourceContent の改善版
func (p *HCLParser) parseResourceContent(body hcl.Body, evalCtx *hcl.EvalContext, resource *types.EnvResource, sourceBytes []byte) error {
    if syntaxBody, ok := body.(*hclsyntax.Body); ok {
        // 属性を解析
        for name, attr := range syntaxBody.Attributes {
            value, diags := attr.Expr.Value(evalCtx)
            if diags.HasErrors() {
                // 元のソーステキストを保存
                exprRange := attr.Expr.Range()
                sourceText := string(exprRange.SliceBytes(sourceBytes))
                resource.Attrs[name] = cty.StringVal(sourceText)
            } else {
                resource.Attrs[name] = value
            }
        }

        // ネストブロックを解析
        for _, block := range syntaxBody.Blocks {
            envBlock := &types.EnvBlock{
                Type:   block.Type,
                Labels: block.Labels,
                Attrs:  make(map[string]cty.Value),
            }

            for name, attr := range block.Body.Attributes {
                value, diags := attr.Expr.Value(evalCtx)
                if diags.HasErrors() {
                    // 元のソーステキストを保存
                    exprRange := attr.Expr.Range()
                    sourceText := string(exprRange.SliceBytes(sourceBytes))
                    envBlock.Attrs[name] = cty.StringVal(sourceText)
                } else {
                    envBlock.Attrs[name] = value
                }
            }

            resource.Blocks[block.Type] = append(resource.Blocks[block.Type], envBlock)
        }

        return nil
    }

    // fallback処理...
    return nil
}
```

## 5. 重要なポイント

### ソースバイト列の取得方法

1. **hclparse.Parser経由**
   ```go
   sources := parser.Sources()
   sourceBytes := sources[filename]
   ```

2. **hcl.File経由**
   ```go
   file, _ := parser.ParseHCL(src, filename)
   sourceBytes := file.Bytes
   ```

### Range.SliceBytes()の使い方

```go
// Expressionの範囲を取得
exprRange := expr.Range()

// 事前にチェック（オプション）
if exprRange.CanSliceBytes(sourceBytes) {
    // ソーステキストを抽出
    sourceText := string(exprRange.SliceBytes(sourceBytes))
}
```

### Start/End位置の詳細

- **Start**: 範囲の開始位置（inclusive、含む）
- **End**: 範囲の終了位置（exclusive、含まない）
- **Byte**: バイトオフセット（0始まり）

実際には、`SliceBytes()`メソッドが`Start.Byte`から`End.Byte`までのバイト列を`sourceBytes[Start.Byte:End.Byte]`として切り出します。

## 6. 注意事項

1. **ファイル名の一致**
   `Range.Filename`とソースバイト列のキーが一致していることを確認してください。

2. **バイト列の保持**
   Parserは内部でソースバイト列を保持するため、複数ファイルを扱う場合はParser.Sources()を使うと便利です。

3. **マルチバイト文字**
   `Range.SliceBytes()`はバイトオフセットで動作するため、UTF-8エンコーディングを正しく扱います。

4. **評価可能な式**
   式が評価可能な場合（リテラル値など）は、`Value()`で得られる`cty.Value`を使い、評価不可能な場合のみ元のソーステキストを使用するのが良いでしょう。

## 7. 実際の動作確認

サンプルコード（`examples/expression_extraction.go`）を実行すると、以下のような出力が得られます：

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

属性: tags
  元のソース: {
    Name = "web-server"
  }
  評価結果: cty.ObjectVal(map[string]cty.Value{"Name":cty.StringVal("web-server")}) (型: object)
```

このように、リテラル値（`"t3.small"`や`tags`のオブジェクト）は評価可能ですが、変数参照（`var.ami_id`）や関数呼び出し（`length(var.subnets)`）は評価できず、元のソーステキストを保持する必要があることが分かります。

## 8. まとめ

### 重要なポイント

- **`hclsyntax.Expression`**は`Range()`メソッドでソースコード範囲を取得できる
- **`hcl.Range.SliceBytes(sourceBytes)`**で元のソーステキストを抽出できる
- ソースバイト列は**`parser.Sources()[filename]`**または**`file.Bytes`**から取得できる
- 変数参照や関数呼び出しなど、評価できない式の元のテキストを保持することで、より正確な差分検出が可能になる

### Expressionインターフェースの構造

```
Expression (interface)
├── Node (埋め込みインターフェース)
│   └── Range() hcl.Range          ← 式全体の範囲
├── Value() (cty.Value, Diagnostics) ← 評価を試みる
├── Variables() []hcl.Traversal    ← 変数参照を取得
└── StartRange() hcl.Range         ← 開始位置の範囲
```

### Range構造体の仕組み

```
Range
├── Filename: string
├── Start: Pos { Line, Column, Byte }  (inclusive)
└── End:   Pos { Line, Column, Byte }  (exclusive)

SliceBytes(b []byte) → b[Start.Byte:End.Byte]
```

### データフロー

```
HCLソース（[]byte）
    ↓
parser.ParseHCL() / parser.ParseHCLFile()
    ↓
hcl.File { Body, Bytes }
    ↓
hclsyntax.Body { Attributes, Blocks }
    ↓
Attribute { Name, Expr }
    ↓
Expression.Range() → hcl.Range
    ↓
Range.SliceBytes(file.Bytes) → 元のソーステキスト
```

### tfspecプロジェクトでの活用

現在のparser.goでは、評価できない式を`"${unresolved_reference}"`という固定文字列で保存していますが、この方法を使えば：

```go
// 改善前
resource.Attrs[name] = cty.StringVal("${unresolved_reference}")

// 改善後（元のソーステキストを保存）
exprRange := attr.Expr.Range()
sourceText := string(exprRange.SliceBytes(sourceBytes))
resource.Attrs[name] = cty.StringVal(sourceText)
```

これにより、以下のような差分をより正確に検出できます：

- `var.ami_id` vs `var.ami_id_v2` → 異なる変数参照として検出
- `length(var.subnets)` vs `length(var.zones)` → 異なる関数呼び出しとして検出
- `"${var.prefix}-web"` vs `"${var.prefix}-app"` → 異なるテンプレート式として検出

# Parser改善提案：元のソーステキスト保持による正確な差分検出

## 現状の問題

現在の`app/parser/parser.go`では、評価できない式（変数参照、関数呼び出しなど）を以下のように固定文字列で保存しています：

```go
value, diags := attr.Expr.Value(evalCtx)
if diags.HasErrors() {
    // 変数参照などの解決不能な値は特別な値として保存
    resource.Attrs[name] = cty.StringVal("${unresolved_reference}")
} else {
    resource.Attrs[name] = value
}
```

### 問題点

この実装では、以下のような式がすべて同じ`"${unresolved_reference}"`として扱われてしまいます：

- `var.ami_id`
- `var.ami_id_v2`
- `length(var.subnets)`
- `length(var.zones)`
- `"${var.prefix}-web"`
- `"${var.prefix}-app"`

結果として、**異なる変数参照や関数呼び出しを持つ環境間の差分が検出できません**。

## 改善提案

`hcl.Range.SliceBytes()`を使用して、元のソーステキストを保持する実装に変更します。

### 実装方法

#### 1. ParseEnvFileの修正

```go
// 標準的なTerraform HCLファイル解析（カスタム関数なし）
func (p *HCLParser) ParseEnvFile(filename string) (*types.EnvResources, error) {
	file, diags := p.parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags
	}

	// 元のソースバイト列を取得
	sourceBytes := file.Bytes

	// ... ブロック解析コード ...

	// 各ブロックタイプを処理
	for _, block := range content.Blocks {
		switch block.Type {
		case "resource":
			envResource := &types.EnvResource{
				Type:   block.Labels[0],
				Name:   block.Labels[1],
				Attrs:  make(map[string]cty.Value),
				Blocks: make(map[string][]*types.EnvBlock),
			}

			// sourceBytesを渡す
			if err := p.parseResourceContent(block.Body, evalCtx, envResource, sourceBytes); err != nil {
				return nil, err
			}

			resources = append(resources, envResource)

		case "module":
			envModule := &types.EnvModule{
				Name:  block.Labels[0],
				Attrs: make(map[string]cty.Value),
			}

			// sourceBytesを渡す
			if err := p.parseSimpleBlockContent(block.Body, evalCtx, envModule.Attrs, sourceBytes); err != nil {
				return nil, err
			}

			modules = append(modules, envModule)

		// ... 他のブロックタイプも同様に修正 ...
		}
	}

	return &types.EnvResources{
		Resources:   resources,
		Modules:     modules,
		Locals:      locals,
		Variables:   variables,
		Outputs:     outputs,
		DataSources: dataSources,
	}, nil
}
```

#### 2. parseResourceContentの修正

```go
// リソース内のコンテンツを再帰的に解析（属性とネストブロック）
func (p *HCLParser) parseResourceContent(body hcl.Body, evalCtx *hcl.EvalContext, resource *types.EnvResource, sourceBytes []byte) error {
	// 低レベルのhclsyntax.Bodyを使用して動的解析
	if syntaxBody, ok := body.(*hclsyntax.Body); ok {
		// 属性を解析
		for name, attr := range syntaxBody.Attributes {
			value, diags := attr.Expr.Value(evalCtx)
			if diags.HasErrors() {
				// 【改善】元のソーステキストを保存
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

			// ネストブロック内の属性を解析
			for name, attr := range block.Body.Attributes {
				value, diags := attr.Expr.Value(evalCtx)
				if diags.HasErrors() {
					// 【改善】元のソーステキストを保存
					exprRange := attr.Expr.Range()
					sourceText := string(exprRange.SliceBytes(sourceBytes))
					envBlock.Attrs[name] = cty.StringVal(sourceText)
				} else {
					envBlock.Attrs[name] = value
				}
			}

			// ブロック型別にグループ化
			resource.Blocks[block.Type] = append(resource.Blocks[block.Type], envBlock)
		}

		return nil
	}

	// fallback: 高レベルAPI（制限がある）
	attrs, diags := body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range attrs {
		value, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			// 【改善】fallback時も元のソーステキストを保存
			exprRange := attr.Expr.Range()
			sourceText := string(exprRange.SliceBytes(sourceBytes))
			resource.Attrs[name] = cty.StringVal(sourceText)
		} else {
			resource.Attrs[name] = value
		}
	}

	return nil
}
```

#### 3. parseSimpleBlockContentの修正

```go
// parseSimpleBlockContent は単純なブロック（module、variable、outputなど）の属性を解析
func (p *HCLParser) parseSimpleBlockContent(body hcl.Body, evalCtx *hcl.EvalContext, attrs map[string]cty.Value, sourceBytes []byte) error {
	if syntaxBody, ok := body.(*hclsyntax.Body); ok {
		// 属性を解析
		for name, attr := range syntaxBody.Attributes {
			value, diags := attr.Expr.Value(evalCtx)
			if diags.HasErrors() {
				// 【改善】元のソーステキストを保存
				exprRange := attr.Expr.Range()
				sourceText := string(exprRange.SliceBytes(sourceBytes))
				attrs[name] = cty.StringVal(sourceText)
			} else {
				attrs[name] = value
			}
		}
		return nil
	}

	// fallback: 高レベルAPI
	attributes, diags := body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range attributes {
		value, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			// 【改善】fallback時も元のソーステキストを保存
			exprRange := attr.Expr.Range()
			sourceText := string(exprRange.SliceBytes(sourceBytes))
			attrs[name] = cty.StringVal(sourceText)
		} else {
			attrs[name] = value
		}
	}

	return nil
}
```

#### 4. parseLocalsContentの修正

```go
// parseLocalsContent はlocalsブロック内のローカル変数を解析
func (p *HCLParser) parseLocalsContent(body hcl.Body, evalCtx *hcl.EvalContext, locals *[]*types.EnvLocal, sourceBytes []byte) error {
	if syntaxBody, ok := body.(*hclsyntax.Body); ok {
		// localsブロック内の属性を解析
		for name, attr := range syntaxBody.Attributes {
			value, diags := attr.Expr.Value(evalCtx)
			if diags.HasErrors() {
				// 【改善】元のソーステキストを保存
				exprRange := attr.Expr.Range()
				sourceText := string(exprRange.SliceBytes(sourceBytes))
				value = cty.StringVal(sourceText)
			}

			envLocal := &types.EnvLocal{
				Name:  name,
				Value: value,
			}

			*locals = append(*locals, envLocal)
		}
		return nil
	}

	// fallback: 高レベルAPI
	attributes, diags := body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range attributes {
		value, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			// 【改善】fallback時も元のソーステキストを保存
			exprRange := attr.Expr.Range()
			sourceText := string(exprRange.SliceBytes(sourceBytes))
			value = cty.StringVal(sourceText)
		}

		envLocal := &types.EnvLocal{
			Name:  name,
			Value: value,
		}

		*locals = append(*locals, envLocal)
	}

	return nil
}
```

#### 5. ParseMultipleFilesの修正

複数ファイルを扱う場合、各ファイルごとに`sourceBytes`を取得する必要があります：

```go
// ParseMultipleFiles は複数の.tf/.hclファイルを結合して解析する
func (p *HCLParser) ParseMultipleFiles(filenames []string) (*types.EnvResources, error) {
	var allResources []*types.EnvResource
	var allModules []*types.EnvModule
	var allLocals []*types.EnvLocal
	var allVariables []*types.EnvVariable
	var allOutputs []*types.EnvOutput
	var allDataSources []*types.EnvData

	// 各ファイルを順番に解析して結合
	for _, filename := range filenames {
		envResources, err := p.ParseEnvFile(filename)
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, envResources.Resources...)
		allModules = append(allModules, envResources.Modules...)
		allLocals = append(allLocals, envResources.Locals...)
		allVariables = append(allVariables, envResources.Variables...)
		allOutputs = append(allOutputs, envResources.Outputs...)
		allDataSources = append(allDataSources, envResources.DataSources...)
	}

	return &types.EnvResources{
		Resources:   allResources,
		Modules:     allModules,
		Locals:      allLocals,
		Variables:   variables,
		Outputs:     allOutputs,
		DataSources: allDataSources,
	}, nil
}
```

この実装では、各ファイルを`ParseEnvFile()`で個別に解析するため、ファイルごとの`sourceBytes`が正しく取得されます。

## 期待される効果

### 1. より正確な差分検出

**改善前：**
```
env1: ami = var.ami_id           → "${unresolved_reference}"
env2: ami = var.ami_id_v2        → "${unresolved_reference}"
結果: 差分なし（誤検出なし）
```

**改善後：**
```
env1: ami = var.ami_id           → "var.ami_id"
env2: ami = var.ami_id_v2        → "var.ami_id_v2"
結果: 差分あり（正確に検出）
```

### 2. 関数呼び出しの差分検出

**改善前：**
```
env1: count = length(var.subnets) → "${unresolved_reference}"
env2: count = length(var.zones)   → "${unresolved_reference}"
結果: 差分なし（誤検出なし）
```

**改善後：**
```
env1: count = length(var.subnets) → "length(var.subnets)"
env2: count = length(var.zones)   → "length(var.zones)"
結果: 差分あり（正確に検出）
```

### 3. テンプレート式の差分検出

**改善前：**
```
env1: name = "${var.prefix}-web"  → "${unresolved_reference}"
env2: name = "${var.prefix}-app"  → "${unresolved_reference}"
結果: 差分なし（誤検出なし）
```

**改善後：**
```
env1: name = "${var.prefix}-web"  → "${var.prefix}-web"
env2: name = "${var.prefix}-app"  → "${var.prefix}-app"
結果: 差分あり（正確に検出）
```

## 実装上の注意点

1. **sourceBytesの取得**
   - `file.Bytes`から取得する（推奨）
   - または`parser.Sources()[filename]`から取得

2. **fallback処理**
   - `hclsyntax.Body`にキャストできない場合も、`attr.Expr.Range()`は使用可能
   - fallback時も同じロジックを適用する

3. **複数ファイルの扱い**
   - `ParseMultipleFiles()`は各ファイルを`ParseEnvFile()`で個別に解析
   - ファイルごとに正しい`sourceBytes`が使用される

4. **パフォーマンス**
   - `Range.SliceBytes()`は単純なスライス操作（`sourceBytes[Start.Byte:End.Byte]`）
   - パフォーマンスへの影響は無視できるレベル

## 後方互換性

この変更は内部実装の改善であり、以下の点で後方互換性が保たれます：

- ✅ 外部APIは変更なし
- ✅ 評価可能な式の動作は変更なし（リテラル値は引き続き評価結果を使用）
- ✅ `.tfspecignore`の構文は変更なし
- ✅ レポート生成ロジックは変更なし

唯一の変化は、**評価できない式の表現がより正確になる**ことです。これは機能強化であり、既存の動作を壊すものではありません。

## テスト戦略

以下のテストケースを追加することを推奨します：

1. **変数参照の差分検出**
   ```
   test/variable_reference_diff/
   ├── .tfspec/
   │   └── .tfspecignore (空)
   ├── env1/
   │   └── main.hcl (ami = var.ami_id)
   └── env2/
       └── main.hcl (ami = var.ami_id_v2)
   ```

2. **関数呼び出しの差分検出**
   ```
   test/function_call_diff/
   ├── .tfspec/
   │   └── .tfspecignore (空)
   ├── env1/
   │   └── main.hcl (count = length(var.subnets))
   └── env2/
       └── main.hcl (count = length(var.zones))
   ```

3. **テンプレート式の差分検出**
   ```
   test/template_expression_diff/
   ├── .tfspec/
   │   └── .tfspecignore (空)
   ├── env1/
   │   └── main.hcl (name = "${var.prefix}-web")
   └── env2/
       └── main.hcl (name = "${var.prefix}-app")
   ```

## まとめ

この改善により、tfspecは**変数参照や関数呼び出しの差分を正確に検出できる**ようになります。実装はシンプルで、HCL v2ライブラリの標準機能（`Range.SliceBytes()`）を使用するだけです。

主な変更点：
- ❌ `"${unresolved_reference}"` （固定文字列）
- ✅ `"var.ami_id"` （元のソーステキスト）

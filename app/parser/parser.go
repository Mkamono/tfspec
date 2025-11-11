package parser

import (
	"os"
	"strings"

	"github.com/Mkamono/tfspec/app/types"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type HCLParser struct {
	parser *hclparse.Parser
	// ファイルのソースバイト列をキャッシュ（Range.SliceBytes用）
	sourceCache map[string][]byte
}

func NewHCLParser() *HCLParser {
	return &HCLParser{
		parser:      hclparse.NewParser(),
		sourceCache: make(map[string][]byte),
	}
}

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
		Variables:   allVariables,
		Outputs:     allOutputs,
		DataSources: allDataSources,
	}, nil
}

// 標準的なTerraform HCLファイル解析（カスタム関数なし）
func (p *HCLParser) ParseEnvFile(filename string) (*types.EnvResources, error) {
	file, diags := p.parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags
	}

	// ソースバイト列をキャッシュに保存（Range.SliceBytes用）
	p.sourceCache[filename] = file.Bytes

	// Terraformの全ブロックタイプを解析
	content, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
			},
			{
				Type:       "module",
				LabelNames: []string{"name"},
			},
			{
				Type:       "locals",
				LabelNames: []string{},
			},
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
			{
				Type:       "output",
				LabelNames: []string{"name"},
			},
			{
				Type:       "data",
				LabelNames: []string{"type", "name"},
			},
		},
	})

	if diags.HasErrors() {
		return nil, diags
	}

	var resources []*types.EnvResource
	var modules []*types.EnvModule
	var locals []*types.EnvLocal
	var variables []*types.EnvVariable
	var outputs []*types.EnvOutput
	var dataSources []*types.EnvData

	// 評価コンテキスト（空）
	evalCtx := &hcl.EvalContext{}

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

			if err := p.parseResourceContent(block.Body, filename, evalCtx, envResource); err != nil {
				return nil, err
			}

			resources = append(resources, envResource)

		case "module":
			envModule := &types.EnvModule{
				Name:  block.Labels[0],
				Attrs: make(map[string]cty.Value),
			}

			if err := p.parseSimpleBlockContent(block.Body, filename, evalCtx, envModule.Attrs); err != nil {
				return nil, err
			}

			modules = append(modules, envModule)

		case "locals":
			if err := p.parseLocalsContent(block.Body, filename, evalCtx, &locals); err != nil {
				return nil, err
			}

		case "variable":
			envVariable := &types.EnvVariable{
				Name:  block.Labels[0],
				Attrs: make(map[string]cty.Value),
			}

			if err := p.parseSimpleBlockContent(block.Body, filename, evalCtx, envVariable.Attrs); err != nil {
				return nil, err
			}

			variables = append(variables, envVariable)

		case "output":
			envOutput := &types.EnvOutput{
				Name:  block.Labels[0],
				Attrs: make(map[string]cty.Value),
			}

			if err := p.parseSimpleBlockContent(block.Body, filename, evalCtx, envOutput.Attrs); err != nil {
				return nil, err
			}

			outputs = append(outputs, envOutput)

		case "data":
			envData := &types.EnvData{
				Type:   block.Labels[0],
				Name:   block.Labels[1],
				Attrs:  make(map[string]cty.Value),
				Blocks: make(map[string][]*types.EnvBlock),
			}

			if err := p.parseResourceContent(block.Body, filename, evalCtx, &types.EnvResource{
				Type:   envData.Type,
				Name:   envData.Name,
				Attrs:  envData.Attrs,
				Blocks: envData.Blocks,
			}); err != nil {
				return nil, err
			}

			dataSources = append(dataSources, envData)
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

// parseAttributesFromBody はHCL Bodyから属性を解析する汎用ヘルパー関数
func (p *HCLParser) parseAttributesFromBody(body hcl.Body, filename string, evalCtx *hcl.EvalContext, attrs map[string]cty.Value) error {
	sourceBytes := p.sourceCache[filename]

	// 低レベルのhclsyntax.Bodyを試す
	if syntaxBody, ok := body.(*hclsyntax.Body); ok {
		for name, attr := range syntaxBody.Attributes {
			value, diags := attr.Expr.Value(evalCtx)
			if diags.HasErrors() {
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
	hlAttrs, diags := body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hlAttrs {
		value, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			exprRange := attr.Expr.Range()
			sourceText := string(exprRange.SliceBytes(sourceBytes))
			attrs[name] = cty.StringVal(sourceText)
		} else {
			attrs[name] = value
		}
	}

	return nil
}

// リソース内のコンテンツを再帰的に解析（属性とネストブロック）
func (p *HCLParser) parseResourceContent(body hcl.Body, filename string, evalCtx *hcl.EvalContext, resource *types.EnvResource) error {
	// 属性を解析
	if err := p.parseAttributesFromBody(body, filename, evalCtx, resource.Attrs); err != nil {
		return err
	}

	// 低レベルのhclsyntax.Bodyでネストブロックを解析
	syntaxBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return nil // fallbackの場合はネストブロック不可
	}

	for _, block := range syntaxBody.Blocks {
		envBlock := &types.EnvBlock{
			Type:   block.Type,
			Labels: block.Labels,
			Attrs:  make(map[string]cty.Value),
		}

		// ネストブロック内の属性を解析
		if err := p.parseAttributesFromBody(block.Body, filename, evalCtx, envBlock.Attrs); err != nil {
			return err
		}

		// ブロック型別にグループ化
		resource.Blocks[block.Type] = append(resource.Blocks[block.Type], envBlock)
	}

	return nil
}

// parseSimpleBlockContent は単純なブロック（module、variable、outputなど）の属性を解析
func (p *HCLParser) parseSimpleBlockContent(body hcl.Body, filename string, evalCtx *hcl.EvalContext, attrs map[string]cty.Value) error {
	return p.parseAttributesFromBody(body, filename, evalCtx, attrs)
}

// parseLocalsContent はlocalsブロック内のローカル変数を解析
func (p *HCLParser) parseLocalsContent(body hcl.Body, filename string, evalCtx *hcl.EvalContext, locals *[]*types.EnvLocal) error {
	// 一時的な属性マップを作成
	attrs := make(map[string]cty.Value)
	if err := p.parseAttributesFromBody(body, filename, evalCtx, attrs); err != nil {
		return err
	}

	// 属性をEnvLocalに変換
	for name, value := range attrs {
		envLocal := &types.EnvLocal{
			Name:  name,
			Value: value,
		}
		*locals = append(*locals, envLocal)
	}

	return nil
}

// LoadIgnoreRules は.tfspecignoreファイルの読み込みを行う
func LoadIgnoreRules(tfspecDir string) ([]string, error) {
	// .tfspecディレクトリが存在しない場合は空のルールを返す
	if tfspecDir == "" {
		return []string{}, nil
	}

	var rules []string

	// 単一ファイル形式をチェック
	if fileRules, err := loadSingleIgnoreFile(tfspecDir + "/.tfspecignore"); err == nil {
		rules = append(rules, fileRules...)
	}

	// ディレクトリ形式をチェック
	if dirRules, err := loadIgnoreDirectory(tfspecDir + "/.tfspecignore/"); err == nil {
		rules = append(rules, dirRules...)
	}

	return rules, nil
}

// コメント付きignoreルールを読み込み（rule -> comment のマップを返す）
func LoadIgnoreRulesWithComments(tfspecDir string) (map[string]string, error) {
	ruleComments := make(map[string]string)

	// .tfspecディレクトリが存在しない場合は空のマップを返す
	if tfspecDir == "" {
		return ruleComments, nil
	}

	// 単一ファイル形式をチェック
	ignoreFile := tfspecDir + "/.tfspecignore"
	if content, err := os.ReadFile(ignoreFile); err == nil {
		parseIgnoreContentWithComments(string(content), ruleComments)
	}

	// ディレクトリ形式をチェック
	ignoreDir := tfspecDir + "/.tfspecignore/"
	if entries, err := os.ReadDir(ignoreDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && entry.Name()[len(entry.Name())-4:] == ".txt" {
				filepath := ignoreDir + entry.Name()
				if content, err := os.ReadFile(filepath); err == nil {
					parseIgnoreContentWithComments(string(content), ruleComments)
				}
			}
		}
	}

	return ruleComments, nil
}

// .tfspecignoreの内容をパースしてルールとコメントを抽出
func parseIgnoreContentWithComments(content string, ruleComments map[string]string) {
	lines := strings.Split(content, "\n")
	var currentComment string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 空行の場合はコメントをリセット
		if line == "" {
			currentComment = ""
			continue
		}

		// コメント行の場合は蓄積（連続する#コメントを結合）
		if strings.HasPrefix(line, "#") {
			comment := strings.TrimPrefix(line, "#")
			comment = strings.TrimSpace(comment)
			if currentComment == "" {
				currentComment = comment
			} else {
				currentComment += "<br>" + comment
			}
			continue
		}

		// 行末コメントをチェック
		var rule, inlineComment string
		if hashIndex := strings.Index(line, "#"); hashIndex != -1 {
			rule = strings.TrimSpace(line[:hashIndex])
			inlineComment = strings.TrimSpace(strings.TrimPrefix(line[hashIndex:], "#"))
		} else {
			rule = line
		}

		// 空のルールはスキップ
		if rule == "" {
			continue
		}

		// コメントを決定（行末コメント優先、なければ直前のコメント）
		var finalComment string
		if inlineComment != "" {
			finalComment = inlineComment
		} else {
			finalComment = currentComment
		}

		ruleComments[rule] = finalComment
		currentComment = "" // コメントをリセット
	}
}

// .tfspecignoreの内容をパースして無視ルールを抽出
func parseIgnoreContent(content string) []string {
	var rules []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 空行やコメント行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 行末コメントを除去
		if hashIndex := strings.Index(line, "#"); hashIndex != -1 {
			line = strings.TrimSpace(line[:hashIndex])
		}

		// 空のルールはスキップ
		if line == "" {
			continue
		}

		rules = append(rules, line)
	}

	return rules
}

// loadSingleIgnoreFile は単一の.tfspecignoreファイルを読み込む
func loadSingleIgnoreFile(filepath string) ([]string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return parseIgnoreContent(string(content)), nil
}

// loadIgnoreDirectory は.tfspecignoreディレクトリ内の.txtファイルを読み込む
func loadIgnoreDirectory(dirPath string) ([]string, error) {
	var rules []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			filepath := dirPath + entry.Name()
			if content, err := os.ReadFile(filepath); err == nil {
				rules = append(rules, parseIgnoreContent(string(content))...)
			}
		}
	}

	return rules, nil
}

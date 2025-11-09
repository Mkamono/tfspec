package app

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type HCLParser struct {
	parser *hclparse.Parser
}

func NewHCLParser() *HCLParser {
	return &HCLParser{
		parser: hclparse.NewParser(),
	}
}

// 標準的なTerraform HCLファイル解析（カスタム関数なし）
func (p *HCLParser) ParseEnvFile(filename string) (*EnvResources, error) {
	file, diags := p.parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags
	}

	// Terraformの resource ブロックを解析
	content, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
			},
		},
	})

	if diags.HasErrors() {
		return nil, diags
	}

	var resources []*EnvResource

	// 評価コンテキスト（空）
	evalCtx := &hcl.EvalContext{}

	// 各リソースブロックを処理
	for _, block := range content.Blocks {
		envResource := &EnvResource{
			Type:   block.Labels[0],
			Name:   block.Labels[1],
			Attrs:  make(map[string]cty.Value),
			Blocks: make(map[string][]*EnvBlock),
		}

		// リソース内のコンテンツを解析（属性とネストブロック）
		if err := p.parseResourceContent(block.Body, evalCtx, envResource); err != nil {
			return nil, err
		}

		resources = append(resources, envResource)
	}

	return &EnvResources{Resources: resources}, nil
}

// リソース内のコンテンツを再帰的に解析（属性とネストブロック）
func (p *HCLParser) parseResourceContent(body hcl.Body, evalCtx *hcl.EvalContext, resource *EnvResource) error {
	// 低レベルのhclsyntax.Bodyを使用して動的解析
	if syntaxBody, ok := body.(*hclsyntax.Body); ok {
		// 属性を解析
		for name, attr := range syntaxBody.Attributes {
			value, diags := attr.Expr.Value(evalCtx)
			if diags.HasErrors() {
				// 変数参照などの解決不能な値は特別な値として保存
				resource.Attrs[name] = cty.StringVal("${unresolved_reference}")
			} else {
				resource.Attrs[name] = value
			}
		}

		// ネストブロックを解析
		for _, block := range syntaxBody.Blocks {
			envBlock := &EnvBlock{
				Type:   block.Type,
				Labels: block.Labels,
				Attrs:  make(map[string]cty.Value),
			}

			// ネストブロック内の属性を解析
			for name, attr := range block.Body.Attributes {
				value, diags := attr.Expr.Value(evalCtx)
				if diags.HasErrors() {
					// 変数参照などの解決不能な値は特別な値として保存
					envBlock.Attrs[name] = cty.StringVal("${unresolved_reference}")
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
			// 変数参照などの解決不能な値は特別な値として保存
			resource.Attrs[name] = cty.StringVal("${unresolved_reference}")
		} else {
			resource.Attrs[name] = value
		}
	}

	return nil
}

// LoadIgnoreRules は.tfspecignoreファイルの読み込みを行う
func LoadIgnoreRules(tfspecDir string) ([]string, error) {
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

package app

import (
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/zclconf/go-cty/cty"
)

// ResultReporter はテーブル形式の結果出力を担当する
type ResultReporter struct {
	formatter *ValueFormatter
}

func NewResultReporter() *ResultReporter {
	return &ResultReporter{
		formatter: NewValueFormatter(),
	}
}

// GenerateMarkdown は差分結果をMarkdownテーブル形式で出力する
func (r *ResultReporter) GenerateMarkdown(diffs []*DiffResult, envNames []string, ruleComments map[string]string, envResources map[string]*EnvResources) string {
	driftTable, ignoredTable := r.buildTables(diffs, envNames, ruleComments, envResources)
	return r.generateMarkdownTables(driftTable, ignoredTable, envNames)
}

// buildTables は差分データをテーブル形式に変換する
func (r *ResultReporter) buildTables(diffs []*DiffResult, envNames []string, ruleComments map[string]string, envResources map[string]*EnvResources) ([]TableRow, []TableRow) {
	driftRows := make(map[string]*TableRow)
	ignoredRows := make(map[string]*TableRow)

	for _, diff := range diffs {
		fullPath := diff.Resource + "." + diff.Path
		key := fullPath

		var targetMap map[string]*TableRow
		if diff.IsIgnored {
			targetMap = ignoredRows
		} else {
			targetMap = driftRows
		}

		row, exists := targetMap[key]
		if !exists {
			row = &TableRow{
				Resource: diff.Resource,
				Path:     diff.Path,
				Values:   make(map[string]string),
				Comment:  "",
			}
			targetMap[key] = row

			if diff.IsIgnored {
				for rule, comment := range ruleComments {
					if strings.Contains(rule, diff.Resource) && strings.Contains(rule, diff.Path) {
						row.Comment = comment
						break
					}
				}
			}
		}

		row.Values[diff.Environment] = r.formatter.FormatValue(diff.Actual)

		if !diff.Expected.IsNull() {
			baseEnv := envNames[0]
			if _, exists := row.Values[baseEnv]; !exists {
				row.Values[baseEnv] = r.formatter.FormatValue(diff.Expected)
			}
		}
	}

	r.fillMissingValues(driftRows, envNames, envResources)
	r.fillMissingValues(ignoredRows, envNames, envResources)

	var driftTable, ignoredTable []TableRow
	for _, row := range driftRows {
		driftTable = append(driftTable, *row)
	}
	for _, row := range ignoredRows {
		ignoredTable = append(ignoredTable, *row)
	}

	sort.Slice(driftTable, func(i, j int) bool {
		return driftTable[i].Resource+"."+driftTable[i].Path < driftTable[j].Resource+"."+driftTable[j].Path
	})
	sort.Slice(ignoredTable, func(i, j int) bool {
		return ignoredTable[i].Resource+"."+ignoredTable[i].Path < ignoredTable[j].Resource+"."+ignoredTable[j].Path
	})

	return driftTable, ignoredTable
}

// fillMissingValues は欠損している環境の値を補填する
func (r *ResultReporter) fillMissingValues(rows map[string]*TableRow, envNames []string, envResources map[string]*EnvResources) {
	for _, row := range rows {
		for _, envName := range envNames {
			if _, exists := row.Values[envName]; exists {
				continue
			}

			if envResource, exists := envResources[envName]; exists {
				resource := r.findResource(envResource, row.Resource)
				if resource != nil {
					value := r.getResourceValue(resource, row.Path)
					if !value.IsNull() {
						row.Values[envName] = r.formatter.FormatValue(value)
					} else {
						row.Values[envName] = ""
					}
				} else {
					row.Values[envName] = ""
				}
			}
		}
	}
}

// findResource はリソースを名前で検索する
func (r *ResultReporter) findResource(envResources *EnvResources, resourceName string) *EnvResource {
	for _, resource := range envResources.Resources {
		fullName := resource.Type + "." + resource.Name
		if fullName == resourceName {
			return resource
		}
	}
	return nil
}

// getResourceValue はリソースから指定パスの値を取得する
func (r *ResultReporter) getResourceValue(resource *EnvResource, path string) cty.Value {
	if path == "" {
		// リソース存在差分の場合：リソースが存在するかどうかを返す
		if resource != nil {
			return cty.BoolVal(true)
		}
		return cty.BoolVal(false)
	}

	if value, exists := resource.Attrs[path]; exists {
		return value
	}

	return cty.NullVal(cty.String)
}

// generateMarkdownTables はMarkdownテーブルを生成する
func (r *ResultReporter) generateMarkdownTables(driftTable, ignoredTable []TableRow, envNames []string) string {
	var md strings.Builder

	md.WriteString("# Tfspec Check Results\n\n")

	// 意図されていない差分テーブル
	if len(driftTable) > 0 {
		md.WriteString("## 🚨 意図されていない差分\n\n")
		md.WriteString(r.buildMarkdownTable(driftTable, envNames, false))
		md.WriteString("\n")
	} else {
		md.WriteString("## ✅ 意図されていない差分\n\n")
		md.WriteString("意図されていない差分は検出されませんでした。\n\n")
	}

	// 無視された差分テーブル
	if len(ignoredTable) > 0 {
		md.WriteString("## 📝 無視された差分（意図的）\n\n")
		md.WriteString(r.buildMarkdownTable(ignoredTable, envNames, true))
		md.WriteString("\n")
	}

	return md.String()
}

// buildMarkdownTable はtablewriterを使用してMarkdownテーブルを生成する
func (r *ResultReporter) buildMarkdownTable(rows []TableRow, envNames []string, includeComment bool) string {
	var buffer strings.Builder
	table := tablewriter.NewTable(&buffer,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)

	// ヘッダー作成
	headers := []string{"該当箇所"}
	headers = append(headers, envNames...)
	if includeComment {
		headers = append(headers, "理由")
	}
	table.Header(headers)

	// 各行のデータ作成
	data := make([][]any, 0, len(rows))
	for _, row := range rows {
		fullPath := row.Resource
		if row.Path != "" {
			fullPath += "." + row.Path
		}

		rowData := []any{fullPath}
		for _, env := range envNames {
			value := row.Values[env]
			if value == "" {
				value = "-"
			}
			// 無視されたリソース存在差分で false の場合は "-" に置換
			if includeComment && row.Path == "" && value == "false" {
				value = "-"
			}
			rowData = append(rowData, value)
		}

		if includeComment {
			comment := row.Comment
			if comment == "" {
				comment = "-"
			}
			rowData = append(rowData, comment)
		}

		data = append(data, rowData)
	}

	table.Bulk(data)
	table.Render()

	return buffer.String()
}
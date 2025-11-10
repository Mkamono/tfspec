package reporter

import (
	"sort"
	"strings"

	"github.com/Mkamono/tfspec/app/parser"
	"github.com/Mkamono/tfspec/app/types"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/zclconf/go-cty/cty"
)

// ResultReporter はテーブル形式の結果出力を担当する
type ResultReporter struct {
	formatter *parser.ValueFormatter
}

func NewResultReporter() *ResultReporter {
	return &ResultReporter{
		formatter: parser.NewValueFormatter(),
	}
}

// GenerateMarkdown は差分結果をMarkdownテーブル形式で出力する
func (r *ResultReporter) GenerateMarkdown(diffs []*types.DiffResult, envNames []string, ruleComments map[string]string, envResources map[string]*types.EnvResources) string {
	driftTable, ignoredTable := r.buildTables(diffs, envNames, ruleComments, envResources)
	return r.generateMarkdownReport(driftTable, ignoredTable, envNames)
}

// buildTables は差分データをテーブル形式に変換する
func (r *ResultReporter) buildTables(diffs []*types.DiffResult, envNames []string, ruleComments map[string]string, envResources map[string]*types.EnvResources) ([]types.TableRow, []types.TableRow) {
	driftRows := make(map[string]*types.TableRow)
	ignoredRows := make(map[string]*types.TableRow)

	// DiffResultをTableRowに変換
	for _, diff := range diffs {
		key := diff.Resource + "." + diff.Path
		var targetMap map[string]*types.TableRow
		if diff.IsIgnored {
			targetMap = ignoredRows
		} else {
			targetMap = driftRows
		}

		row := r.getOrCreateRow(targetMap, key, diff.Resource, diff.Path)
		row.Values[diff.Environment] = r.formatter.FormatValue(diff.Actual)

		// 期待値があればベース環境の値として設定
		if !diff.Expected.IsNull() {
			baseEnv := envNames[0]
			if _, exists := row.Values[baseEnv]; !exists {
				row.Values[baseEnv] = r.formatter.FormatValue(diff.Expected)
			}
		}
	}

	// コメントを付与（無視された項目のみ）
	r.enrichWithComments(ignoredRows, ruleComments)

	// 欠損値を補填
	r.fillMissingValues(driftRows, envNames, envResources)
	r.fillMissingValues(ignoredRows, envNames, envResources)

	return r.mapToSortedSlice(driftRows), r.mapToSortedSlice(ignoredRows)
}

// getOrCreateRow は既存の行を取得するか新しい行を作成する
func (r *ResultReporter) getOrCreateRow(targetMap map[string]*types.TableRow, key, resource, path string) *types.TableRow {
	if row, exists := targetMap[key]; exists {
		return row
	}

	row := &types.TableRow{
		Resource: resource,
		Path:     path,
		Values:   make(map[string]string),
		Comment:  "",
	}
	targetMap[key] = row
	return row
}

// enrichWithComments は無視されたルールにコメントを付与する
func (r *ResultReporter) enrichWithComments(rows map[string]*types.TableRow, ruleComments map[string]string) {
	for _, row := range rows {
		for rule, comment := range ruleComments {
			if strings.Contains(rule, row.Resource) && strings.Contains(rule, row.Path) {
				row.Comment = comment
				break
			}
		}
	}
}

// fillMissingValues は欠損している環境の値を補填する
func (r *ResultReporter) fillMissingValues(rows map[string]*types.TableRow, envNames []string, envResources map[string]*types.EnvResources) {
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
func (r *ResultReporter) findResource(envResources *types.EnvResources, resourceName string) *types.EnvResource {
	for _, resource := range envResources.Resources {
		fullName := resource.Type + "." + resource.Name
		if fullName == resourceName {
			return resource
		}
	}
	return nil
}

// getResourceValue はリソースから指定パスの値を取得する
func (r *ResultReporter) getResourceValue(resource *types.EnvResource, path string) cty.Value {
	if path == "" {
		// リソース存在差分の場合
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

// mapToSortedSlice はマップをソート済みスライスに変換する
func (r *ResultReporter) mapToSortedSlice(rows map[string]*types.TableRow) []types.TableRow {
	result := make([]types.TableRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, *row)
	}

	sort.Slice(result, func(i, j int) bool {
		keyA := result[i].Resource + "." + result[i].Path
		keyB := result[j].Resource + "." + result[j].Path
		return keyA < keyB
	})

	return result
}

// generateMarkdownReport はMarkdownレポート全体を生成する
func (r *ResultReporter) generateMarkdownReport(driftTable, ignoredTable []types.TableRow, envNames []string) string {
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
func (r *ResultReporter) buildMarkdownTable(rows []types.TableRow, envNames []string, includeComment bool) string {
	var buffer strings.Builder
	table := tablewriter.NewTable(&buffer,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)

	// ヘッダー設定
	headers := []string{"該当箇所"}
	headers = append(headers, envNames...)
	if includeComment {
		headers = append(headers, "理由")
	}
	table.Header(headers)

	// データ構築
	data := make([][]any, 0, len(rows))
	for _, row := range rows {
		fullPath := row.Resource
		if row.Path != "" {
			fullPath += "." + row.Path
		}

		rowData := []any{fullPath}
		for _, env := range envNames {
			value := row.Values[env]

			// リソース存在差分の場合のみ、boolean値をアイコンに変換
			if row.Path == "" && isResourceExistenceDiff(row.Resource, value) {
				// 空文字列の場合は「存在しない」として扱う
				if value == "" {
					value = "false"
				}
				switch value {
				case "true":
					value = "✅"
				case "false":
					value = "❌"
				}
			} else {
				// 通常の属性差分の場合は空文字列を"-"に変換
				if value == "" {
					value = "-"
				}
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

// isResourceExistenceDiff はリソース存在差分かどうかを判定する
// リソース存在差分は、リソースの存在自体が差分として検出される場合
func isResourceExistenceDiff(resource, value string) bool {
	// boolean値（true/false）で、かつリソース名が適切な形式の場合のみリソース存在差分として扱う
	// local.*, var.*, output.* のような設定値は除外
	return (value == "true" || value == "false" || value == "") &&
		   strings.Contains(resource, ".") &&
		   !strings.HasPrefix(resource, "local.") &&
		   !strings.HasPrefix(resource, "var.") &&
		   !strings.HasPrefix(resource, "output.")
}
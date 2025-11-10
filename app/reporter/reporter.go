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

		// 値の設定
		if diff.Path == "" && strings.HasPrefix(diff.Resource, "local.") {
			// local存在差分の場合は実際の値を取得
			row.Values[diff.Environment] = r.getLocalValue(envResources[diff.Environment], diff.Resource)
		} else if diff.Path == "" && strings.HasPrefix(diff.Resource, "var.") {
			// variable存在差分の場合は実際の値を取得
			row.Values[diff.Environment] = r.getVariableValue(envResources[diff.Environment], diff.Resource)
		} else {
			row.Values[diff.Environment] = r.formatter.FormatValue(diff.Actual)
		}

		// 期待値があればベース環境の値として設定
		if !diff.Expected.IsNull() {
			baseEnv := envNames[0]
			if _, exists := row.Values[baseEnv]; !exists {
				if diff.Path == "" && strings.HasPrefix(diff.Resource, "local.") {
					// local存在差分の場合は実際の値を取得
					row.Values[baseEnv] = r.getLocalValue(envResources[baseEnv], diff.Resource)
				} else if diff.Path == "" && strings.HasPrefix(diff.Resource, "var.") {
					// variable存在差分の場合は実際の値を取得
					row.Values[baseEnv] = r.getVariableValue(envResources[baseEnv], diff.Resource)
				} else {
					row.Values[baseEnv] = r.formatter.FormatValue(diff.Expected)
				}
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
				if row.Path == "" && strings.HasPrefix(row.Resource, "local.") {
					// local値の補填
					row.Values[envName] = r.getLocalValue(envResource, row.Resource)
				} else if row.Path == "" && strings.HasPrefix(row.Resource, "var.") {
					// variable値の補填
					row.Values[envName] = r.getVariableValue(envResource, row.Resource)
				} else {
					// 通常のリソース処理
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
}

// getLocalValue はlocal値を取得する
func (r *ResultReporter) getLocalValue(envResource *types.EnvResources, resourceName string) string {
	if envResource == nil {
		return "-"
	}

	localName := strings.TrimPrefix(resourceName, "local.")
	for _, local := range envResource.Locals {
		if local.Name == localName {
			return r.formatter.FormatValue(local.Value)
		}
	}
	return "-"
}

// getVariableValue はvariable値を取得する
func (r *ResultReporter) getVariableValue(envResource *types.EnvResources, resourceName string) string {
	if envResource == nil {
		return "-"
	}

	varName := strings.TrimPrefix(resourceName, "var.")
	for _, variable := range envResource.Variables {
		if variable.Name == varName {
			if defaultVal, hasDefault := variable.Attrs["default"]; hasDefault && !defaultVal.IsNull() {
				return r.formatter.FormatValue(defaultVal)
			} else if descVal, hasDesc := variable.Attrs["description"]; hasDesc && !descVal.IsNull() {
				return r.formatter.FormatValue(descVal)
			}
			return "-"
		}
	}
	return "-"
}

// findResource はリソースを名前で検索する（通常のresourceとdataリソース両方に対応）
func (r *ResultReporter) findResource(envResources *types.EnvResources, resourceName string) *types.EnvResource {
	// 通常のリソースを検索
	for _, resource := range envResources.Resources {
		fullName := resource.Type + "." + resource.Name
		if fullName == resourceName {
			return resource
		}
	}

	// dataリソースを検索（data.aws_ami.ubuntu形式）
	if strings.HasPrefix(resourceName, "data.") {
		// "data." プレフィックスを削除
		nameWithoutPrefix := strings.TrimPrefix(resourceName, "data.")
		for _, dataSource := range envResources.DataSources {
			fullName := dataSource.Type + "." + dataSource.Name
			if fullName == nameWithoutPrefix {
				// EnvData を EnvResource として扱えるように変換
				return &types.EnvResource{
					Type:   dataSource.Type,
					Name:   dataSource.Name,
					Attrs:  dataSource.Attrs,
					Blocks: dataSource.Blocks,
				}
			}
		}
	}

	return nil
}

// getResourceValue はリソースから指定パスの値を取得する
func (r *ResultReporter) getResourceValue(resource *types.EnvResource, path string) cty.Value {
	if path == "" {
		// リソース存在差分の場合（真のリソース存在差分のみ）
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
		md.WriteString(r.buildHierarchicalMarkdownTable(driftTable, envNames, false))
		md.WriteString("\n")
	} else {
		md.WriteString("## ✅ 意図されていない差分\n\n")
		md.WriteString("意図されていない差分は検出されませんでした。\n\n")
	}

	// 無視された差分テーブル
	if len(ignoredTable) > 0 {
		md.WriteString("## 📝 無視された差分（意図的）\n\n")
		md.WriteString(r.buildHierarchicalMarkdownTable(ignoredTable, envNames, true))
		md.WriteString("\n")
	}

	return md.String()
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

// buildHierarchicalMarkdownTable は階層化されたMarkdownテーブルを生成する
func (r *ResultReporter) buildHierarchicalMarkdownTable(rows []types.TableRow, envNames []string, includeComment bool) string {
	groupedRows := r.convertToGroupedRows(rows)
	return r.buildGroupedMarkdownTable(groupedRows, envNames, includeComment)
}

// convertToGroupedRows はTableRowを階層化されたGroupedTableRowに変換する
func (r *ResultReporter) convertToGroupedRows(rows []types.TableRow) []types.GroupedTableRow {
	grouped := make([]types.GroupedTableRow, 0, len(rows))

	// リソースタイプとリソース名でソート
	sort.Slice(rows, func(i, j int) bool {
		typeA, nameA := r.parseResourceName(rows[i].Resource)
		typeB, nameB := r.parseResourceName(rows[j].Resource)

		if typeA != typeB {
			return typeA < typeB
		}
		if nameA != nameB {
			return nameA < nameB
		}
		return rows[i].Path < rows[j].Path
	})

	var prevType, prevName string
	for _, row := range rows {
		resourceType, resourceName := r.parseResourceName(row.Resource)

		groupedRow := types.GroupedTableRow{
			ResourceType:      resourceType,
			ResourceName:      resourceName,
			Path:              row.Path,
			Values:            row.Values,
			Comment:           row.Comment,
			IsFirstInGroup:    resourceType != prevType,
			IsFirstInResource: resourceType != prevType || resourceName != prevName,
		}

		grouped = append(grouped, groupedRow)
		prevType, prevName = resourceType, resourceName
	}

	return grouped
}

// parseResourceName はリソース名をタイプと名前に分割する
func (r *ResultReporter) parseResourceName(resource string) (string, string) {
	// local, output, variable, data等の特殊なケースを処理
	if after, found := strings.CutPrefix(resource, "local."); found {
		return "local", after
	}
	if after, found := strings.CutPrefix(resource, "output."); found {
		return "output", after
	}
	if after, found := strings.CutPrefix(resource, "var."); found {
		return "variable", after
	}
	if after, found := strings.CutPrefix(resource, "data."); found {
		// data.aws_instance.web -> type: data.aws_instance, name: web
		parts := strings.SplitN(after, ".", 2)
		if len(parts) >= 2 {
			return "data." + parts[0], parts[1]
		}
		return "data", resource
	}

	// 通常のリソース（aws_instance.web -> type: resource, name: aws_instance.web）
	parts := strings.SplitN(resource, ".", 2)
	if len(parts) >= 2 {
		return "resource", resource
	}
	return resource, ""
}

// buildGroupedMarkdownTable は階層化されたデータでMarkdownテーブルを生成する
func (r *ResultReporter) buildGroupedMarkdownTable(rows []types.GroupedTableRow, envNames []string, includeComment bool) string {
	var buffer strings.Builder
	table := tablewriter.NewTable(&buffer,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)

	// ヘッダー設定
	headers := []string{"リソースタイプ", "リソース名", "属性パス"}
	headers = append(headers, envNames...)
	if includeComment {
		headers = append(headers, "理由")
	}
	table.Header(headers)

	// データ構築
	data := make([][]any, 0, len(rows))
	for _, row := range rows {
		var resourceType, resourceName string

		// グループの最初の行のみリソースタイプを表示
		if row.IsFirstInGroup {
			resourceType = row.ResourceType
		} else {
			resourceType = ""  // 空欄で上のセルと同じグループであることを表現
		}

		// リソースの最初の行のみリソース名を表示
		if row.IsFirstInResource {
			resourceName = row.ResourceName
		} else {
			resourceName = ""  // 空欄で上のセルと同じリソースであることを表現
		}

		// 属性パス（空の場合は空欄）
		pathDisplay := row.Path

		rowData := []any{resourceType, resourceName, pathDisplay}

		// 各環境の値
		for _, env := range envNames {
			value := row.Values[env]

			// リソース存在差分の場合のみ、boolean値をアイコンに変換
			if row.Path == "" && isResourceExistenceDiff(row.ResourceType+"."+row.ResourceName, value) {
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
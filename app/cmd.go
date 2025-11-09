package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
)

type TfspecApp struct {
	parser *HCLParser
	differ *HCLDiffer
}

// Markdownテーブル用のデータ構造
type TableRow struct {
	Resource string
	Path     string
	Values   map[string]string // 環境名 -> 値
	Comment  string            // .tfspecignoreのコメント（無視された差分用）
}

func NewTfspecApp() *TfspecApp {
	return &TfspecApp{
		parser: NewHCLParser(),
		// differ: は実行時にignoreRulesロード後に初期化
	}
}

func (app *TfspecApp) CreateRootCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "tfspec",
		Short: "Terraformの環境間構成差分を自動検出し、意図的差分以外を構成ドリフトとして報告するツール",
		Long: `tfspecは、Terraformの環境間構成差分を自動検出し、「意図的な差分」として宣言されたもの以外を「構成ドリフト」として報告するツールです。

.tfspec/ディレクトリに設定が集約され、意図的な差分は.tfspec/.tfspecignore（単一ファイル）または.tfspec/.tfspecignore/（分割ファイル）で管理されます。
シンプルなリソース名・属性名のリスト形式で記述します。`,
	}

	checkCmd := &cobra.Command{
		Use:   "check [環境ディレクトリ...]",
		Short: "環境間の構成差分をチェックし、意図しない構成ドリフトを検出します",
		Long: `環境間の構成差分をチェックし、意図しない構成ドリフトを検出します。

引数として環境ディレクトリを指定すると、それらの環境のみをチェックします。
引数を省略した場合は、現在のディレクトリから環境ディレクトリを自動検出します。

.tfspecignoreに記載された意図的な差分は除外され、残った差分のみが構成ドリフトとして報告されます。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			outputFile, _ := cmd.Flags().GetString("output")
			outputFlag := cmd.Flags().Changed("output")
			noFail, _ := cmd.Flags().GetBool("no-fail")
			return app.runCheck(args, verbose, outputFile, outputFlag, noFail)
		},
	}

	checkCmd.Flags().BoolP("verbose", "v", false, "詳細な差分情報を表示")
	checkCmd.Flags().StringP("output", "o", "", "結果をMarkdownファイルに出力 (例: -o report.md, -o単体で.tfspec/report.mdに出力)")
	checkCmd.Flags().Lookup("output").NoOptDefVal = ".tfspec/report.md"
	checkCmd.Flags().Bool("no-fail", false, "構成ドリフトが検出されてもエラーコードで終了しない")

	rootCmd.AddCommand(checkCmd)
	return rootCmd
}

func (app *TfspecApp) runCheck(envDirs []string, verbose bool, outputFile string, outputFlag bool, noFail bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("現在のディレクトリを取得できません: %w", err)
	}

	tfspecDir := filepath.Join(cwd, ".tfspec")
	if _, err := os.Stat(tfspecDir); os.IsNotExist(err) {
		return fmt.Errorf(".tfspecディレクトリが見つかりません: %s", tfspecDir)
	}

	// .tfspecignoreルールを読み込み
	ignoreRules, err := LoadIgnoreRules(tfspecDir)
	if err != nil {
		return fmt.Errorf(".tfspecignoreの読み込みに失敗しました: %w", err)
	}

	// コメント付きルールもロード
	ruleComments, err := LoadIgnoreRulesWithComments(tfspecDir)
	if err != nil {
		return fmt.Errorf(".tfspecignoreのコメント読み込みに失敗しました: %w", err)
	}

	// differをignoreRulesで初期化
	app.differ = NewHCLDiffer(ignoreRules)

	if len(envDirs) == 0 {
		envDirs, err = app.detectEnvDirs(cwd)
		if err != nil {
			return fmt.Errorf("環境ディレクトリの検出に失敗しました: %w", err)
		}
	}

	if len(envDirs) == 0 {
		return fmt.Errorf("環境ディレクトリが見つかりません")
	}

	fmt.Printf("環境ディレクトリ: %v\n", envDirs)
	fmt.Printf("無視ルール: %d件\n", len(ignoreRules))

	// 全環境のリソースを解析
	envResources := make(map[string]*EnvResources)
	for _, envDir := range envDirs {
		envName := filepath.Base(envDir)
		envFile := filepath.Join(envDir, "main.hcl")

		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			fmt.Printf("警告: 環境ファイルが見つかりません: %s\n", envFile)
			continue
		}

		envResource, err := app.parser.ParseEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("環境ファイルの解析に失敗しました (%s): %w", envFile, err)
		}

		envResources[envName] = envResource
	}

	// 環境間差分を検出
	diffs, err := app.differ.Compare(envResources)
	if err != nil {
		return fmt.Errorf("差分検出に失敗しました: %w", err)
	}

	// 結果を分類
	var ignoredDiffs, driftDiffs []*DiffResult
	for _, diff := range diffs {
		if diff.IsIgnored {
			ignoredDiffs = append(ignoredDiffs, diff)
		} else {
			driftDiffs = append(driftDiffs, diff)
		}
	}

	// 環境名のリストを作成
	envNames := make([]string, 0, len(envResources))
	for envName := range envResources {
		envNames = append(envNames, envName)
	}
	sort.Strings(envNames)

	// テーブル形式でデータを構築
	driftTable, ignoredTable := app.buildTables(diffs, envNames, ruleComments, envResources)

	// Markdownテーブル出力
	markdownOutput := app.generateMarkdownTables(driftTable, ignoredTable, envNames)

	// 標準出力にMarkdownテーブルを表示
	fmt.Print(markdownOutput)

	// -oオプションが指定されていればファイルに出力
	if outputFlag {
		// .tfspecディレクトリが存在しない場合は作成（.tfspec/report.mdを出力する可能性があるため）
		if strings.Contains(outputFile, ".tfspec/") {
			if err := os.MkdirAll(".tfspec", 0755); err != nil {
				return fmt.Errorf(".tfspecディレクトリの作成に失敗しました: %w", err)
			}
		}
		err := os.WriteFile(outputFile, []byte(markdownOutput), 0644)
		if err != nil {
			return fmt.Errorf("ファイル出力に失敗しました: %w", err)
		}
		fmt.Printf("結果を %s に出力しました。\n", outputFile)
	}

	// 従来の簡潔なサマリー
	fmt.Printf("\n=== サマリー ===\n")
	fmt.Printf("意図的な差分: %d件\n", len(ignoredDiffs))
	fmt.Printf("構成ドリフト: %d件\n", len(driftDiffs))

	if len(driftDiffs) > 0 && !noFail {
		return fmt.Errorf("%d件の構成ドリフトが検出されました", len(driftDiffs))
	}

	return nil
}

func (app *TfspecApp) detectEnvDirs(baseDir string) ([]string, error) {
	var envDirs []string

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if entry.Name() == ".tfspec" {
			continue
		}

		envPath := filepath.Join(baseDir, entry.Name())
		mainFile := filepath.Join(envPath, "main.hcl")

		if _, err := os.Stat(mainFile); err == nil {
			envDirs = append(envDirs, envPath)
		}
	}

	return envDirs, nil
}

func (app *TfspecApp) printDiff(diff *DiffResult) {
	status := "❌"
	if diff.IsIgnored {
		status = "✅"
	}

	fmt.Printf("%s [%s] %s.%s\n", status, diff.Environment, diff.Resource, diff.Path)
	fmt.Printf("   期待値: %s\n", app.formatValue(diff.Expected))
	fmt.Printf("   実際値: %s\n", app.formatValue(diff.Actual))
	if diff.IsIgnored {
		fmt.Printf("   状態: 意図的な差分（.tfspecignoreで無視）\n")
	} else {
		fmt.Printf("   状態: 構成ドリフト（予期しない差分）\n")
	}
	fmt.Println()
}

func (app *TfspecApp) formatValue(val interface{}) string {
	if val == nil {
		return "(存在しない)"
	}

	if ctyVal, ok := val.(cty.Value); ok {
		if ctyVal.IsNull() {
			return "(存在しない)"
		}
		if ctyVal.Type() == cty.String {
			return ctyVal.AsString()
		}
		if ctyVal.Type() == cty.Number {
			if bigFloat := ctyVal.AsBigFloat(); bigFloat.IsInt() {
				if val, accuracy := bigFloat.Int64(); accuracy == 0 {
					return fmt.Sprintf("%d", val)
				}
			}
			return ctyVal.AsBigFloat().String()
		}
		if ctyVal.Type() == cty.Bool {
			if ctyVal.True() {
				return "true"
			}
			return "false"
		}
		// リストまたはタプル型の場合
		if ctyVal.Type().IsListType() || ctyVal.Type().IsTupleType() || ctyVal.Type().IsSetType() {
			var elements []string
			for it := ctyVal.ElementIterator(); it.Next(); {
				_, val := it.Element()
				elements = append(elements, app.formatValue(val))
			}
			return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
		}
		// オブジェクト型またはマップ型の場合
		if ctyVal.Type().IsObjectType() || ctyVal.Type().IsMapType() {
			var pairs []string
			for it := ctyVal.ElementIterator(); it.Next(); {
				key, val := it.Element()
				pairs = append(pairs, fmt.Sprintf("%s: %s", app.formatValue(key), app.formatValue(val)))
			}
			return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
		}
		// その他の型の場合
		return fmt.Sprintf("%s", ctyVal)
	}

	return fmt.Sprintf("%v", val)
}

// 差分データをテーブル形式に変換
func (app *TfspecApp) buildTables(diffs []*DiffResult, envNames []string, ruleComments map[string]string, envResources map[string]*EnvResources) ([]TableRow, []TableRow) {
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

		// 既存の行を取得または新規作成
		row, exists := targetMap[key]
		if !exists {
			row = &TableRow{
				Resource: diff.Resource,
				Path:     diff.Path,
				Values:   make(map[string]string),
				Comment:  "",
			}
			targetMap[key] = row

			// 無視された差分の場合、コメントを設定
			if diff.IsIgnored {
				for rule, comment := range ruleComments {
					if strings.Contains(rule, diff.Resource) && strings.Contains(rule, diff.Path) {
						row.Comment = comment
						break
					}
				}
			}
		}

		// 差分の期待値と実際値を適切に設定
		// Expected（基準環境の値）とActual（現在環境の値）を両方記録
		row.Values[diff.Environment] = app.formatValue(diff.Actual)

		// 基準環境の値も記録（環境名を推測）
		if !diff.Expected.IsNull() {
			// 基準環境は通常最初の環境（アルファベット順で最初）
			baseEnv := envNames[0]
			if _, exists := row.Values[baseEnv]; !exists {
				row.Values[baseEnv] = app.formatValue(diff.Expected)
			}
		}
	}

	// 差分があるパスについて、全環境の値を収集
	app.fillMissingValues(driftRows, envNames, envResources)
	app.fillMissingValues(ignoredRows, envNames, envResources)

	// マップからスライスに変換
	var driftTable, ignoredTable []TableRow
	for _, row := range driftRows {
		driftTable = append(driftTable, *row)
	}
	for _, row := range ignoredRows {
		ignoredTable = append(ignoredTable, *row)
	}

	// ソート
	sort.Slice(driftTable, func(i, j int) bool {
		return driftTable[i].Resource+"."+driftTable[i].Path < driftTable[j].Resource+"."+driftTable[j].Path
	})
	sort.Slice(ignoredTable, func(i, j int) bool {
		return ignoredTable[i].Resource+"."+ignoredTable[i].Path < ignoredTable[j].Resource+"."+ignoredTable[j].Path
	})

	return driftTable, ignoredTable
}

// 欠損している環境の値を補填
func (app *TfspecApp) fillMissingValues(rows map[string]*TableRow, envNames []string, envResources map[string]*EnvResources) {
	for _, row := range rows {
		for _, envName := range envNames {
			// 既に値が設定されている場合はスキップ
			if _, exists := row.Values[envName]; exists {
				continue
			}

			// 環境から該当するリソースと属性を取得
			if envResource, exists := envResources[envName]; exists {
				resource := app.findResource(envResource, row.Resource)
				if resource != nil {
					value := app.getResourceValue(resource, row.Path)
					if !value.IsNull() {
						row.Values[envName] = app.formatValue(value)
					} else {
						row.Values[envName] = "(存在しない)"
					}
				} else {
					row.Values[envName] = "(存在しない)"
				}
			}
		}
	}
}

// リソースを名前で検索
func (app *TfspecApp) findResource(envResources *EnvResources, resourceName string) *EnvResource {
	for _, resource := range envResources.Resources {
		fullName := resource.Type + "." + resource.Name
		if fullName == resourceName {
			return resource
		}
	}
	return nil
}

// リソースから指定パスの値を取得
func (app *TfspecApp) getResourceValue(resource *EnvResource, path string) cty.Value {
	if path == "" {
		return cty.NullVal(cty.String)
	}

	// 属性の値を取得
	if value, exists := resource.Attrs[path]; exists {
		return value
	}

	// ブロックの値も確認（簡易版）
	// より複雑なパス解析が必要な場合は差分検出ロジックから移植
	return cty.NullVal(cty.String)
}

// Markdownテーブルを生成
func (app *TfspecApp) generateMarkdownTables(driftTable, ignoredTable []TableRow, envNames []string) string {
	var md strings.Builder

	md.WriteString("# Tfspec Check Results\n\n")

	// 意図されていない差分テーブル
	if len(driftTable) > 0 {
		md.WriteString("## 🚨 意図されていない差分\n\n")
		md.WriteString("| 該当箇所 |")
		for _, env := range envNames {
			md.WriteString(" " + env + " |")
		}
		md.WriteString("\n")

		md.WriteString("|----------|")
		for range envNames {
			md.WriteString("-------|")
		}
		md.WriteString("\n")

		for _, row := range driftTable {
			fullPath := row.Resource
			if row.Path != "" {
				fullPath += "." + row.Path
			}
			md.WriteString("| " + fullPath + " |")

			for _, env := range envNames {
				value := row.Values[env]
				if value == "" {
					value = "-"
				}
				md.WriteString(" " + value + " |")
			}
			md.WriteString("\n")
		}
		md.WriteString("\n")
	} else {
		md.WriteString("## ✅ 意図されていない差分\n\n")
		md.WriteString("意図されていない差分は検出されませんでした。\n\n")
	}

	// 無視された差分テーブル
	if len(ignoredTable) > 0 {
		md.WriteString("## 📝 無視された差分（意図的）\n\n")
		md.WriteString("| 該当箇所 |")
		for _, env := range envNames {
			md.WriteString(" " + env + " |")
		}
		md.WriteString(" 理由 |\n")

		md.WriteString("|----------|")
		for range envNames {
			md.WriteString("-------|")
		}
		md.WriteString("------|\n")

		for _, row := range ignoredTable {
			fullPath := row.Resource
			if row.Path != "" {
				fullPath += "." + row.Path
			}
			md.WriteString("| " + fullPath + " |")

			for _, env := range envNames {
				value := row.Values[env]
				if value == "" {
					value = "-"
				}
				md.WriteString(" " + value + " |")
			}

			comment := row.Comment
			if comment == "" {
				comment = "-"
			}
			md.WriteString(" " + comment + " |\n")
		}
		md.WriteString("\n")
	}

	return md.String()
}

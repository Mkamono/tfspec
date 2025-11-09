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

func (app *TfspecApp) runCheck(envDirs []string, _ bool, outputFile string, outputFlag bool, noFail bool) error {
	tfspecDir, err := app.setupTfspecDir()
	if err != nil {
		return err
	}

	ignoreRules, ruleComments, err := app.loadIgnoreRules(tfspecDir)
	if err != nil {
		return err
	}

	app.differ = NewHCLDiffer(ignoreRules)

	envDirs, err = app.resolveEnvDirs(envDirs)
	if err != nil {
		return err
	}

	envResources, err := app.parseEnvironments(envDirs)
	if err != nil {
		return err
	}

	diffs, err := app.differ.Compare(envResources)
	if err != nil {
		return fmt.Errorf("差分検出に失敗しました: %w", err)
	}

	ignoredDiffs, driftDiffs := app.classifyDiffs(diffs)
	envNames := app.extractEnvNames(envResources)

	if err := app.outputResults(diffs, envNames, ruleComments, envResources, outputFile, outputFlag); err != nil {
		return err
	}

	app.printSummary(ignoredDiffs, driftDiffs)

	if len(driftDiffs) > 0 && !noFail {
		return fmt.Errorf("%d件の構成ドリフトが検出されました", len(driftDiffs))
	}

	return nil
}

// setupTfspecDir は.tfspecディレクトリの存在を確認し、パスを返す
func (app *TfspecApp) setupTfspecDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("現在のディレクトリを取得できません: %w", err)
	}

	tfspecDir := filepath.Join(cwd, ".tfspec")
	if _, err := os.Stat(tfspecDir); os.IsNotExist(err) {
		return "", fmt.Errorf(".tfspecディレクトリが見つかりません: %s", tfspecDir)
	}

	return tfspecDir, nil
}

// loadIgnoreRules は無視ルールとコメントを読み込む
func (app *TfspecApp) loadIgnoreRules(tfspecDir string) ([]string, map[string]string, error) {
	ignoreRules, err := LoadIgnoreRules(tfspecDir)
	if err != nil {
		return nil, nil, fmt.Errorf(".tfspecignoreの読み込みに失敗しました: %w", err)
	}

	ruleComments, err := LoadIgnoreRulesWithComments(tfspecDir)
	if err != nil {
		return nil, nil, fmt.Errorf(".tfspecignoreのコメント読み込みに失敗しました: %w", err)
	}

	fmt.Printf("無視ルール: %d件\n", len(ignoreRules))
	return ignoreRules, ruleComments, nil
}

// resolveEnvDirs は環境ディレクトリを解決する
func (app *TfspecApp) resolveEnvDirs(envDirs []string) ([]string, error) {
	if len(envDirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("現在のディレクトリを取得できません: %w", err)
		}

		envDirs, err = app.detectEnvDirs(cwd)
		if err != nil {
			return nil, fmt.Errorf("環境ディレクトリの検出に失敗しました: %w", err)
		}
	}

	if len(envDirs) == 0 {
		return nil, fmt.Errorf("環境ディレクトリが見つかりません")
	}

	fmt.Printf("環境ディレクトリ: %v\n", envDirs)
	return envDirs, nil
}

// parseEnvironments は全環境のリソースを解析する
func (app *TfspecApp) parseEnvironments(envDirs []string) (map[string]*EnvResources, error) {
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
			return nil, fmt.Errorf("環境ファイルの解析に失敗しました (%s): %w", envFile, err)
		}

		envResources[envName] = envResource
	}
	return envResources, nil
}

// classifyDiffs は差分を分類する
func (app *TfspecApp) classifyDiffs(diffs []*DiffResult) ([]*DiffResult, []*DiffResult) {
	var ignoredDiffs, driftDiffs []*DiffResult
	for _, diff := range diffs {
		if diff.IsIgnored {
			ignoredDiffs = append(ignoredDiffs, diff)
		} else {
			driftDiffs = append(driftDiffs, diff)
		}
	}
	return ignoredDiffs, driftDiffs
}

// extractEnvNames は環境名リストを抽出してソートする
func (app *TfspecApp) extractEnvNames(envResources map[string]*EnvResources) []string {
	envNames := make([]string, 0, len(envResources))
	for envName := range envResources {
		envNames = append(envNames, envName)
	}
	sort.Strings(envNames)
	return envNames
}

// outputResults は結果を出力する
func (app *TfspecApp) outputResults(diffs []*DiffResult, envNames []string, ruleComments map[string]string, envResources map[string]*EnvResources, outputFile string, outputFlag bool) error {
	reporter := NewResultReporter()
	markdownOutput := reporter.GenerateMarkdown(diffs, envNames, ruleComments, envResources, app)

	fmt.Print(markdownOutput)

	if outputFlag {
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
	return nil
}

// printSummary はサマリーを出力する
func (app *TfspecApp) printSummary(ignoredDiffs, driftDiffs []*DiffResult) {
	fmt.Printf("\n=== サマリー ===\n")
	fmt.Printf("意図的な差分: %d件\n", len(ignoredDiffs))
	fmt.Printf("構成ドリフト: %d件\n", len(driftDiffs))
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


func (app *TfspecApp) formatValue(val interface{}) string {
	if val == nil {
		return ""
	}

	if ctyVal, ok := val.(cty.Value); ok {
		if ctyVal.IsNull() {
			return ""
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


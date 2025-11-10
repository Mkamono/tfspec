package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type TfspecApp struct {
	parser    *HCLParser
	differ    *HCLDiffer
	formatter *ValueFormatter
	envDirs   []string
}

func NewTfspecApp() *TfspecApp {
	return &TfspecApp{
		parser:    NewHCLParser(),
		formatter: NewValueFormatter(),
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
	// 初期化フェーズ
	if err := app.initialize(envDirs); err != nil {
		return err
	}

	// 差分分析フェーズ
	diffs, envResources, ruleComments, envNames, err := app.analyzeDifferences()
	if err != nil {
		return err
	}

	// 結果出力フェーズ
	if err := app.outputResults(diffs, envNames, ruleComments, envResources, outputFile, outputFlag); err != nil {
		return err
	}

	// 結果評価フェーズ
	return app.evaluateResults(diffs, noFail)
}

// initialize は初期化処理を担当する
func (app *TfspecApp) initialize(envDirs []string) error {
	tfspecDir, err := app.setupTfspecDir()
	if err != nil {
		return err
	}

	ignoreRules, _, err := app.loadIgnoreRules(tfspecDir)
	if err != nil {
		return err
	}

	app.differ = NewHCLDiffer(ignoreRules)

	app.envDirs, err = app.resolveEnvDirs(envDirs)
	return err
}

// analyzeDifferences は差分分析処理を担当する
func (app *TfspecApp) analyzeDifferences() ([]*DiffResult, map[string]*EnvResources, map[string]string, []string, error) {
	envResources, err := app.parseEnvironments(app.envDirs)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	diffs, err := app.differ.Compare(envResources)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("差分検出に失敗しました: %w", err)
	}

	// .tfspecignoreの警告を表示
	app.displayIgnoreWarnings()

	// 無視ルールコメント情報を再取得
	tfspecDir, _ := app.setupTfspecDir()
	_, ruleComments, _ := app.loadIgnoreRules(tfspecDir)

	envNames := app.extractEnvNames(envResources)

	return diffs, envResources, ruleComments, envNames, nil
}

// displayIgnoreWarnings は無視ルールの警告を表示する
func (app *TfspecApp) displayIgnoreWarnings() {
	warnings := app.differ.GetIgnoreWarnings()
	for _, warning := range warnings {
		fmt.Printf("⚠️  %s\n", warning)
	}
	if len(warnings) > 0 {
		fmt.Println()
	}
}

// evaluateResults は結果を評価し、適切な終了コードを決定する
func (app *TfspecApp) evaluateResults(diffs []*DiffResult, noFail bool) error {
	ignoredDiffs, driftDiffs := app.classifyDiffs(diffs)
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
		return "", fmt.Errorf("現在のディレクトリを取得できませんでした: %w", err)
	}

	tfspecDir := filepath.Join(cwd, ".tfspec")
	if _, err := os.Stat(tfspecDir); os.IsNotExist(err) {
		return "", fmt.Errorf(".tfspecディレクトリが見つかりません。パス: %s\n" +
			"ヒント: プロジェクトルートで '.tfspec' ディレクトリを作成してください", tfspecDir)
	}

	return tfspecDir, nil
}

// loadIgnoreRules は無視ルールとコメントを読み込む
func (app *TfspecApp) loadIgnoreRules(tfspecDir string) ([]string, map[string]string, error) {
	ignoreRules, err := LoadIgnoreRules(tfspecDir)
	if err != nil {
		return nil, nil, fmt.Errorf(".tfspecignoreファイルの読み込みに失敗しました: %w\n" +
			"ヒント: .tfspec/.tfspecignore ファイルまたは .tfspec/.tfspecignore/ ディレクトリを確認してください", err)
	}

	ruleComments, err := LoadIgnoreRulesWithComments(tfspecDir)
	if err != nil {
		return nil, nil, fmt.Errorf(".tfspecignoreのコメント情報の読み込みに失敗しました: %w", err)
	}

	fmt.Printf("無視ルールを読み込みました: %d件\n", len(ignoreRules))
	return ignoreRules, ruleComments, nil
}

// resolveEnvDirs は環境ディレクトリを解決する
func (app *TfspecApp) resolveEnvDirs(envDirs []string) ([]string, error) {
	if len(envDirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("現在のディレクトリを取得できませんでした: %w", err)
		}

		envDirs, err = app.detectEnvDirs(cwd)
		if err != nil {
			return nil, fmt.Errorf("環境ディレクトリの自動検出に失敗しました: %w", err)
		}
	}

	if len(envDirs) == 0 {
		return nil, fmt.Errorf("環境ディレクトリが見つかりませんでした\n" +
			"ヒント: main.hclファイルを含むディレクトリを作成するか、コマンドライン引数で環境ディレクトリを指定してください")
	}

	fmt.Printf("対象環境: %v\n", envDirs)
	return envDirs, nil
}

// parseEnvironments は全環境のリソースを解析する
func (app *TfspecApp) parseEnvironments(envDirs []string) (map[string]*EnvResources, error) {
	envResources := make(map[string]*EnvResources)
	var skippedFiles []string

	for _, envDir := range envDirs {
		envName := filepath.Base(envDir)
		envFile := filepath.Join(envDir, "main.hcl")

		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			skippedFiles = append(skippedFiles, envFile)
			continue
		}

		envResource, err := app.parser.ParseEnvFile(envFile)
		if err != nil {
			return nil, fmt.Errorf("環境ファイルの解析に失敗しました:\n  ファイル: %s\n  エラー: %w\n" +
				"ヒント: HCL構文を確認してください", envFile, err)
		}

		envResources[envName] = envResource
	}

	if len(skippedFiles) > 0 {
		fmt.Printf("⚠️  以下のファイルをスキップしました: %v\n", skippedFiles)
	}

	if len(envResources) == 0 {
		return nil, fmt.Errorf("解析可能な環境ファイルが見つかりませんでした\n" +
			"ヒント: 各環境ディレクトリに main.hcl ファイルを作成してください")
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
	markdownOutput := reporter.GenerateMarkdown(diffs, envNames, ruleComments, envResources)

	fmt.Print(markdownOutput)

	if outputFlag {
		if strings.Contains(outputFile, ".tfspec/") {
			if err := os.MkdirAll(".tfspec", 0755); err != nil {
				return fmt.Errorf(".tfspecディレクトリの作成に失敗しました:\n  パス: %s\n  エラー: %w",
					".tfspec", err)
			}
		}
		err := os.WriteFile(outputFile, []byte(markdownOutput), 0644)
		if err != nil {
			return fmt.Errorf("レポートファイルの出力に失敗しました:\n  ファイル: %s\n  エラー: %w\n" +
				"ヒント: ディレクトリの書き込み権限を確認してください", outputFile, err)
		}
		fmt.Printf("📄 結果レポートを出力しました: %s\n", outputFile)
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




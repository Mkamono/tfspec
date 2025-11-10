package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Mkamono/tfspec/app/config"
	"github.com/Mkamono/tfspec/app/differ"
	"github.com/Mkamono/tfspec/app/parser"
	"github.com/Mkamono/tfspec/app/types"
	"github.com/Mkamono/tfspec/app/interfaces"
)

// AnalysisResult は分析結果を表す（interfacesパッケージのものを使用）
// type AnalysisResult = interfaces.AnalysisResult

// AnalyzerService は分析処理を担当する
type AnalyzerService struct {
	parser *parser.HCLParser
	differ *differ.HCLDiffer
}

func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{
		parser: parser.NewHCLParser(),
	}
}

// Analyze は環境の分析を実行する
func (s *AnalyzerService) Analyze(config *config.Config) (*interfaces.AnalysisResult, error) {
	// 無視ルールを読み込み
	ignoreRules, ruleComments, err := s.loadIgnoreRules(config.TfspecDir)
	if err != nil {
		return nil, err
	}

	// Differを初期化
	s.differ = differ.NewHCLDiffer(ignoreRules)

	// 環境をパース
	envResources, err := s.parseEnvironments(config.EnvDirs)
	if err != nil {
		return nil, err
	}

	// 差分を検出
	diffs, err := s.differ.Compare(envResources)
	if err != nil {
		return nil, fmt.Errorf("差分検出に失敗しました: %w", err)
	}

	// 警告を表示
	s.displayIgnoreWarnings()

	// 環境名を抽出
	envNames := s.extractEnvNames(envResources)

	return &interfaces.AnalysisResult{
		Diffs:        diffs,
		EnvResources: envResources,
		RuleComments: ruleComments,
		EnvNames:     envNames,
	}, nil
}

// loadIgnoreRules は無視ルールとコメントを読み込む
func (s *AnalyzerService) loadIgnoreRules(tfspecDir string) ([]string, map[string]string, error) {
	ignoreRules, err := parser.LoadIgnoreRules(tfspecDir)
	if err != nil {
		return nil, nil, fmt.Errorf(".tfspecignoreファイルの読み込みに失敗しました: %w\n"+
			"ヒント: .tfspec/.tfspecignore ファイルまたは .tfspec/.tfspecignore/ ディレクトリを確認してください", err)
	}

	ruleComments, err := parser.LoadIgnoreRulesWithComments(tfspecDir)
	if err != nil {
		return nil, nil, fmt.Errorf(".tfspecignoreのコメント情報の読み込みに失敗しました: %w", err)
	}

	fmt.Printf("無視ルールを読み込みました: %d件\n", len(ignoreRules))
	return ignoreRules, ruleComments, nil
}

// parseEnvironments は全環境のリソースを解析する
func (s *AnalyzerService) parseEnvironments(envDirs []string) (map[string]*types.EnvResources, error) {
	envResources := make(map[string]*types.EnvResources)
	var skippedFiles []string

	for _, envDir := range envDirs {
		envName := filepath.Base(envDir)
		envFile := filepath.Join(envDir, "main.hcl")

		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			skippedFiles = append(skippedFiles, envFile)
			continue
		}

		envResource, err := s.parser.ParseEnvFile(envFile)
		if err != nil {
			return nil, fmt.Errorf("環境ファイルの解析に失敗しました:\n  ファイル: %s\n  エラー: %w\n"+
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

// displayIgnoreWarnings は無視ルールの警告を表示する
func (s *AnalyzerService) displayIgnoreWarnings() {
	warnings := s.differ.GetIgnoreWarnings()
	for _, warning := range warnings {
		fmt.Printf("⚠️  %s\n", warning)
	}
	if len(warnings) > 0 {
		fmt.Println()
	}
}

// extractEnvNames は環境名リストを抽出してソートする
func (s *AnalyzerService) extractEnvNames(envResources map[string]*types.EnvResources) []string {
	envNames := make([]string, 0, len(envResources))
	for envName := range envResources {
		envNames = append(envNames, envName)
	}
	sort.Strings(envNames)
	return envNames
}
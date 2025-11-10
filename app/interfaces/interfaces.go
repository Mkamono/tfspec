package interfaces

import (
	"github.com/Mkamono/tfspec/app/config"
	"github.com/Mkamono/tfspec/app/types"
)

// ConfigServiceInterface は設定サービスのインターフェース
type ConfigServiceInterface interface {
	LoadConfig(envDirs []string, verbose, noFail bool, excludeDirs []string) (*config.Config, error)
}

// AnalyzerServiceInterface は分析サービスのインターフェース
type AnalyzerServiceInterface interface {
	Analyze(config *config.Config) (*AnalysisResult, error)
}

// OutputServiceInterface は出力サービスのインターフェース
type OutputServiceInterface interface {
	OutputResults(result *AnalysisResult, outputFile string, outputFlag bool) error
	PrintSummary(diffs []*types.DiffResult) (int, int)
}

// ParserInterface はHCLパーサーのインターフェース
type ParserInterface interface {
	ParseEnvFile(filename string) (*types.EnvResources, error)
}

// DifferInterface は差分検出のインターフェース
type DifferInterface interface {
	Compare(envResources map[string]*types.EnvResources) ([]*types.DiffResult, error)
	GetIgnoreWarnings() []string
}

// ReporterInterface はレポート生成のインターフェース
type ReporterInterface interface {
	GenerateMarkdown(diffs []*types.DiffResult, envNames []string, ruleComments map[string]string, envResources map[string]*types.EnvResources) string
}

// AnalysisResult は分析結果を表す（循環参照回避のためここに定義）
type AnalysisResult struct {
	Diffs        []*types.DiffResult
	EnvResources map[string]*types.EnvResources
	RuleComments map[string]string
	EnvNames     []string
}
package service

import (
	"fmt"

	"github.com/Mkamono/tfspec/app/config"
	"github.com/Mkamono/tfspec/app/interfaces"
)

// AppService はアプリケーションのメインロジックを担当する
type AppService struct {
	configService   interfaces.ConfigServiceInterface
	analyzerService interfaces.AnalyzerServiceInterface
	outputService   interfaces.OutputServiceInterface
}

func NewAppService() *AppService {
	return &AppService{
		configService:   config.NewConfigService(),
		analyzerService: NewAnalyzerService(),
		outputService:   NewOutputService(),
	}
}

// NewAppServiceWithDeps は依存性注入を使用してAppServiceを作成する
func NewAppServiceWithDeps(
	configService interfaces.ConfigServiceInterface,
	analyzerService interfaces.AnalyzerServiceInterface,
	outputService interfaces.OutputServiceInterface,
) *AppService {
	return &AppService{
		configService:   configService,
		analyzerService: analyzerService,
		outputService:   outputService,
	}
}

// RunCheck はcheckコマンドのメインロジックを実行する
func (s *AppService) RunCheck(envDirs []string, verbose bool, outputFile string, outputFlag bool, noFail bool, excludeDirs []string) error {
	// 設定の読み込み
	config, err := s.configService.LoadConfig(envDirs, verbose, noFail, excludeDirs)
	if err != nil {
		return err
	}

	// 分析の実行
	result, err := s.analyzerService.Analyze(config)
	if err != nil {
		return err
	}

	// 結果の出力
	if err := s.outputService.OutputResults(result, outputFile, outputFlag); err != nil {
		return err
	}

	// サマリーの表示と結果評価
	_, driftCount := s.outputService.PrintSummary(result.Diffs)

	if driftCount > 0 && !config.NoFail {
		return fmt.Errorf("%d件の構成ドリフトが検出されました", driftCount)
	}

	return nil
}
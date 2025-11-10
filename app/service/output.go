package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/Mkamono/tfspec/app/reporter"
	"github.com/Mkamono/tfspec/app/types"
	"github.com/Mkamono/tfspec/app/interfaces"
)

// OutputService は結果出力を担当する
type OutputService struct {
	reporter *reporter.ResultReporter
}

func NewOutputService() *OutputService {
	return &OutputService{
		reporter: reporter.NewResultReporter(),
	}
}

// OutputResults は結果を出力する
func (s *OutputService) OutputResults(result *interfaces.AnalysisResult, outputFile string, outputFlag bool) error {
	markdownOutput := s.reporter.GenerateMarkdown(
		result.Diffs,
		result.EnvNames,
		result.RuleComments,
		result.EnvResources,
	)

	// コンソール出力
	fmt.Print(markdownOutput)

	// ファイル出力
	if outputFlag {
		if err := s.writeToFile(markdownOutput, outputFile); err != nil {
			return err
		}
		fmt.Printf("📄 結果レポートを出力しました: %s\n", outputFile)
	}

	return nil
}

// writeToFile はMarkdownをファイルに書き込む
func (s *OutputService) writeToFile(content, outputFile string) error {
	// .tfspecディレクトリが含まれている場合は作成
	if strings.Contains(outputFile, ".tfspec/") {
		if err := os.MkdirAll(".tfspec", 0755); err != nil {
			return fmt.Errorf(".tfspecディレクトリの作成に失敗しました:\n  パス: %s\n  エラー: %w",
				".tfspec", err)
		}
	}

	err := os.WriteFile(outputFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("レポートファイルの出力に失敗しました:\n  ファイル: %s\n  エラー: %w\n"+
			"ヒント: ディレクトリの書き込み権限を確認してください", outputFile, err)
	}

	return nil
}

// PrintSummary はサマリーを出力する
func (s *OutputService) PrintSummary(diffs []*types.DiffResult) (int, int) {
	ignoredCount, driftCount := s.classifyDiffs(diffs)

	fmt.Printf("\n=== サマリー ===\n")
	fmt.Printf("意図的な差分: %d件\n", ignoredCount)
	fmt.Printf("構成ドリフト: %d件\n", driftCount)

	return ignoredCount, driftCount
}

// classifyDiffs は差分を分類してカウントする
func (s *OutputService) classifyDiffs(diffs []*types.DiffResult) (int, int) {
	var ignoredCount, driftCount int
	for _, diff := range diffs {
		if diff.IsIgnored {
			ignoredCount++
		} else {
			driftCount++
		}
	}
	return ignoredCount, driftCount
}
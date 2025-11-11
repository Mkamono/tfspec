package cmd

import (
	"github.com/Mkamono/tfspec/app/service"
	"github.com/spf13/cobra"
)

type TfspecApp struct {
	appService *service.AppService
}

func NewTfspecApp() *TfspecApp {
	return &TfspecApp{
		appService: service.NewAppService(),
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
			excludeDirs, _ := cmd.Flags().GetStringSlice("exclude-dirs")
			maxValueLength, _ := cmd.Flags().GetInt("max-value-length")
			trimCell, _ := cmd.Flags().GetBool("trim-cell")
			return app.appService.RunCheck(args, verbose, outputFile, outputFlag, noFail, excludeDirs, maxValueLength, trimCell)
		},
	}

	checkCmd.Flags().BoolP("verbose", "v", false, "詳細な差分情報を表示")
	checkCmd.Flags().StringP("output", "o", "", "結果をMarkdownファイルに出力 (例: -o report.md, -o単体で.tfspec/report.mdに出力)")
	checkCmd.Flags().Lookup("output").NoOptDefVal = ".tfspec/report.md"
	checkCmd.Flags().Bool("no-fail", false, "構成ドリフトが検出されてもエラーコードで終了しない")
	checkCmd.Flags().StringSliceP("exclude-dirs", "e", []string{}, "除外するディレクトリ名 (例: --exclude-dirs node_modules,vendor)")
	checkCmd.Flags().Int("max-value-length", 200, "テーブルに表示する値の最大文字数 (デフォルト: 200)")
	checkCmd.Flags().Bool("trim-cell", false, "テーブルのセル前後の余白を削除")

	rootCmd.AddCommand(checkCmd)
	return rootCmd
}

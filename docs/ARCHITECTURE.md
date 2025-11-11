# tfspec アーキテクチャ詳細ドキュメント

## 概要

tfspecは、Terraformの環境間構成差分を検出し、「意図的な差分」として宣言されたもの以外を「構成ドリフト」として報告するツールです。Go 1.25.1で実装され、モジュール化されたレイヤーアーキテクチャで構成されています。

## クリーンアーキテクチャレイヤー

```
┌─────────────────────────────────────────────┐
│  コマンド層 (cmd/cmd.go)                    │ ← ユーザーインターフェース
│  - TfspecApp                                 │
│  - Cobraコマンド処理                         │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│  サービス層 (service/)                      │ ← ビジネスロジック統合
│  - AppService                                │
│  - AnalyzerService                           │
│  - OutputService                             │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│  ドメイン層 (config, parser, differ, etc)  │ ← コアビジネスロジック
│  - ConfigService                             │
│  - HCLParser                                 │
│  - HCLDiffer                                 │
│  - IgnoreMatcher                             │
│  - ValueFormatter                            │
│  - ResultReporter                            │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│  データ層 (types/)                          │ ← データ構造定義
│  - EnvResource, EnvBlock                    │
│  - DiffResult                                │
│  - その他の型定義                           │
└─────────────────────────────────────────────┘
```

## 主要コンポーネント詳細

### 1. コマンド層 - cmd/cmd.go

**責務**: ユーザーインターフェース、CLI処理

```go
type TfspecApp struct {
    // Cobraコマンド設定
}

func (a *TfspecApp) CreateRootCommand() *cobra.Command
```

**checkコマンドのフラグ:**
- `-v, --verbose` - 詳細出力
- `-o, --output [FILE]` - ファイル出力（デフォルト: .tfspec/report.md）
- `--no-fail` - エラー終了なし
- `-e, --exclude-dirs` - 除外ディレクトリ
- `--max-value-length N` - テーブル値の最大文字数
- `--trim-cell` - セル余白削除

### 2. サービス層 - service/

**責務**: ビジネスロジックの統合、依存性注入

#### AppService (service.go)
```go
func (s *AppService) RunCheck(
    envDirs []string,
    verbose, outputFile, outputFlag, noFail bool,
    excludeDirs []string,
    maxValueLength int,
    trimCell bool,
) error
```

**処理フロー:**
1. 設定読み込み → ConfigService.LoadConfig()
2. 分析実行 → AnalyzerService.Analyze()
3. 結果出力 → OutputService.OutputResults()
4. サマリー表示
5. エラー処理

#### AnalyzerService (analyzer.go)
```go
func (a *AnalyzerService) Analyze(config *Config) (*AnalysisResult, error)
```

**処理:**
1. `.tfspecignore`ルール読み込み（コメント付き）
2. 全環境のHCLファイル解析
3. 差分検出
4. 検証警告表示

#### OutputService (output.go)
```go
func (o *OutputService) OutputResults(
    result *AnalysisResult,
    outputFile string,
    outputFlag bool,
    maxValueLength int,
    trimCell bool,
) error
```

**処理:**
1. Markdownレポート生成
2. コンソール出力
3. ファイル出力（指定時）

### 3. ドメイン層 - 各パッケージ

#### 3.1 ConfigService (config/config.go)

**責務**: 設定管理、環境ディレクトリ検出

```go
type Config struct {
    TfspecDir   string     // .tfspecディレクトリ
    EnvDirs     []string   // 環境ディレクトリ
    Verbose     bool
    NoFail      bool
    ExcludeDirs []string
}
```

**主要メソッド:**
- `LoadConfig()` - 設定読み込み
- `setupTfspecDir()` - .tfspecディレクトリ検出
- `detectEnvDirs()` - 環境ディレクトリ自動検出
- `hasTerraformFiles()` - HCLファイル存在確認

#### 3.2 HCLParser (parser/parser.go)

**責務**: HCLファイル解析、.tfspecignore読み込み

```go
type HCLParser struct {
    parser      *hclparse.Parser
    sourceCache map[string][]byte  // Range.SliceBytes用
}
```

**主要メソッド:**
- `ParseEnvFile(filename)` - 単一ファイル解析
- `ParseMultipleFiles(filenames)` - 複数ファイル解析
- `LoadIgnoreRules(tfspecDir)` - ルール読み込み
- `LoadIgnoreRulesWithComments(tfspecDir)` - コメント付きルール読み込み

**対応ブロック:**
- resource (Terraformリソース)
- module (モジュール参照)
- variable (入力変数)
- output (出力値)
- locals (ローカル変数)
- data (データソース)

#### 3.3 ValueFormatter (parser/formatter.go)

**責務**: cty.Value値のフォーマット

```go
type ValueFormatter struct {
    useMarkdownLineBreaks bool
    maxLength             int
}
```

**機能:**
- Markdown形式への変換（改行を `<br>` に）
- 最大文字数制限
- リスト・マップ・ネストされた値の整形

#### 3.4 HCLDiffer (differ/differ.go)

**責座**: 環境間差分検出

```go
type HCLDiffer struct {
    ignoreMatcher *IgnoreMatcher
}

// 属性比較のコールバック関数型
type ComparisonCallback func(
    attrName string,
    baseValue, value cty.Value,
    baseExists, exists bool
) *types.DiffResult
```

**比較対象:**
- リソース存在差分
- 属性差分（tags含むネスト属性）
- ネストブロック差分（ingress[0]等）
- モジュール差分
- ローカル変数差分
- 入力変数差分
- 出力値差分
- データソース差分

**主要メソッド:**
- `Compare(envResources)` - 全体差分検出
- `compareResourceExistence()` - リソース存在比較
- `compareAttributes()` - 属性比較
- `compareBlocks()` - ブロック比較
- `compareMapAttributes()` - 汎用属性比較（コールバック使用）

#### 3.5 IgnoreMatcher (differ/ignore_matcher.go)

**責座**: 無視ルール判定、検証

```go
type IgnoreMatcher struct {
    rules          []string
    validatedRules map[string]bool
    warnings       []string
}
```

**主要メソッド:**
- `IsIgnored(resourcePath)` - ルールマッチング判定
- `ValidateRules(envs)` - ルール検証
- `GetWarnings()` - 検証警告取得
- 互換性エイリアス:
  - `IsIgnoredWithBlock()`
  - `IsIgnoredWithBlockAttribute()`

#### 3.6 ResultReporter (reporter/reporter.go)

**責座**: Markdownレポート生成

```go
type ResultReporter struct {
    formatter      *parser.ValueFormatter
    maxValueLength int
    trimCell       bool
}
```

**機能:**
- 階層化テーブル形式（リソースタイプ → リソース名 → 属性）
- ルールコメント付与
- 欠落値の補填
- テーブル値の最大文字数制限
- セル余白削除

**主要メソッド:**
- `GenerateMarkdown()` - Markdown生成
- `buildTables()` - テーブル構築
- `fillMissingValues()` - 欠落値補填
- `buildGroupedMarkdownTable()` - 階層化テーブル生成

### 4. データ層 - types/types.go

**主要型:**

```go
// Terraformリソース
type EnvResource struct {
    Type   string                   // aws_instance
    Name   string                   // web
    Attrs  map[string]cty.Value     // 属性
    Blocks map[string][]*EnvBlock   // ネストブロック
}

// ネストブロック
type EnvBlock struct {
    Type   string              // ingress
    Labels []string            // ラベル
    Attrs  map[string]cty.Value // ブロック属性
}

// 環境全体のリソース集合
type EnvResources struct {
    Resources   []*EnvResource
    Modules     []*EnvModule
    Locals      []*EnvLocal
    Variables   []*EnvVariable
    Outputs     []*EnvOutput
    DataSources []*EnvData
}

// 差分検出結果
type DiffResult struct {
    Resource    string      // aws_instance.web
    Environment string      // env1
    Path        string      // instance_type / tags.Environment
    Expected    cty.Value   // 基準環境の値
    Actual      cty.Value   // 比較環境の値
    IsIgnored   bool        // 無視フラグ
}

// テーブル表示用
type TableRow struct {
    Resource string              // リソース名
    Path     string              // 属性パス
    Values   map[string]string   // 環境名 → 値（表示用）
    Comment  string              // ルールコメント
}
```

## 処理フロー全体

```
main.go
  ↓
TfspecApp.CreateRootCommand()
  ↓
checkコマンド実行
  ↓
AppService.RunCheck()
  ├─ 1. ConfigService.LoadConfig()
  │    └─ 環境ディレクトリ検出
  │
  ├─ 2. AnalyzerService.Analyze()
  │    ├─ parser.LoadIgnoreRules()
  │    ├─ parser.LoadIgnoreRulesWithComments()
  │    ├─ parseEnvironments()
  │    │   └─ parser.ParseMultipleFiles()
  │    │       └─ parser.ParseEnvFile() (各.tf/.hclファイル)
  │    ├─ differ.Compare()
  │    │   ├─ ignoreMatcher.ValidateRules()
  │    │   ├─ compareResourceExistence()
  │    │   ├─ compareAttributes()
  │    │   ├─ compareBlocks()
  │    │   ├─ compareModules()
  │    │   ├─ compareLocals()
  │    │   ├─ compareVariables()
  │    │   ├─ compareOutputs()
  │    │   └─ compareDataSources()
  │    └─ ignoreMatcher.GetWarnings()
  │
  └─ 3. OutputService.OutputResults()
       ├─ reporter.GenerateMarkdown()
       │   ├─ buildTables()
       │   ├─ fillMissingValues()
       │   ├─ convertToGroupedRows()
       │   └─ buildGroupedMarkdownTable()
       ├─ コンソール出力
       ├─ ファイル出力
       └─ PrintSummary()
```

## 設計パターン

### 1. 依存性注入（DI）

インターフェース駆動設計により、サービス間の結合度を低減：

```go
type ConfigServiceInterface interface {
    LoadConfig(envDirs []string, verbose, noFail bool, excludeDirs []string) (*Config, error)
}

type AnalyzerServiceInterface interface {
    Analyze(config *Config) (*AnalysisResult, error)
}

type OutputServiceInterface interface {
    OutputResults(...) error
    PrintSummary(...) (int, int)
}
```

**利点:**
- テスト可能性向上
- モックの挿入が容易
- 実装の入れ替えが可能

### 2. コールバック関数型

属性比較ロジックを統一化：

```go
type ComparisonCallback func(
    attrName string,
    baseValue, value cty.Value,
    baseExists, exists bool
) *types.DiffResult

func (d *HCLDiffer) compareMapAttributes(
    baseMap, targetMap map[string]cty.Value,
    callback ComparisonCallback,
) []*types.DiffResult
```

**利点:**
- コード重複削減
- 新しいブロック型の追加が容易
- 比較ロジックの一元化

### 3. ヘルパー関数パターン

複雑な処理を小さな責務に分割：

```go
- findTerraformFiles()   // ファイル検索
- fillMissingValues()    // 欠落値補填
- parseResourceName()    // リソース名解析
- trimCellPadding()      // セル余白削除
```

## 最近の改善（コミット 3060ce8）

### 属性比較リファクタリング

**追加機能:**

1. **ComparisonCallback型** - 属性比較のコールバック関数化
2. **compareMapAttributes()ヘルパー** - 汎用属性比較ロジック一元化
3. **互換性エイリアス** - IsIgnoredWithBlock()等による後方互換性

**効果:**
- differ.go: -118行の削減
- コード重複削減
- 新ブロック型追加時の拡張性向上
- メンテナンス性向上

## テストカバレッジ

26種類の包括的なテストケース：

| カテゴリ | テスト数 | 対象 |
|----------|---------|------|
| 基本機能 | 8 | 属性差分、コメント、ファイル処理 |
| リソース | 4 | 存在差分、部分差分、データソース |
| ブロック | 4 | ネストブロック、深いネスト |
| 複合 | 4 | 複数ファイル、複数仕様書、モジュール |
| エッジケース | 6 | 空ファイル、null値、Unicode、特殊文字 |

## 依存ライブラリ

```
github.com/hashicorp/hcl/v2 v2.21.0  - HCL解析
github.com/zclconf/go-cty v1.14.4    - Terraform値型
github.com/olekukonko/tablewriter v1.1.1  - Markdownテーブル
github.com/spf13/cobra v1.8.0        - CLIフレームワーク
```

## 拡張性

### 新しいブロック型の追加

1. `parseXxxContent()` を types.go で定義
2. `EnvResources` に新フィールド追加
3. `compareXxx()` を differ.go で実装（ComparisonCallbackを活用）
4. `fillMissingValuesForXxx()` を reporter.go で実装

### 新しいコマンドの追加

1. `cmd/cmd.go` に `CreateXxxCommand()` を追加
2. `service/` に `XxxService` を作成
3. `AppService` に `RunXxx()` メソッド追加

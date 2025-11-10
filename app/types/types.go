package types

import (
	"github.com/zclconf/go-cty/cty"
)

// シンプルな構造体定義（新しい.tfspecignore設計用）

type EnvResource struct {
	Type   string
	Name   string
	Attrs  map[string]cty.Value
	Blocks map[string][]*EnvBlock
}

// 新しいブロックタイプ用の構造体
type EnvModule struct {
	Name   string
	Attrs  map[string]cty.Value
}

type EnvLocal struct {
	Name  string
	Value cty.Value
}

type EnvVariable struct {
	Name   string
	Attrs  map[string]cty.Value
}

type EnvOutput struct {
	Name   string
	Attrs  map[string]cty.Value
}

type EnvData struct {
	Type   string
	Name   string
	Attrs  map[string]cty.Value
	Blocks map[string][]*EnvBlock
}

type EnvResources struct {
	Resources []*EnvResource
	Modules   []*EnvModule
	Locals    []*EnvLocal
	Variables []*EnvVariable
	Outputs   []*EnvOutput
	DataSources []*EnvData
}

type EnvBlock struct {
	Type   string
	Labels []string
	Attrs  map[string]cty.Value
}

type DiffResult struct {
	Resource    string
	Environment string
	Path        string
	Expected    cty.Value
	Actual      cty.Value
	IsIgnored   bool // 新設計：.tfspecignoreに記載されているかどうか
}

// TableRow はMarkdownテーブル用のデータ構造
type TableRow struct {
	Resource string
	Path     string
	Values   map[string]string // 環境名 -> 値
	Comment  string            // .tfspecignoreのコメント（無視された差分用）
}

// GroupedTableRow は階層化されたテーブル用のデータ構造
type GroupedTableRow struct {
	ResourceType string    // リソースタイプ (aws_instance, local, output等)
	ResourceName string    // リソース名 (web, db等)
	Path         string    // 属性パス
	Values       map[string]string // 環境名 -> 値
	Comment      string    // .tfspecignoreのコメント（無視された差分用）
	IsFirstInGroup bool    // グループの最初の行かどうか
	IsFirstInResource bool // リソースの最初の行かどうか
}


package app

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

type EnvResources struct {
	Resources []*EnvResource
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


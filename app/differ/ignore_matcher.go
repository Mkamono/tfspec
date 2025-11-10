package differ

import (
	"fmt"
	"strings"

	"github.com/Mkamono/tfspec/app/types"
)

// IgnoreMatcher は無視ルールの判定を担当する
type IgnoreMatcher struct {
	rules          []string
	validatedRules map[string]bool
	warnings       []string
}

func NewIgnoreMatcher(rules []string) *IgnoreMatcher {
	return &IgnoreMatcher{
		rules:          rules,
		validatedRules: make(map[string]bool),
		warnings:       make([]string, 0),
	}
}

// IsIgnored はリソース・属性パスが無視ルールにマッチするかチェックする
func (m *IgnoreMatcher) IsIgnored(resourcePath string) bool {
	for _, rule := range m.rules {
		if rule == resourcePath {
			return true
		}
		if m.isChildPath(resourcePath, rule) {
			return true
		}
		if m.matchesPattern(rule, resourcePath) {
			return true
		}
	}
	return false
}

// IsIgnoredWithBlock はブロック情報を考慮した無視判定を行う
func (m *IgnoreMatcher) IsIgnoredWithBlock(resourcePath string, block *types.EnvBlock, blockType string) bool {
	return m.IsIgnored(resourcePath)
}

// IsIgnoredWithBlockAttribute はブロック属性の無視判定を行う
func (m *IgnoreMatcher) IsIgnoredWithBlockAttribute(resourcePath string, block *types.EnvBlock, blockType string) bool {
	return m.IsIgnored(resourcePath)
}

// isChildPath は指定されたパスが親ルールの子パスかどうかをチェックする
func (m *IgnoreMatcher) isChildPath(resourcePath, parentRule string) bool {
	return strings.HasPrefix(resourcePath, parentRule+".")
}

// matchesPattern はより柔軟なパターンマッチングを行う（将来の拡張用）
func (m *IgnoreMatcher) matchesPattern(rule, resourcePath string) bool {
	return false
}

// ValidateRules は与えられたリソースデータに対して無視ルールの検証を行う
func (m *IgnoreMatcher) ValidateRules(envs map[string]map[string]*types.EnvResource) {
	for _, rule := range m.rules {
		if m.isValidRule(rule, envs) {
			m.validatedRules[rule] = true
		} else {
			m.warnings = append(m.warnings, fmt.Sprintf("無視ルール '%s' は実際のリソース構成に存在しません", rule))
		}
	}
}

// GetWarnings は検証で発見された警告を返す
func (m *IgnoreMatcher) GetWarnings() []string {
	return m.warnings
}

// isValidRule は無視ルールが実際のリソース構成に存在するかチェックする
func (m *IgnoreMatcher) isValidRule(rule string, envs map[string]map[string]*types.EnvResource) bool {
	parts := strings.Split(rule, ".")
	if len(parts) < 2 {
		return false
	}

	resourceType := parts[0]
	resourceName := parts[1]
	resourceKey := resourceType + "." + resourceName

	// 少なくとも1つの環境でリソースが存在するかチェック
	for _, envResources := range envs {
		if resource, exists := envResources[resourceKey]; exists {
			if len(parts) == 2 {
				// リソース自体の指定
				return true
			}

			// 属性の存在チェック
			attributePath := strings.Join(parts[2:], ".")
			return m.hasAttribute(resource, attributePath)
		}
	}

	return false
}

// hasAttribute は指定されたリソースに属性が存在するかチェックする
func (m *IgnoreMatcher) hasAttribute(resource *types.EnvResource, attributePath string) bool {
	parts := strings.Split(attributePath, ".")

	// 単純な属性チェック（ここでは基本的な属性のみチェック）
	if len(parts) == 1 {
		// 基本属性の存在チェック
		attr := parts[0]
		_, exists := resource.Attrs[attr]
		return exists
	}

	// ネストした属性の場合（tags.Environment等）
	if len(parts) >= 2 && parts[0] == "tags" {
		if tagsVal, exists := resource.Attrs["tags"]; exists && tagsVal.Type().IsObjectType() {
			tagKey := parts[1]
			tagsMap := tagsVal.AsValueMap()
			_, tagExists := tagsMap[tagKey]
			return tagExists
		}
	}

	return false
}
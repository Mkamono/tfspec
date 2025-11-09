package app

import (
	"strings"
)

// IgnoreMatcher は無視ルールの判定を担当する
type IgnoreMatcher struct {
	rules []string
}

func NewIgnoreMatcher(rules []string) *IgnoreMatcher {
	return &IgnoreMatcher{rules: rules}
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
func (m *IgnoreMatcher) IsIgnoredWithBlock(resourcePath string, block *EnvBlock, blockType string) bool {
	return m.IsIgnored(resourcePath)
}

// IsIgnoredWithBlockAttribute はブロック属性の無視判定を行う
func (m *IgnoreMatcher) IsIgnoredWithBlockAttribute(resourcePath string, block *EnvBlock, blockType string) bool {
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
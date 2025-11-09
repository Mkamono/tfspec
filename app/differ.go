package app

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

type HCLDiffer struct {
	ignoreRules []string
}

func NewHCLDiffer(ignoreRules []string) *HCLDiffer {
	return &HCLDiffer{
		ignoreRules: ignoreRules,
	}
}

// 環境間差分を検出し、.tfspecignoreルールでフィルタリング
func (d *HCLDiffer) Compare(envResources map[string]*EnvResources) ([]*DiffResult, error) {
	var results []*DiffResult

	// 環境名のスライスを作成（決定的な順序でソート）
	var envNames []string
	for envName := range envResources {
		envNames = append(envNames, envName)
	}
	sort.Strings(envNames)

	if len(envNames) < 2 {
		return results, nil // 比較対象が1つ以下の場合は差分なし
	}

	// 基準環境を1つ目とし、他の環境と比較
	baseEnv := envNames[0]
	baseEnvResources := envResources[baseEnv]

	for i := 1; i < len(envNames); i++ {
		env := envNames[i]
		envResourceList := envResources[env]

		// リソース存在差分を検出
		existenceDiffs := d.compareResourceExistence(baseEnvResources, envResourceList, baseEnv, env)
		results = append(results, existenceDiffs...)

		// 共通リソースの属性・ブロック差分を検出
		for _, baseResource := range baseEnvResources.Resources {
			for _, resource := range envResourceList.Resources {
				// リソース種別・名前が同じかチェック
				if baseResource.Type == resource.Type && baseResource.Name == resource.Name {
					// 属性を比較
					envDiffs := d.compareAttributes(baseResource, resource, baseEnv, env)
					results = append(results, envDiffs...)

					// ネストブロックを比較
					blockDiffs := d.compareBlocks(baseResource, resource, baseEnv, env)
					results = append(results, blockDiffs...)
				}
			}
		}
	}

	return results, nil
}

// リソース存在差分を検出
func (d *HCLDiffer) compareResourceExistence(baseResources, envResources *EnvResources, baseEnv, env string) []*DiffResult {
	var results []*DiffResult

	// 基準環境のリソースをマップ化
	baseResourceMap := make(map[string]*EnvResource)
	for _, resource := range baseResources.Resources {
		key := fmt.Sprintf("%s.%s", resource.Type, resource.Name)
		baseResourceMap[key] = resource
	}

	// 比較環境のリソースをマップ化
	envResourceMap := make(map[string]*EnvResource)
	for _, resource := range envResources.Resources {
		key := fmt.Sprintf("%s.%s", resource.Type, resource.Name)
		envResourceMap[key] = resource
	}

	// 全リソースキーを収集
	allResourceKeys := make(map[string]bool)
	for key := range baseResourceMap {
		allResourceKeys[key] = true
	}
	for key := range envResourceMap {
		allResourceKeys[key] = true
	}

	// 各リソースの存在を比較
	for resourceKey := range allResourceKeys {
		baseExists := baseResourceMap[resourceKey] != nil
		envExists := envResourceMap[resourceKey] != nil

		if baseExists != envExists {
			// リソース存在差分を記録
			diff := &DiffResult{
				Resource:    resourceKey,
				Environment: env,
				Path:        "",  // リソース全体の存在差分なのでパスは空
				Expected:    cty.BoolVal(baseExists),
				Actual:      cty.BoolVal(envExists),
				IsIgnored:   d.isIgnored(resourceKey),
			}
			results = append(results, diff)
		}
	}

	return results
}

// 2つのリソースの属性を比較
func (d *HCLDiffer) compareAttributes(baseResource, resource *EnvResource, baseEnv, env string) []*DiffResult {
	var results []*DiffResult

	// 全ての属性名を収集
	allAttrNames := make(map[string]bool)
	for name := range baseResource.Attrs {
		allAttrNames[name] = true
	}
	for name := range resource.Attrs {
		allAttrNames[name] = true
	}

	// 各属性を比較
	for attrName := range allAttrNames {
		baseValue, baseExists := baseResource.Attrs[attrName]
		value, exists := resource.Attrs[attrName]

		// 値が存在しない場合の処理
		if !baseExists {
			baseValue = cty.NullVal(cty.DynamicPseudoType)
		}
		if !exists {
			value = cty.NullVal(cty.DynamicPseudoType)
		}

		// 値が異なる場合、差分として記録
		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("%s.%s.%s", baseResource.Type, baseResource.Name, attrName)

			// tags属性の場合は、ネストした属性も個別にチェック
			if attrName == "tags" && baseValue.Type().IsObjectType() && value.Type().IsObjectType() {
				nestedDiffs := d.compareTagAttributes(baseResource, resource, baseValue, value, baseEnv, env)
				results = append(results, nestedDiffs...)
			} else {
				diff := &DiffResult{
					Resource:    fmt.Sprintf("%s.%s", baseResource.Type, baseResource.Name),
					Environment: env,
					Path:        attrName,
					Expected:    baseValue,
					Actual:      value,
					IsIgnored:   d.isIgnored(resourcePath),
				}
				results = append(results, diff)
			}
		}
	}

	return results
}

// tagsのネストした属性を比較
func (d *HCLDiffer) compareTagAttributes(baseResource, resource *EnvResource, baseTags, tags cty.Value, baseEnv, env string) []*DiffResult {
	var results []*DiffResult

	if !baseTags.Type().IsObjectType() || !tags.Type().IsObjectType() {
		return results
	}

	// 全てのタグキーを収集
	allTagKeys := make(map[string]bool)
	for key := range baseTags.AsValueMap() {
		allTagKeys[key] = true
	}
	for key := range tags.AsValueMap() {
		allTagKeys[key] = true
	}

	baseTagMap := baseTags.AsValueMap()
	tagMap := tags.AsValueMap()

	// 各タグキーを比較
	for tagKey := range allTagKeys {
		baseValue, baseExists := baseTagMap[tagKey]
		value, exists := tagMap[tagKey]

		if !baseExists {
			baseValue = cty.NullVal(cty.String)
		}
		if !exists {
			value = cty.NullVal(cty.String)
		}

		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("%s.%s.tags.%s", baseResource.Type, baseResource.Name, tagKey)
			diff := &DiffResult{
				Resource:    fmt.Sprintf("%s.%s", baseResource.Type, baseResource.Name),
				Environment: env,
				Path:        fmt.Sprintf("tags.%s", tagKey),
				Expected:    baseValue,
				Actual:      value,
				IsIgnored:   d.isIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	return results
}

// ネストブロックを比較
func (d *HCLDiffer) compareBlocks(baseResource, resource *EnvResource, baseEnv, env string) []*DiffResult {
	var results []*DiffResult

	// 全ブロック型を収集
	allBlockTypes := make(map[string]bool)
	for blockType := range baseResource.Blocks {
		allBlockTypes[blockType] = true
	}
	for blockType := range resource.Blocks {
		allBlockTypes[blockType] = true
	}

	// 各ブロック型を比較
	for blockType := range allBlockTypes {
		baseBlocks := baseResource.Blocks[blockType]
		blocks := resource.Blocks[blockType]

		// ブロック数の差分をチェック
		maxLen := len(baseBlocks)
		if len(blocks) > maxLen {
			maxLen = len(blocks)
		}

		for i := 0; i < maxLen; i++ {
			var baseBlock, block *EnvBlock

			if i < len(baseBlocks) {
				baseBlock = baseBlocks[i]
			}
			if i < len(blocks) {
				block = blocks[i]
			}

			// ブロック存在差分をチェック
			if baseBlock == nil && block != nil {
				// 新しいブロックが追加された
				resourcePath := fmt.Sprintf("%s.%s.%s[%d]", baseResource.Type, baseResource.Name, blockType, i)
				diff := &DiffResult{
					Resource:    fmt.Sprintf("%s.%s", baseResource.Type, baseResource.Name),
					Environment: env,
					Path:        fmt.Sprintf("%s[%d]", blockType, i),
					Expected:    cty.NullVal(cty.DynamicPseudoType),
					Actual:      cty.StringVal("block_exists"),
					IsIgnored:   d.isIgnoredWithBlock(resourcePath, block, blockType),
				}
				results = append(results, diff)
			} else if baseBlock != nil && block == nil {
				// ブロックが削除された
				resourcePath := fmt.Sprintf("%s.%s.%s[%d]", baseResource.Type, baseResource.Name, blockType, i)
				diff := &DiffResult{
					Resource:    fmt.Sprintf("%s.%s", baseResource.Type, baseResource.Name),
					Environment: env,
					Path:        fmt.Sprintf("%s[%d]", blockType, i),
					Expected:    cty.StringVal("block_exists"),
					Actual:      cty.NullVal(cty.DynamicPseudoType),
					IsIgnored:   d.isIgnoredWithBlock(resourcePath, baseBlock, blockType),
				}
				results = append(results, diff)
			} else if baseBlock != nil && block != nil {
				// ブロック内属性を比較
				blockDiffs := d.compareBlockAttributes(baseResource, baseBlock, block, blockType, i, env)
				results = append(results, blockDiffs...)
			}
		}
	}

	return results
}

// ブロック内属性を比較
func (d *HCLDiffer) compareBlockAttributes(resource *EnvResource, baseBlock, block *EnvBlock, blockType string, index int, env string) []*DiffResult {
	var results []*DiffResult

	// 全属性名を収集
	allAttrNames := make(map[string]bool)
	for name := range baseBlock.Attrs {
		allAttrNames[name] = true
	}
	for name := range block.Attrs {
		allAttrNames[name] = true
	}

	// 各属性を比較
	for attrName := range allAttrNames {
		baseValue, baseExists := baseBlock.Attrs[attrName]
		value, exists := block.Attrs[attrName]

		if !baseExists {
			baseValue = cty.NullVal(cty.DynamicPseudoType)
		}
		if !exists {
			value = cty.NullVal(cty.DynamicPseudoType)
		}

		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("%s.%s.%s[%d].%s", resource.Type, resource.Name, blockType, index, attrName)
			diff := &DiffResult{
				Resource:    fmt.Sprintf("%s.%s", resource.Type, resource.Name),
				Environment: env,
				Path:        fmt.Sprintf("%s[%d].%s", blockType, index, attrName),
				Expected:    baseValue,
				Actual:      value,
				IsIgnored:   d.isIgnoredWithBlockAttribute(resourcePath, block, blockType),
			}
			results = append(results, diff)
		}
	}

	return results
}

// リソース・属性パスが.tfspecignoreルールにマッチするかチェック
func (d *HCLDiffer) isIgnored(resourcePath string) bool {
	for _, rule := range d.ignoreRules {
		if rule == resourcePath {
			return true
		}
		// 階層的マッチング: 親パスが無視される場合、子パスも無視
		if d.isChildPath(resourcePath, rule) {
			return true
		}
		// aws_security_group.web.ingress[443] 形式への対応
		if d.matchesPortRule(rule, resourcePath) {
			return true
		}
		// より柔軟なパターンマッチング
		if d.matchesPattern(rule, resourcePath) {
			return true
		}
	}
	return false
}

// 指定されたパスが親ルールの子パスかどうかをチェック
func (d *HCLDiffer) isChildPath(resourcePath, parentRule string) bool {
	// resourcePath が parentRule の子パスかどうか
	// 例: resourcePath="aws_security_group.web.ingress[443].from_port", parentRule="aws_security_group.web.ingress[443]"
	if strings.HasPrefix(resourcePath, parentRule+".") {
		return true
	}
	return false
}

// ポート番号指定ルールとのマッチング
func (d *HCLDiffer) matchesPortRule(rule, resourcePath string) bool {
	// パターン1: aws_security_group.web.ingress[443] 形式
	// これはポート443のingressブロック全体を無視する

	// パターン2: aws_security_group.web.ingress[22].cidr_blocks 形式
	// これはポート22のingressブロックのcidr_blocks属性を無視する

	// TODO: 将来の実装では実際のポート番号でマッチングする
	// 現在は属性ベースマッチングで対応
	return false
}

// ブロック情報を考慮した無視判定
func (d *HCLDiffer) isIgnoredWithBlock(resourcePath string, block *EnvBlock, blockType string) bool {
	// まず通常の無視判定を試行
	if d.isIgnored(resourcePath) {
		return true
	}

	// ポート番号ベースの無視判定（ingress/egressブロック用）
	if blockType == "ingress" || blockType == "egress" {
		if fromPort, exists := block.Attrs["from_port"]; exists && !fromPort.IsNull() {
			if fromPort.Type() == cty.Number {
				portNum := fromPort.AsBigFloat()
				if portNum.IsInt() {
					if port, accuracy := portNum.Int64(); accuracy == 0 {
						// ポート番号ベースのルールをチェック
						portBasedPath := strings.Replace(resourcePath, fmt.Sprintf("[%d]",
							extractIndexFromPath(resourcePath)), fmt.Sprintf("[%d]", port), 1)
						if d.isIgnored(portBasedPath) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// パスからインデックスを抽出
func extractIndexFromPath(resourcePath string) int {
	// aws_security_group.web.ingress[1] から 1 を抽出
	start := strings.LastIndex(resourcePath, "[")
	end := strings.LastIndex(resourcePath, "]")
	if start != -1 && end != -1 && end > start {
		indexStr := resourcePath[start+1 : end]
		if idx := parseIntSafe(indexStr); idx >= 0 {
			return idx
		}
	}
	return -1
}

// 安全な整数パース
func parseIntSafe(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return -1
}

// ブロック属性の無視判定（ポート番号ベース対応）
func (d *HCLDiffer) isIgnoredWithBlockAttribute(resourcePath string, block *EnvBlock, blockType string) bool {
	// まず通常の無視判定を試行
	if d.isIgnored(resourcePath) {
		return true
	}

	// ポート番号ベースの無視判定（ingress/egressブロック用）
	if blockType == "ingress" || blockType == "egress" {
		if fromPort, exists := block.Attrs["from_port"]; exists && !fromPort.IsNull() {
			if fromPort.Type() == cty.Number {
				portNum := fromPort.AsBigFloat()
				if portNum.IsInt() {
					if port, accuracy := portNum.Int64(); accuracy == 0 {
						// ポート番号ベースのルールをチェック
						// aws_security_group.web.ingress[1].cidr_blocks → aws_security_group.web.ingress[22].cidr_blocks
						portBasedPath := strings.Replace(resourcePath, fmt.Sprintf("[%d]",
							extractIndexFromPath(resourcePath)), fmt.Sprintf("[%d]", port), 1)
						if d.isIgnored(portBasedPath) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// より柔軟なパターンマッチング
func (d *HCLDiffer) matchesPattern(rule, resourcePath string) bool {
	// ワイルドカード的なマッチング
	// 例: "aws_security_group.web.ingress[*].cidr_blocks" のような将来の拡張
	return false
}

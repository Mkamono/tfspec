package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

type HCLDiffer struct {
	ignoreMatcher *IgnoreMatcher
}

func NewHCLDiffer(ignoreRules []string) *HCLDiffer {
	return &HCLDiffer{
		ignoreMatcher: NewIgnoreMatcher(ignoreRules),
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
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourceKey),
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
					IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
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
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
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
					Actual:      d.formatBlockContent(block),
					IsIgnored:   d.ignoreMatcher.IsIgnoredWithBlock(resourcePath, block, blockType),
				}
				results = append(results, diff)
			} else if baseBlock != nil && block == nil {
				// ブロックが削除された
				resourcePath := fmt.Sprintf("%s.%s.%s[%d]", baseResource.Type, baseResource.Name, blockType, i)
				diff := &DiffResult{
					Resource:    fmt.Sprintf("%s.%s", baseResource.Type, baseResource.Name),
					Environment: env,
					Path:        fmt.Sprintf("%s[%d]", blockType, i),
					Expected:    d.formatBlockContent(baseBlock),
					Actual:      cty.NullVal(cty.DynamicPseudoType),
					IsIgnored:   d.ignoreMatcher.IsIgnoredWithBlock(resourcePath, baseBlock, blockType),
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
				IsIgnored:   d.ignoreMatcher.IsIgnoredWithBlockAttribute(resourcePath, block, blockType),
			}
			results = append(results, diff)
		}
	}

	return results
}

// formatBlockContent はブロックの内容を文字列としてフォーマットする
func (d *HCLDiffer) formatBlockContent(block *EnvBlock) cty.Value {
	if block == nil {
		return cty.NullVal(cty.String)
	}

	// ブロックの属性を文字列形式で表現（ソート済み順序で）
	var attrNames []string
	for name := range block.Attrs {
		attrNames = append(attrNames, name)
	}
	sort.Strings(attrNames)

	var attrs []string
	for _, name := range attrNames {
		value := block.Attrs[name]
		if value.Type() == cty.String {
			attrs = append(attrs, fmt.Sprintf("%s: \"%s\"", name, value.AsString()))
		} else if value.Type() == cty.Number {
			attrs = append(attrs, fmt.Sprintf("%s: %s", name, value.AsBigFloat().String()))
		} else if value.Type() == cty.Bool {
			boolVal := "false"
			if value.True() {
				boolVal = "true"
			}
			attrs = append(attrs, fmt.Sprintf("%s: %s", name, boolVal))
		} else if value.Type().IsListType() || value.Type().IsTupleType() {
			// リストやタプルの場合
			var elements []string
			for it := value.ElementIterator(); it.Next(); {
				_, val := it.Element()
				if val.Type() == cty.String {
					elements = append(elements, fmt.Sprintf("\"%s\"", val.AsString()))
				} else {
					elements = append(elements, fmt.Sprintf("%v", val))
				}
			}
			attrs = append(attrs, fmt.Sprintf("%s: [%s]", name, fmt.Sprintf("%s", elements)))
		} else {
			attrs = append(attrs, fmt.Sprintf("%s: %v", name, value))
		}
	}

	// 属性をHTMLの<br>タグで改行して表示
	if len(attrs) == 0 {
		return cty.StringVal("{}")
	}
	if len(attrs) == 1 {
		return cty.StringVal(fmt.Sprintf("{ %s }", attrs[0]))
	}

	return cty.StringVal(fmt.Sprintf("{ %s }", strings.Join(attrs, ",<br>&nbsp;&nbsp;")))
}


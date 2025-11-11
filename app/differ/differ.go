package differ

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Mkamono/tfspec/app/types"
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

// ComparisonCallback は属性比較時のコールバック関数型
type ComparisonCallback func(attrName string, baseValue, value cty.Value, baseExists, exists bool) *types.DiffResult

// compareMapAttributes は属性マップを比較する汎用ヘルパー関数
// baseMap: ベース環境の属性マップ
// targetMap: 比較対象環境の属性マップ
// callback: 各属性について差分判定と結果生成を行うコールバック関数
func (d *HCLDiffer) compareMapAttributes(baseMap, targetMap map[string]cty.Value, callback ComparisonCallback) []*types.DiffResult {
	var results []*types.DiffResult

	// 全ての属性名を収集
	allAttrNames := make(map[string]bool)
	for name := range baseMap {
		allAttrNames[name] = true
	}
	for name := range targetMap {
		allAttrNames[name] = true
	}

	// 各属性を比較
	for attrName := range allAttrNames {
		baseValue, baseExists := baseMap[attrName]
		value, exists := targetMap[attrName]

		// 値が存在しない場合の処理
		if !baseExists {
			baseValue = cty.NullVal(cty.DynamicPseudoType)
		}
		if !exists {
			value = cty.NullVal(cty.DynamicPseudoType)
		}

		// コールバックで差分判定と結果生成
		if diff := callback(attrName, baseValue, value, baseExists, exists); diff != nil {
			results = append(results, diff)
		}
	}

	return results
}

// checkExistenceDiff は名前付きアイテムの存在差分をチェックする汎用ヘルパー関数
// baseMap, targetMap: 比較するマップ
// resourcePrefix: リソースパスのプレフィックス（例："module", "local", "var"）
// env: 環境名
func (d *HCLDiffer) checkExistenceDiff(baseMap, targetMap map[string]bool, resourcePrefix, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 全名を収集
	allNames := make(map[string]bool)
	for name := range baseMap {
		allNames[name] = true
	}
	for name := range targetMap {
		allNames[name] = true
	}

	// 各名の存在を比較
	for name := range allNames {
		baseExists := baseMap[name]
		targetExists := targetMap[name]

		if baseExists != targetExists {
			resourcePath := fmt.Sprintf("%s.%s", resourcePrefix, name)
			diff := &types.DiffResult{
				Resource:    resourcePath,
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(baseExists),
				Actual:      cty.BoolVal(targetExists),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	return results
}

// 環境間差分を検出し、.tfspecignoreルールでフィルタリング
func (d *HCLDiffer) Compare(envResources map[string]*types.EnvResources) ([]*types.DiffResult, error) {
	var results []*types.DiffResult

	// .tfspecignoreルールの検証を実行
	envResourcesMap := make(map[string]map[string]*types.EnvResource)
	for envName, envRes := range envResources {
		envResourcesMap[envName] = make(map[string]*types.EnvResource)
		for _, resource := range envRes.Resources {
			key := fmt.Sprintf("%s.%s", resource.Type, resource.Name)
			envResourcesMap[envName][key] = resource
		}
	}
	d.ignoreMatcher.ValidateRules(envResourcesMap)

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
		existenceDiffs := d.compareResourceExistence(baseEnvResources, envResourceList, env)
		results = append(results, existenceDiffs...)

		// 共通リソースの属性・ブロック差分を検出
		for _, baseResource := range baseEnvResources.Resources {
			for _, resource := range envResourceList.Resources {
				// リソース種別・名前が同じかチェック
				if baseResource.Type == resource.Type && baseResource.Name == resource.Name {
					// 属性を比較
					envDiffs := d.compareAttributes(baseResource, resource, env)
					results = append(results, envDiffs...)

					// ネストブロックを比較
					blockDiffs := d.compareBlocks(baseResource, resource, env)
					results = append(results, blockDiffs...)
				}
			}
		}

		// 新しいブロックタイプの比較
		// Modules
		moduleDiffs := d.compareModules(baseEnvResources.Modules, envResourceList.Modules, env)
		results = append(results, moduleDiffs...)

		// Locals
		localDiffs := d.compareLocals(baseEnvResources.Locals, envResourceList.Locals, env)
		results = append(results, localDiffs...)

		// Variables
		variableDiffs := d.compareVariables(baseEnvResources.Variables, envResourceList.Variables, env)
		results = append(results, variableDiffs...)

		// Outputs
		outputDiffs := d.compareOutputs(baseEnvResources.Outputs, envResourceList.Outputs, env)
		results = append(results, outputDiffs...)

		// Data Sources
		dataDiffs := d.compareDataSources(baseEnvResources.DataSources, envResourceList.DataSources, env)
		results = append(results, dataDiffs...)
	}

	return results, nil
}

// GetIgnoreWarnings は.tfspecignoreルール検証で発見された警告を返す
func (d *HCLDiffer) GetIgnoreWarnings() []string {
	return d.ignoreMatcher.GetWarnings()
}

// リソース存在差分を検出
func (d *HCLDiffer) compareResourceExistence(baseResources, envResources *types.EnvResources, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境のリソースをマップ化
	baseResourceMap := make(map[string]*types.EnvResource)
	for _, resource := range baseResources.Resources {
		key := fmt.Sprintf("%s.%s", resource.Type, resource.Name)
		baseResourceMap[key] = resource
	}

	// 比較環境のリソースをマップ化
	envResourceMap := make(map[string]*types.EnvResource)
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
			diff := &types.DiffResult{
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
func (d *HCLDiffer) compareAttributes(baseResource, resource *types.EnvResource, env string) []*types.DiffResult {
	callback := func(attrName string, baseValue, value cty.Value, baseExists, exists bool) *types.DiffResult {
		// 値が異なる場合、差分として記録
		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("%s.%s.%s", baseResource.Type, baseResource.Name, attrName)

			// tags属性の場合は、ネストした属性も個別にチェック
			if attrName == "tags" && baseValue.Type().IsObjectType() && value.Type().IsObjectType() {
				// このコールバック内では処理しない（親関数で処理）
				return nil
			}

			return &types.DiffResult{
				Resource:    fmt.Sprintf("%s.%s", baseResource.Type, baseResource.Name),
				Environment: env,
				Path:        attrName,
				Expected:    baseValue,
				Actual:      value,
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
		}
		return nil
	}

	results := d.compareMapAttributes(baseResource.Attrs, resource.Attrs, callback)

	// tags属性の場合は、ネストした属性も個別にチェック
	baseTags, baseHasTags := baseResource.Attrs["tags"]
	tags, hasTags := resource.Attrs["tags"]
	if baseHasTags && hasTags && baseTags.Type().IsObjectType() && tags.Type().IsObjectType() {
		nestedDiffs := d.compareTagAttributes(baseResource, baseTags, tags, env)
		results = append(results, nestedDiffs...)
	}

	return results
}

// tagsのネストした属性を比較
func (d *HCLDiffer) compareTagAttributes(baseResource *types.EnvResource, baseTags, tags cty.Value, env string) []*types.DiffResult {
	if !baseTags.Type().IsObjectType() || !tags.Type().IsObjectType() {
		return []*types.DiffResult{}
	}

	baseTagMap := baseTags.AsValueMap()
	tagMap := tags.AsValueMap()

	callback := func(tagKey string, baseValue, value cty.Value, baseExists, exists bool) *types.DiffResult {
		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("%s.%s.tags.%s", baseResource.Type, baseResource.Name, tagKey)
			return &types.DiffResult{
				Resource:    fmt.Sprintf("%s.%s", baseResource.Type, baseResource.Name),
				Environment: env,
				Path:        fmt.Sprintf("tags.%s", tagKey),
				Expected:    baseValue,
				Actual:      value,
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
		}
		return nil
	}

	return d.compareMapAttributes(baseTagMap, tagMap, callback)
}

// ネストブロックを比較
func (d *HCLDiffer) compareBlocks(baseResource, resource *types.EnvResource, env string) []*types.DiffResult {
	return d.compareBlocksWithPrefix(baseResource, resource, env, "")
}

// compareBlocksWithPrefix はリソースプレフィックス付きでネストブロックを比較
func (d *HCLDiffer) compareBlocksWithPrefix(baseResource, resource *types.EnvResource, env, resourcePrefix string) []*types.DiffResult {
	var results []*types.DiffResult

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
		maxLen := max(len(baseBlocks), len(blocks))

		for i := range maxLen {
			var baseBlock, block *types.EnvBlock

			if i < len(baseBlocks) {
				baseBlock = baseBlocks[i]
			}
			if i < len(blocks) {
				block = blocks[i]
			}

			// リソースパス構築
			var resourcePath, resourceDisplay, pathDisplay string
			if resourcePrefix == "" {
				resourcePath = fmt.Sprintf("%s.%s.%s[%d]", baseResource.Type, baseResource.Name, blockType, i)
				resourceDisplay = fmt.Sprintf("%s.%s", baseResource.Type, baseResource.Name)
				pathDisplay = fmt.Sprintf("%s[%d]", blockType, i)
			} else {
				resourcePath = fmt.Sprintf("%s.%s.%s.%s[%d]", resourcePrefix, baseResource.Type, baseResource.Name, blockType, i)
				resourceDisplay = fmt.Sprintf("%s.%s", resourcePrefix, baseResource.Type)
				pathDisplay = fmt.Sprintf("%s.%s[%d]", baseResource.Name, blockType, i)
			}

			// ブロック存在差分をチェック
			if baseBlock == nil && block != nil {
				// 新しいブロックが追加された
				diff := &types.DiffResult{
					Resource:    resourceDisplay,
					Environment: env,
					Path:        pathDisplay,
					Expected:    cty.NullVal(cty.DynamicPseudoType),
					Actual:      d.formatBlockContent(block),
					IsIgnored:   d.ignoreMatcher.IsIgnoredWithBlock(resourcePath),
				}
				results = append(results, diff)
			} else if baseBlock != nil && block == nil {
				// ブロックが削除された
				diff := &types.DiffResult{
					Resource:    resourceDisplay,
					Environment: env,
					Path:        pathDisplay,
					Expected:    d.formatBlockContent(baseBlock),
					Actual:      cty.NullVal(cty.DynamicPseudoType),
					IsIgnored:   d.ignoreMatcher.IsIgnoredWithBlock(resourcePath),
				}
				results = append(results, diff)
			} else if baseBlock != nil && block != nil {
				// ブロック内属性を比較
				blockDiffs := d.compareBlockAttributesWithPrefix(baseResource, baseBlock, block, blockType, i, env, resourcePrefix)
				results = append(results, blockDiffs...)
			}
		}
	}

	return results
}

// ブロック内属性を比較
func (d *HCLDiffer) compareBlockAttributes(resource *types.EnvResource, baseBlock, block *types.EnvBlock, blockType string, index int, env string) []*types.DiffResult {
	return d.compareBlockAttributesWithPrefix(resource, baseBlock, block, blockType, index, env, "")
}

// compareBlockAttributesWithPrefix はリソースプレフィックス付きでブロック内属性を比較
func (d *HCLDiffer) compareBlockAttributesWithPrefix(resource *types.EnvResource, baseBlock, block *types.EnvBlock, blockType string, index int, env, resourcePrefix string) []*types.DiffResult {
	var results []*types.DiffResult

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
			var resourcePath, resourceDisplay, pathDisplay string
			if resourcePrefix == "" {
				resourcePath = fmt.Sprintf("%s.%s.%s[%d].%s", resource.Type, resource.Name, blockType, index, attrName)
				resourceDisplay = fmt.Sprintf("%s.%s", resource.Type, resource.Name)
				pathDisplay = fmt.Sprintf("%s[%d].%s", blockType, index, attrName)
			} else {
				resourcePath = fmt.Sprintf("%s.%s.%s.%s[%d].%s", resourcePrefix, resource.Type, resource.Name, blockType, index, attrName)
				resourceDisplay = fmt.Sprintf("%s.%s", resourcePrefix, resource.Type)
				pathDisplay = fmt.Sprintf("%s.%s[%d].%s", resource.Name, blockType, index, attrName)
			}

			diff := &types.DiffResult{
				Resource:    resourceDisplay,
				Environment: env,
				Path:        pathDisplay,
				Expected:    baseValue,
				Actual:      value,
				IsIgnored:   d.ignoreMatcher.IsIgnoredWithBlockAttribute(resourcePath),
			}
			results = append(results, diff)
		}
	}

	return results
}

// formatBlockContent はブロックの内容を文字列としてフォーマットする
func (d *HCLDiffer) formatBlockContent(block *types.EnvBlock) cty.Value {
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

	return cty.StringVal(fmt.Sprintf("{<br>&nbsp;&nbsp;%s<br>}", strings.Join(attrs, ",<br>&nbsp;&nbsp;")))
}

// compareNamedAttributes は名前付きアイテム間の属性差分を比較する汎用ヘルパー関数
// baseAttrs, targetAttrs: 比較する属性マップ
// resourcePrefix, resourceName: リソースパス構築用
// env: 環境名
func (d *HCLDiffer) compareNamedAttributes(baseAttrs, targetAttrs map[string]cty.Value, resourcePrefix, resourceName, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 全属性名を収集
	allAttrNames := make(map[string]bool)
	for name := range baseAttrs {
		allAttrNames[name] = true
	}
	for name := range targetAttrs {
		allAttrNames[name] = true
	}

	// 各属性を比較
	for attrName := range allAttrNames {
		baseValue, baseExists := baseAttrs[attrName]
		value, targetExists := targetAttrs[attrName]

		if !baseExists {
			baseValue = cty.NullVal(cty.DynamicPseudoType)
		}
		if !targetExists {
			value = cty.NullVal(cty.DynamicPseudoType)
		}

		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("%s.%s.%s", resourcePrefix, resourceName, attrName)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("%s.%s", resourcePrefix, resourceName),
				Environment: env,
				Path:        attrName,
				Expected:    baseValue,
				Actual:      value,
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	return results
}

// compareModules はモジュール間の差分を比較
func (d *HCLDiffer) compareModules(baseModules, envModules []*types.EnvModule, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境のモジュールをマップ化
	baseModuleMap := make(map[string]*types.EnvModule)
	baseExistenceMap := make(map[string]bool)
	for _, module := range baseModules {
		baseModuleMap[module.Name] = module
		baseExistenceMap[module.Name] = true
	}

	// 比較環境のモジュールをマップ化
	envModuleMap := make(map[string]*types.EnvModule)
	envExistenceMap := make(map[string]bool)
	for _, module := range envModules {
		envModuleMap[module.Name] = module
		envExistenceMap[module.Name] = true
	}

	// 存在差分をチェック
	existenceDiffs := d.checkExistenceDiff(baseExistenceMap, envExistenceMap, "module", env)
	results = append(results, existenceDiffs...)

	// 属性差分をチェック
	for name, baseModule := range baseModuleMap {
		if envModule, exists := envModuleMap[name]; exists {
			attrDiffs := d.compareModuleAttributes(baseModule, envModule, env)
			results = append(results, attrDiffs...)
		}
	}

	return results
}

// compareModuleAttributes はモジュール属性間の差分を比較
func (d *HCLDiffer) compareModuleAttributes(baseModule, envModule *types.EnvModule, env string) []*types.DiffResult {
	return d.compareNamedAttributes(baseModule.Attrs, envModule.Attrs, "module", baseModule.Name, env)
}

// compareLocals はローカル変数間の差分を比較
func (d *HCLDiffer) compareLocals(baseLocals, envLocals []*types.EnvLocal, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境のローカル変数をマップ化
	baseLocalMap := make(map[string]*types.EnvLocal)
	baseExistenceMap := make(map[string]bool)
	for _, local := range baseLocals {
		baseLocalMap[local.Name] = local
		baseExistenceMap[local.Name] = true
	}

	// 比較環境のローカル変数をマップ化
	envLocalMap := make(map[string]*types.EnvLocal)
	envExistenceMap := make(map[string]bool)
	for _, local := range envLocals {
		envLocalMap[local.Name] = local
		envExistenceMap[local.Name] = true
	}

	// 存在差分をチェック
	existenceDiffs := d.checkExistenceDiff(baseExistenceMap, envExistenceMap, "local", env)
	results = append(results, existenceDiffs...)

	// 値差分をチェック
	for name, baseLocal := range baseLocalMap {
		if envLocal, exists := envLocalMap[name]; exists {
			if !baseLocal.Value.Equals(envLocal.Value).True() {
				resourcePath := fmt.Sprintf("local.%s", name)
				diff := &types.DiffResult{
					Resource:    fmt.Sprintf("local.%s", name),
					Environment: env,
					Path:        "",
					Expected:    baseLocal.Value,
					Actual:      envLocal.Value,
					IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
				}
				results = append(results, diff)
			}
		}
	}

	return results
}

// compareVariables は変数間の差分を比較
func (d *HCLDiffer) compareVariables(baseVariables, envVariables []*types.EnvVariable, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境の変数をマップ化
	baseVariableMap := make(map[string]*types.EnvVariable)
	baseExistenceMap := make(map[string]bool)
	for _, variable := range baseVariables {
		baseVariableMap[variable.Name] = variable
		baseExistenceMap[variable.Name] = true
	}

	// 比較環境の変数をマップ化
	envVariableMap := make(map[string]*types.EnvVariable)
	envExistenceMap := make(map[string]bool)
	for _, variable := range envVariables {
		envVariableMap[variable.Name] = variable
		envExistenceMap[variable.Name] = true
	}

	// 存在差分をチェック
	existenceDiffs := d.checkExistenceDiff(baseExistenceMap, envExistenceMap, "var", env)
	results = append(results, existenceDiffs...)

	// 属性差分をチェック
	for name, baseVariable := range baseVariableMap {
		if envVariable, exists := envVariableMap[name]; exists {
			attrDiffs := d.compareVariableAttributes(baseVariable, envVariable, env)
			results = append(results, attrDiffs...)
		}
	}

	return results
}

// compareVariableAttributes は変数属性間の差分を比較
func (d *HCLDiffer) compareVariableAttributes(baseVariable, envVariable *types.EnvVariable, env string) []*types.DiffResult {
	return d.compareNamedAttributes(baseVariable.Attrs, envVariable.Attrs, "var", baseVariable.Name, env)
}

// compareOutputs は出力変数間の差分を比較
func (d *HCLDiffer) compareOutputs(baseOutputs, envOutputs []*types.EnvOutput, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境の出力変数をマップ化
	baseOutputMap := make(map[string]*types.EnvOutput)
	baseExistenceMap := make(map[string]bool)
	for _, output := range baseOutputs {
		baseOutputMap[output.Name] = output
		baseExistenceMap[output.Name] = true
	}

	// 比較環境の出力変数をマップ化
	envOutputMap := make(map[string]*types.EnvOutput)
	envExistenceMap := make(map[string]bool)
	for _, output := range envOutputs {
		envOutputMap[output.Name] = output
		envExistenceMap[output.Name] = true
	}

	// 存在差分をチェック
	existenceDiffs := d.checkExistenceDiff(baseExistenceMap, envExistenceMap, "output", env)
	results = append(results, existenceDiffs...)

	// 属性差分をチェック
	for name, baseOutput := range baseOutputMap {
		if envOutput, exists := envOutputMap[name]; exists {
			attrDiffs := d.compareOutputAttributes(baseOutput, envOutput, env)
			results = append(results, attrDiffs...)
		}
	}

	return results
}

// compareOutputAttributes は出力変数属性間の差分を比較
func (d *HCLDiffer) compareOutputAttributes(baseOutput, envOutput *types.EnvOutput, env string) []*types.DiffResult {
	return d.compareNamedAttributes(baseOutput.Attrs, envOutput.Attrs, "output", baseOutput.Name, env)
}

// compareDataSources はデータソース間の差分を比較
func (d *HCLDiffer) compareDataSources(baseDataSources, envDataSources []*types.EnvData, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境のデータソースをマップ化
	baseDataMap := make(map[string]*types.EnvData)
	baseExistenceMap := make(map[string]bool)
	for _, data := range baseDataSources {
		key := fmt.Sprintf("%s.%s", data.Type, data.Name)
		baseDataMap[key] = data
		baseExistenceMap[key] = true
	}

	// 比較環境のデータソースをマップ化
	envDataMap := make(map[string]*types.EnvData)
	envExistenceMap := make(map[string]bool)
	for _, data := range envDataSources {
		key := fmt.Sprintf("%s.%s", data.Type, data.Name)
		envDataMap[key] = data
		envExistenceMap[key] = true
	}

	// 存在差分をチェック（"data"プレフィックスを付けるため、存在差分チェック前にマップ変換）
	allKeys := make(map[string]bool)
	for key := range baseExistenceMap {
		allKeys[key] = true
	}
	for key := range envExistenceMap {
		allKeys[key] = true
	}

	for key := range allKeys {
		baseExists := baseExistenceMap[key]
		envExists := envExistenceMap[key]

		if baseExists != envExists {
			resourcePath := fmt.Sprintf("data.%s", key)
			diff := &types.DiffResult{
				Resource:    resourcePath,
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(baseExists),
				Actual:      cty.BoolVal(envExists),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	// 属性差分とブロック差分をチェック
	for key, baseData := range baseDataMap {
		if envData, exists := envDataMap[key]; exists {
			// 属性差分を比較
			attrDiffs := d.compareDataSourceAttributes(baseData, envData, env)
			results = append(results, attrDiffs...)

			// ブロック差分を比較（EnvDataをEnvResourceに変換）
			baseResource := &types.EnvResource{
				Type:   baseData.Type,
				Name:   baseData.Name,
				Attrs:  baseData.Attrs,
				Blocks: baseData.Blocks,
			}
			envResource := &types.EnvResource{
				Type:   envData.Type,
				Name:   envData.Name,
				Attrs:  envData.Attrs,
				Blocks: envData.Blocks,
			}
			blockDiffs := d.compareBlocksWithPrefix(baseResource, envResource, env, "data")
			results = append(results, blockDiffs...)
		}
	}

	return results
}

// compareDataSourceAttributes はデータソース属性間の差分を比較
func (d *HCLDiffer) compareDataSourceAttributes(baseData, envData *types.EnvData, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 全属性名を収集
	allAttrNames := make(map[string]bool)
	for name := range baseData.Attrs {
		allAttrNames[name] = true
	}
	for name := range envData.Attrs {
		allAttrNames[name] = true
	}

	// 各属性を比較
	for attrName := range allAttrNames {
		baseValue, baseExists := baseData.Attrs[attrName]
		value, exists := envData.Attrs[attrName]

		if !baseExists {
			baseValue = cty.NullVal(cty.DynamicPseudoType)
		}
		if !exists {
			value = cty.NullVal(cty.DynamicPseudoType)
		}

		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("data.%s.%s.%s", baseData.Type, baseData.Name, attrName)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("data.%s", baseData.Type),
				Environment: env,
				Path:        fmt.Sprintf("%s.%s", baseData.Name, attrName),
				Expected:    baseValue,
				Actual:      value,
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	return results
}



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

		// 新しいブロックタイプの比較
		// Modules
		moduleDiffs := d.compareModules(baseEnvResources.Modules, envResourceList.Modules, baseEnv, env)
		results = append(results, moduleDiffs...)

		// Locals
		localDiffs := d.compareLocals(baseEnvResources.Locals, envResourceList.Locals, baseEnv, env)
		results = append(results, localDiffs...)

		// Variables
		variableDiffs := d.compareVariables(baseEnvResources.Variables, envResourceList.Variables, baseEnv, env)
		results = append(results, variableDiffs...)

		// Outputs
		outputDiffs := d.compareOutputs(baseEnvResources.Outputs, envResourceList.Outputs, baseEnv, env)
		results = append(results, outputDiffs...)

		// Data Sources
		dataDiffs := d.compareDataSources(baseEnvResources.DataSources, envResourceList.DataSources, baseEnv, env)
		results = append(results, dataDiffs...)
	}

	return results, nil
}

// GetIgnoreWarnings は.tfspecignoreルール検証で発見された警告を返す
func (d *HCLDiffer) GetIgnoreWarnings() []string {
	return d.ignoreMatcher.GetWarnings()
}

// リソース存在差分を検出
func (d *HCLDiffer) compareResourceExistence(baseResources, envResources *types.EnvResources, baseEnv, env string) []*types.DiffResult {
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
func (d *HCLDiffer) compareAttributes(baseResource, resource *types.EnvResource, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

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
				diff := &types.DiffResult{
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
func (d *HCLDiffer) compareTagAttributes(baseResource, resource *types.EnvResource, baseTags, tags cty.Value, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

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
			diff := &types.DiffResult{
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
func (d *HCLDiffer) compareBlocks(baseResource, resource *types.EnvResource, baseEnv, env string) []*types.DiffResult {
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
		maxLen := len(baseBlocks)
		if len(blocks) > maxLen {
			maxLen = len(blocks)
		}

		for i := 0; i < maxLen; i++ {
			var baseBlock, block *types.EnvBlock

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
				diff := &types.DiffResult{
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
				diff := &types.DiffResult{
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
func (d *HCLDiffer) compareBlockAttributes(resource *types.EnvResource, baseBlock, block *types.EnvBlock, blockType string, index int, env string) []*types.DiffResult {
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
			resourcePath := fmt.Sprintf("%s.%s.%s[%d].%s", resource.Type, resource.Name, blockType, index, attrName)
			diff := &types.DiffResult{
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

// compareModules はモジュール間の差分を比較
func (d *HCLDiffer) compareModules(baseModules, envModules []*types.EnvModule, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境のモジュールをマップ化
	baseModuleMap := make(map[string]*types.EnvModule)
	for _, module := range baseModules {
		baseModuleMap[module.Name] = module
	}

	// 比較環境のモジュールをマップ化
	envModuleMap := make(map[string]*types.EnvModule)
	for _, module := range envModules {
		envModuleMap[module.Name] = module
	}

	// 存在差分をチェック
	for name := range baseModuleMap {
		if _, exists := envModuleMap[name]; !exists {
			resourcePath := fmt.Sprintf("module.%s", name)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("module.%s", name),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(true),
				Actual:      cty.BoolVal(false),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	for name := range envModuleMap {
		if _, exists := baseModuleMap[name]; !exists {
			resourcePath := fmt.Sprintf("module.%s", name)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("module.%s", name),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(false),
				Actual:      cty.BoolVal(true),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	// 属性差分をチェック
	for name, baseModule := range baseModuleMap {
		if envModule, exists := envModuleMap[name]; exists {
			attrDiffs := d.compareModuleAttributes(baseModule, envModule, baseEnv, env)
			results = append(results, attrDiffs...)
		}
	}

	return results
}

// compareModuleAttributes はモジュール属性間の差分を比較
func (d *HCLDiffer) compareModuleAttributes(baseModule, envModule *types.EnvModule, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 全属性名を収集
	allAttrNames := make(map[string]bool)
	for name := range baseModule.Attrs {
		allAttrNames[name] = true
	}
	for name := range envModule.Attrs {
		allAttrNames[name] = true
	}

	// 各属性を比較
	for attrName := range allAttrNames {
		baseValue, baseExists := baseModule.Attrs[attrName]
		value, exists := envModule.Attrs[attrName]

		if !baseExists {
			baseValue = cty.NullVal(cty.DynamicPseudoType)
		}
		if !exists {
			value = cty.NullVal(cty.DynamicPseudoType)
		}

		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("module.%s.%s", baseModule.Name, attrName)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("module.%s", baseModule.Name),
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

// compareLocals はローカル変数間の差分を比較
func (d *HCLDiffer) compareLocals(baseLocals, envLocals []*types.EnvLocal, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境のローカル変数をマップ化
	baseLocalMap := make(map[string]*types.EnvLocal)
	for _, local := range baseLocals {
		baseLocalMap[local.Name] = local
	}

	// 比較環境のローカル変数をマップ化
	envLocalMap := make(map[string]*types.EnvLocal)
	for _, local := range envLocals {
		envLocalMap[local.Name] = local
	}

	// 存在差分をチェック
	for name := range baseLocalMap {
		if _, exists := envLocalMap[name]; !exists {
			resourcePath := fmt.Sprintf("local.%s", name)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("local.%s", name),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(true),
				Actual:      cty.BoolVal(false),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	for name := range envLocalMap {
		if _, exists := baseLocalMap[name]; !exists {
			resourcePath := fmt.Sprintf("local.%s", name)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("local.%s", name),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(false),
				Actual:      cty.BoolVal(true),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

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
func (d *HCLDiffer) compareVariables(baseVariables, envVariables []*types.EnvVariable, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境の変数をマップ化
	baseVariableMap := make(map[string]*types.EnvVariable)
	for _, variable := range baseVariables {
		baseVariableMap[variable.Name] = variable
	}

	// 比較環境の変数をマップ化
	envVariableMap := make(map[string]*types.EnvVariable)
	for _, variable := range envVariables {
		envVariableMap[variable.Name] = variable
	}

	// 存在差分をチェック
	for name := range baseVariableMap {
		if _, exists := envVariableMap[name]; !exists {
			resourcePath := fmt.Sprintf("var.%s", name)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("var.%s", name),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(true),
				Actual:      cty.BoolVal(false),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	for name := range envVariableMap {
		if _, exists := baseVariableMap[name]; !exists {
			resourcePath := fmt.Sprintf("var.%s", name)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("var.%s", name),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(false),
				Actual:      cty.BoolVal(true),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	// 属性差分をチェック
	for name, baseVariable := range baseVariableMap {
		if envVariable, exists := envVariableMap[name]; exists {
			attrDiffs := d.compareVariableAttributes(baseVariable, envVariable, baseEnv, env)
			results = append(results, attrDiffs...)
		}
	}

	return results
}

// compareVariableAttributes は変数属性間の差分を比較
func (d *HCLDiffer) compareVariableAttributes(baseVariable, envVariable *types.EnvVariable, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 全属性名を収集
	allAttrNames := make(map[string]bool)
	for name := range baseVariable.Attrs {
		allAttrNames[name] = true
	}
	for name := range envVariable.Attrs {
		allAttrNames[name] = true
	}

	// 各属性を比較
	for attrName := range allAttrNames {
		baseValue, baseExists := baseVariable.Attrs[attrName]
		value, exists := envVariable.Attrs[attrName]

		if !baseExists {
			baseValue = cty.NullVal(cty.DynamicPseudoType)
		}
		if !exists {
			value = cty.NullVal(cty.DynamicPseudoType)
		}

		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("var.%s.%s", baseVariable.Name, attrName)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("var.%s", baseVariable.Name),
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

// compareOutputs は出力変数間の差分を比較
func (d *HCLDiffer) compareOutputs(baseOutputs, envOutputs []*types.EnvOutput, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境の出力変数をマップ化
	baseOutputMap := make(map[string]*types.EnvOutput)
	for _, output := range baseOutputs {
		baseOutputMap[output.Name] = output
	}

	// 比較環境の出力変数をマップ化
	envOutputMap := make(map[string]*types.EnvOutput)
	for _, output := range envOutputs {
		envOutputMap[output.Name] = output
	}

	// 存在差分をチェック
	for name := range baseOutputMap {
		if _, exists := envOutputMap[name]; !exists {
			resourcePath := fmt.Sprintf("output.%s", name)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("output.%s", name),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(true),
				Actual:      cty.BoolVal(false),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	for name := range envOutputMap {
		if _, exists := baseOutputMap[name]; !exists {
			resourcePath := fmt.Sprintf("output.%s", name)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("output.%s", name),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(false),
				Actual:      cty.BoolVal(true),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	// 属性差分をチェック
	for name, baseOutput := range baseOutputMap {
		if envOutput, exists := envOutputMap[name]; exists {
			attrDiffs := d.compareOutputAttributes(baseOutput, envOutput, baseEnv, env)
			results = append(results, attrDiffs...)
		}
	}

	return results
}

// compareOutputAttributes は出力変数属性間の差分を比較
func (d *HCLDiffer) compareOutputAttributes(baseOutput, envOutput *types.EnvOutput, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 全属性名を収集
	allAttrNames := make(map[string]bool)
	for name := range baseOutput.Attrs {
		allAttrNames[name] = true
	}
	for name := range envOutput.Attrs {
		allAttrNames[name] = true
	}

	// 各属性を比較
	for attrName := range allAttrNames {
		baseValue, baseExists := baseOutput.Attrs[attrName]
		value, exists := envOutput.Attrs[attrName]

		if !baseExists {
			baseValue = cty.NullVal(cty.DynamicPseudoType)
		}
		if !exists {
			value = cty.NullVal(cty.DynamicPseudoType)
		}

		if !baseValue.Equals(value).True() {
			resourcePath := fmt.Sprintf("output.%s.%s", baseOutput.Name, attrName)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("output.%s", baseOutput.Name),
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

// compareDataSources はデータソース間の差分を比較
func (d *HCLDiffer) compareDataSources(baseDataSources, envDataSources []*types.EnvData, baseEnv, env string) []*types.DiffResult {
	var results []*types.DiffResult

	// 基準環境のデータソースをマップ化
	baseDataMap := make(map[string]*types.EnvData)
	for _, data := range baseDataSources {
		key := fmt.Sprintf("%s.%s", data.Type, data.Name)
		baseDataMap[key] = data
	}

	// 比較環境のデータソースをマップ化
	envDataMap := make(map[string]*types.EnvData)
	for _, data := range envDataSources {
		key := fmt.Sprintf("%s.%s", data.Type, data.Name)
		envDataMap[key] = data
	}

	// 存在差分をチェック
	for key := range baseDataMap {
		if _, exists := envDataMap[key]; !exists {
			resourcePath := fmt.Sprintf("data.%s", key)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("data.%s", key),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(true),
				Actual:      cty.BoolVal(false),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	for key := range envDataMap {
		if _, exists := baseDataMap[key]; !exists {
			resourcePath := fmt.Sprintf("data.%s", key)
			diff := &types.DiffResult{
				Resource:    fmt.Sprintf("data.%s", key),
				Environment: env,
				Path:        "",
				Expected:    cty.BoolVal(false),
				Actual:      cty.BoolVal(true),
				IsIgnored:   d.ignoreMatcher.IsIgnored(resourcePath),
			}
			results = append(results, diff)
		}
	}

	// 属性差分をチェック
	for key, baseData := range baseDataMap {
		if envData, exists := envDataMap[key]; exists {
			attrDiffs := d.compareDataSourceAttributes(baseData, envData, baseEnv, env)
			results = append(results, attrDiffs...)
		}
	}

	return results
}

// compareDataSourceAttributes はデータソース属性間の差分を比較
func (d *HCLDiffer) compareDataSourceAttributes(baseData, envData *types.EnvData, baseEnv, env string) []*types.DiffResult {
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
				Resource:    fmt.Sprintf("data.%s.%s", baseData.Type, baseData.Name),
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


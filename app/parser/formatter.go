package parser

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// ValueFormatter は値のフォーマット処理を担当する
type ValueFormatter struct{
	useMarkdownLineBreaks bool
	maxLength             int
}

func NewValueFormatter() *ValueFormatter {
	return &ValueFormatter{}
}

// FormatValueWithMarkdown は値を文字列形式でフォーマット（マークダウン対応）
func (f *ValueFormatter) FormatValueWithMarkdown(val interface{}, maxLength ...int) string {
	f.useMarkdownLineBreaks = true
	if len(maxLength) > 0 {
		f.maxLength = maxLength[0]
	}
	defer func() {
		f.useMarkdownLineBreaks = false
		f.maxLength = 0
	}()
	result := f.FormatValue(val)

	// 最大文字数を超えた場合は省略
	if f.maxLength > 0 && len(result) > f.maxLength {
		result = result[:f.maxLength] + "..."
	}

	return result
}

// FormatValue は値を文字列形式でフォーマットする
func (f *ValueFormatter) FormatValue(val interface{}) string {
	if val == nil {
		return ""
	}

	if ctyVal, ok := val.(cty.Value); ok {
		return f.formatCtyValue(ctyVal)
	}

	return fmt.Sprintf("%v", val)
}

func (f *ValueFormatter) formatCtyValue(ctyVal cty.Value) string {
	if ctyVal.IsNull() {
		return ""
	}

	switch {
	case ctyVal.Type() == cty.String:
		str := ctyVal.AsString()
		// Markdown出力の場合、改行を<br>に変換
		if f.useMarkdownLineBreaks && strings.Contains(str, "\n") {
			str = strings.ReplaceAll(str, "\n", "<br>")
		}
		return str
	case ctyVal.Type() == cty.Number:
		if bigFloat := ctyVal.AsBigFloat(); bigFloat.IsInt() {
			if val, accuracy := bigFloat.Int64(); accuracy == 0 {
				return fmt.Sprintf("%d", val)
			}
		}
		return ctyVal.AsBigFloat().String()
	case ctyVal.Type() == cty.Bool:
		if ctyVal.True() {
			return "true"
		}
		return "false"
	case ctyVal.Type().IsListType() || ctyVal.Type().IsTupleType() || ctyVal.Type().IsSetType():
		return f.formatListValue(ctyVal)
	case ctyVal.Type().IsObjectType() || ctyVal.Type().IsMapType():
		return f.formatMapValue(ctyVal)
	default:
		result := fmt.Sprintf("%s", ctyVal)
		// Markdown出力の場合、改行を<br>に変換
		if f.useMarkdownLineBreaks {
			result = strings.ReplaceAll(result, "\n", "<br>")
		}
		return result
	}
}

func (f *ValueFormatter) formatListValue(ctyVal cty.Value) string {
	var elements []string
	for it := ctyVal.ElementIterator(); it.Next(); {
		_, val := it.Element()
		elements = append(elements, f.FormatValue(val))
	}

	if f.useMarkdownLineBreaks && len(elements) > 2 {
		// 複数要素の場合は改行で区切る
		return fmt.Sprintf("[%s]", strings.Join(elements, "<br>"))
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (f *ValueFormatter) formatMapValue(ctyVal cty.Value) string {
	var pairs []string
	for it := ctyVal.ElementIterator(); it.Next(); {
		key, val := it.Element()
		pairs = append(pairs, fmt.Sprintf("%s: %s", f.FormatValue(key), f.FormatValue(val)))
	}

	if f.useMarkdownLineBreaks && len(pairs) > 2 {
		// 複数ペアの場合は改行で区切る
		return fmt.Sprintf("{%s}", strings.Join(pairs, "<br>"))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}
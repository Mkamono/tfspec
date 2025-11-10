package app

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// ValueFormatter は値のフォーマット処理を担当する
type ValueFormatter struct{}

func NewValueFormatter() *ValueFormatter {
	return &ValueFormatter{}
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
		return ctyVal.AsString()
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
		return fmt.Sprintf("%s", ctyVal)
	}
}

func (f *ValueFormatter) formatListValue(ctyVal cty.Value) string {
	var elements []string
	for it := ctyVal.ElementIterator(); it.Next(); {
		_, val := it.Element()
		elements = append(elements, f.FormatValue(val))
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (f *ValueFormatter) formatMapValue(ctyVal cty.Value) string {
	var pairs []string
	for it := ctyVal.ElementIterator(); it.Next(); {
		key, val := it.Element()
		pairs = append(pairs, fmt.Sprintf("%s: %s", f.FormatValue(key), f.FormatValue(val)))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}
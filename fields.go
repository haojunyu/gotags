package main

import (
	"fmt"
	"regexp"
)

// FieldSet结构体： 一个tag字典，包含各个fields
type FieldSet map[TagField]bool

// Include接口： 判断FieldSet中是否包含field
func (f FieldSet) Includes(field TagField) bool {
	b, ok := f[field]
	return ok && b
}

// ErrInvalidFields结构体： 解析异常要返回的错误信息
type ErrInvalidFields struct {
	Fields string
}

func (e ErrInvalidFields) Error() string {
	return fmt.Sprintf("invalid fields: %s", e.Fields)
}

// currently only "+l" is supported, 不是数字1，是字母l
var fieldsPattern = regexp.MustCompile(`^\+l$`)

func parseFields(fields string) (FieldSet, error) {
	if fields == "" {
		return FieldSet{}, nil
	}
	if fieldsPattern.MatchString(fields) {
		return FieldSet{Language: true}, nil
	}
	return FieldSet{}, ErrInvalidFields{fields}
}

func parseExtraSymbols(symbols string) (FieldSet, error) {
	symbolsPattern := regexp.MustCompile(`^\+q$`)
	if symbols == "" {
		return FieldSet{}, nil
	}
	if symbolsPattern.MatchString(symbols) {
		return FieldSet{ExtraTags: true}, nil
	}
	return FieldSet{}, ErrInvalidFields{fields}
}

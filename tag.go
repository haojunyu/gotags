package main

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Tag结构体： 表示一个tag
type Tag struct {
	Name    string
	File    string
	Address string
	Type    TagType
	Fields  map[TagField]string
}

// TagField类型： 表示tag的字段类型
type TagField string

// Tag fields.
const (
	Access        TagField = "access"
	Signature     TagField = "signature"
	TypeField     TagField = "type"
	ReceiverType  TagField = "ctype"
	Line          TagField = "line"
	InterfaceType TagField = "ntype"
	Language      TagField = "language"
	ExtraTags     TagField = "extraTag"
)

// TagType类： 表示tag的类型
type TagType string

// Tag types.
const (
	Package     TagType = "p"
	Import      TagType = "i"
	Constant    TagType = "c"
	Variable    TagType = "v"
	Type        TagType = "t"
	Interface   TagType = "n"
	Field       TagType = "w"
	Embedded    TagType = "e"
	Method      TagType = "m"
	Constructor TagType = "r"
	Function    TagType = "f"
)

// NewTag函数： 创建一个新tag
func NewTag(name, file string, line int, tagType TagType) Tag {
	l := strconv.Itoa(line)
	return Tag{
		Name:    name,
		File:    file,
		Address: l,
		Type:    tagType,
		Fields:  map[TagField]string{Line: l},
	}
}

// String接口： 结构化输出tag实例
func (t Tag) String() string {
	var b bytes.Buffer

	b.WriteString(t.Name)
	b.WriteByte('\t')
	b.WriteString(t.File)
	b.WriteByte('\t')
	b.WriteString(t.Address)
	b.WriteString(";\"\t")
	b.WriteString(string(t.Type))
	b.WriteByte('\t')

	fields := make([]string, 0, len(t.Fields))
	i := 0
	for k, v := range t.Fields {
		if len(v) == 0 {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s:%s", k, v))
		i++
	}

	sort.Sort(sort.StringSlice(fields))
	b.WriteString(strings.Join(fields, "\t"))

	return b.String()
}

// String接口： 结构化输出tag实例
func (t Tag) Triple() [][]string {
	tagTypeHash := map[TagType]string{
		Package:     "包",
		Import:      "包",
		Constant:    "常量",
		Variable:    "变量",
		Type:        "类型",
		Interface:   "接口",
		Field:       "字段",
		Embedded:    "unknown",
		Method:      "方法",
		Constructor: "构造函数",
		Function:    "函数",
	}
	tagRelaHash := map[TagType]string{
		Package:     "包含",
		Import:      "依赖",
		Constant:    "常量定义",
		Variable:    "变量定义",
		Type:        "类型定义",
		Interface:   "接口定义",
		Field:       "字段定义",
		Embedded:    "unknown定义",
		Method:      "方法定义",
		Constructor: "构造函数定义",
		Function:    "函数定义",
	}

	var triple [][]string
	one := []string{"v", "文件名", t.File, t.File}
	triple = append(triple, one)
	two := []string{"v", tagTypeHash[t.Type], t.Name, t.Name}
	triple = append(triple, two)
	three := []string{"e", tagRelaHash[t.Type], t.File, t.Name, "位置", t.Address}
	if t.Type == Package {
		three = []string{"e", tagRelaHash[t.Type], t.Name, t.File, "位置", t.Address}
	}

	for k, v := range t.Fields {
		if len(v) == 0 {
			continue
		}
		three = append(three, string(k))
		three = append(three, v)
	}
	triple = append(triple, three)
	return triple
	/*
		var b bytes.Buffer

		b.WriteString(t.Name)
		b.WriteByte('\t')
		b.WriteString(t.File)
		b.WriteByte('\t')
		b.WriteString(t.Address)
		b.WriteString("\t")
		b.WriteString(string(t.Type))
		b.WriteByte('\t')

		fields := make([]string, 0, len(t.Fields))
		i := 0
		for k, v := range t.Fields {
			if len(v) == 0 {
				continue
			}
			fields = append(fields, fmt.Sprintf("%s:%s", k, v))
			i++
		}

		sort.Sort(sort.StringSlice(fields))
		b.WriteString(strings.Join(fields, "\t"))

		return b.String()
	*/
}

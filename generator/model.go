package generator

import (
	"encoding/json"
	"fmt"
)

type Struct struct {
	Tags
	Name    string
	Fields  []Field
	Methods []Method
}

func (s Struct) FindMethod(name string) (Method, bool) {
	for _, m := range s.Methods {
		if name == m.Name() {
			return m, true
		}
	}
	return Method{}, false
}

type Method struct {
	FuncName string
	Args     []Field
	Results  []Field
}

func (m Method) Name() string {
	return m.FuncName
}

func (m Method) String() string {
	var s string
	if m.FuncName != "" {
		s += m.FuncName + " "
	}
	s += "func("
	for _, v := range m.Args {
		s += v.String() + ","
	}
	s += ") ("
	for _, v := range m.Results {
		s += v.String() + ","
	}
	s += ")"

	return s
}

func (Method) ZeroCondition(field string) string {
	return fmt.Sprintf("%s == nil", field)
}

type Field struct {
	Tags
	Name string
	Kind Kinder
}

func (f Field) String() string {
	if f.Name == "" {
		return f.Kind.String()
	}
	return f.Name + " " + f.Kind.String()
}

func (f Field) NameOrKindName() string {
	if f.Name != "" {
		return f.Name
	}
	return f.Kind.Name()
}

func (f Field) NameForField() string {
	if f.Name == "" {
		return f.Kind.String()
	}
	return UncapFirst(f.Name)
}

func (f Field) IsNested() bool {
	return f.Name == ""
}

type TypeEnum int

type Kinder interface {
	Name() string
	String() string
	ZeroCondition(string) string
}

type Basic struct {
	Pck  string
	Type string
}

func (b Basic) Name() string {
	if b.Pck != "" {
		return b.Pck + "." + b.Type
	}
	return b.Type
}

func (b Basic) String() string {
	return b.Name()
}

func (b Basic) ZeroCondition(field string) string {
	if zero, ok := zeros[b.Name()]; ok {
		return fmt.Sprintf("%s == %s", field, zero)
	}
	return fmt.Sprintf("(%s == %s{})", field, b.String())
}

var zeros = map[string]string{
	"bool":       "false",
	"string":     `""`,
	"int":        "0",
	"int8":       "0",
	"int16":      "0",
	"int32":      "0",
	"int64":      "0",
	"uint":       "0",
	"uint8":      "0",
	"uint16":     "0",
	"uint32":     "0",
	"uint64":     "0",
	"uintptr":    "nil",
	"byte":       "0",
	"rune":       "0",
	"float32":    "0",
	"float64":    "0",
	"complex64":  "0",
	"complex128": "0",
}

type Pointer struct {
	Kinder
}

func (p Pointer) String() string {
	return "*" + p.Name()
}

type Array struct {
	Kinder
}

func (a Array) String() string {
	return "[]" + a.Kinder.String()
}

func (Array) ZeroCondition(field string) string {
	return fmt.Sprintf("len(%s) == 0", field)
}

type Map struct {
	Key Kinder
	Val Kinder
}

func (m Map) Name() string {
	return "map"
}

func (m Map) String() string {
	return fmt.Sprintf("map[%s]%s", m.Key.String(), m.Val.String())
}

func (Map) ZeroCondition(field string) string {
	return fmt.Sprintf("len(%s) == 0", field)
}

type Tag struct {
	Name string
	Args string
}

func (t Tag) Unmarshal(v interface{}) error {
	if t.Args == "" {
		return nil
	}
	return json.Unmarshal([]byte(t.Args), v)
}

type Tags []Tag

func (t Tags) HasTag(tag string) bool {
	for _, v := range t {
		if v.Name == tag {
			return true
		}
	}
	return false
}

func (t Tags) FindTag(tag string) (Tag, bool) {
	for _, t := range t {
		if t.Name == tag {
			return t, true
		}
	}
	return Tag{}, false
}

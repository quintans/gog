package generator

import (
	"encoding/json"
	"fmt"
	"strings"
)

const zeroNil = "nil"

type MapperType string

const (
	StructMapper    MapperType = "struct"
	InterfaceMapper MapperType = "interface"
)

type Mapper interface {
	Type() MapperType
	GetTags() Tags
	GetName() string
	GetFields() []Field
	GetMethods() []Method
	AddMethod(Method)
	FindMethod(name string) (Method, bool)
}

type Struct struct {
	Tags
	Name    string
	Fields  []Field
	Methods []Method
}

func (s *Struct) Type() MapperType {
	return StructMapper
}

func (s *Struct) GetTags() Tags {
	return s.Tags
}

func (s *Struct) GetName() string {
	return s.Name
}

func (s *Struct) GetFields() []Field {
	return s.Fields
}

func (s *Struct) GetMethods() []Method {
	return s.Methods
}

func (s *Struct) AddMethod(m Method) {
	s.Methods = append(s.Methods, m)
}

func (s *Struct) FindMethod(name string) (Method, bool) {
	for _, m := range s.Methods {
		if name == m.Name() {
			return m, true
		}
	}
	return Method{}, false
}

type Interface struct {
	Tags
	Name    string
	Methods []Method
}

func (s *Interface) Type() MapperType {
	return InterfaceMapper
}

func (s *Interface) GetTags() Tags {
	return s.Tags
}

func (s *Interface) GetName() string {
	return s.Name
}

func (s *Interface) GetFields() []Field {
	return nil
}

func (s *Interface) GetMethods() []Method {
	return s.Methods
}

func (s *Interface) AddMethod(m Method) {
	s.Methods = append(s.Methods, m)
}

func (s *Interface) FindMethod(name string) (Method, bool) {
	for _, m := range s.Methods {
		if name == m.Name() {
			return m, true
		}
	}
	return Method{}, false
}

type Method struct {
	Tags
	FuncName string
	Args     []Field
	Results  []Field
}

func (m *Method) IsExported() bool {
	if m.FuncName == "" {
		return false
	}
	first := m.FuncName[:1]
	upper := strings.ToUpper(first)
	return first == upper
}

func (m *Method) Name() string {
	return m.FuncName
}

func (m *Method) String() string {
	s := &Scribler{}
	if m.FuncName != "" {
		s.BPrint(m.FuncName, " ")
	}
	s.BPrint("func")
	s.BPrint("(", m.Parameters(false), ")")
	s.BPrint(" ")
	s.BPrint("(", m.Returns(), ")")

	return s.String()
}

func (m *Method) HasResults() bool {
	return len(m.Results) > 0
}

func (m *Method) Signature(withName bool) string {
	s := &Scribler{}
	if withName && m.FuncName != "" {
		s.BPrint(m.FuncName)
	}
	s.BPrint("(", m.Parameters(false), ")")
	s.BPrint(" ")
	s.BPrint("(", m.Returns(), ")")

	return s.String()
}

func (m *Method) Call(withName bool) string {
	s := &Scribler{}
	if withName && m.FuncName != "" {
		s.BPrint(m.FuncName)
	}
	s.BPrint("(", m.Parameters(false), ")")
	s.BPrint(" ")
	s.BPrint("(", m.Returns(), ")")

	return s.String()
}

func (m *Method) Parameters(onlyName bool) string {
	args := make([]string, len(m.Args))
	if onlyName {
		for k, v := range m.Args {
			args[k] = v.Name
		}
	} else {
		for k, v := range m.Args {
			args[k] = v.String()
		}
	}

	return strings.Join(args, ",")
}

func (m *Method) Returns() string {
	res := make([]string, len(m.Results))
	for k, v := range m.Results {
		res[k] = v.String()
	}

	return strings.Join(res, ",")
}

func (m *Method) ReturnZerosWithError(errVar string) string {
	res := make([]string, len(m.Results))
	for k, v := range m.Results {
		if v.IsError() {
			res[k] = errVar
		} else {
			res[k] = v.Kind.Zero()
		}
	}

	return strings.Join(res, ",")
}

func (Method) ZeroCondition(field string) string {
	return fmt.Sprintf("%s == nil", field)
}

func (Method) Zero() string {
	return zeroNil
}

func (m *Method) ContextArgName() string {
	for _, a := range m.Args {
		if a.IsContext() {
			return a.Name
		}
	}
	return ""
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

func (f Field) IsPrimitive() bool {
	_, ok := Zero(f.Kind.String())
	return ok
}

func (f Field) IsError() bool {
	return f.Kind.Name() == "error"
}

func (f Field) IsContext() bool {
	return f.Kind.Name() == "context.Context"
}

type TypeEnum int

type Kinder interface {
	Name() string
	String() string
	ZeroCondition(string) string
	Zero() string
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

func (b Basic) Zero() string {
	if zero, ok := zeros[b.Name()]; ok {
		return zero
	}
	return fmt.Sprintf("%s{}", b.String())
}

func Zero(typ string) (string, bool) {
	z, ok := zeros[typ]
	return z, ok
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
	"uintptr":    zeroNil,
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

func (Pointer) ZeroCondition(field string) string {
	return fmt.Sprintf("%s == nil", field)
}

func (p Pointer) Zero() string {
	return zeroNil
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

func (a Array) Zero() string {
	return zeroNil
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

func (m Map) Zero() string {
	return zeroNil
}

type InterfaceVar struct {
	Tags
	Pck     string
	Type    string
	Methods []Method
}

func (b *InterfaceVar) Name() string {
	if b.Pck == "" && b.Type == "" {
		return "interface{}"
	}

	if b.Pck != "" {
		return b.Pck + "." + b.Type
	}

	return b.Type
}

func (b *InterfaceVar) String() string {
	return b.Name()
}

func (b *InterfaceVar) ZeroCondition(field string) string {
	return fmt.Sprintf("%s == nil", field)
}

func (b *InterfaceVar) Zero() string {
	return zeroNil
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

func (t Tags) Filter(filter ...string) []Tag {
	var tags []Tag

	for _, t := range t {
		for _, f := range filter {
			if t.Name == f {
				tags = append(tags, t)
			}
		}
	}

	return tags
}

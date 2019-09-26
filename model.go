package main

import "fmt"

type Struct struct {
	Tags
	Name    string
	Fields  []Field
	Methods []Method
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

type TypeEnum int

type Kinder interface {
	Name() string
	String() string
}

type Basic struct {
	Type string
}

func (b Basic) Name() string {
	return b.Type
}

func (b Basic) String() string {
	return b.Type
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

type Tag struct {
	Name string
	Args string
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

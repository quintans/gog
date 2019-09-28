package plugins

import (
	"strings"

	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&ValueObj{})
}

// Equaler defines the interface to be implemented by struct that wish to verify equality
type Equaler interface {
	Equals(o interface{}) bool
}

// Cloner defines the interface to be implemented by struct that wish to create a copy of them selfs
type Cloner interface {
	Clone() interface{}
}

type ValueObj struct {
	generator.Scribler
}

func (b *ValueObj) Name() string {
	return "value"
}

func (b *ValueObj) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{
		"github.com/quintans/gog": "",
	}
}

func (b *ValueObj) Generate(mapper generator.Struct) []byte {
	for _, field := range mapper.Fields {
		if !field.HasTag(IgnoreTag) {
			switch t := field.Kind.(type) {
			case generator.Map:
				b.genMap(mapper, field.NameOrKindName(), t)
			case generator.Array:
				b.genArray(mapper, field.NameOrKindName(), t)
			default:
				b.genDefault(mapper, field)
			}
		}
	}

	return b.Flush()
}

func (b *ValueObj) genDefault(mapper generator.Struct, field generator.Field) {
	fieldName := field.NameOrKindName()
	b.Printf("\nfunc (t %s) %s() %s {\n", mapper.Name, strings.Title(fieldName), field.Kind.String())
	b.Printf("  return t.%s\n", fieldName)
	b.Printf("}\n")
}

func (b *ValueObj) genArray(mapper generator.Struct, fieldName string, arr generator.Array) {
	subKind := arr.Kinder.String()
	b.Printf("\nfunc (t %s) %s() []%s {\n", mapper.Name, strings.Title(fieldName), subKind)
	b.Printf("	if len(t.%s) == 0 {\n", fieldName)
	b.Printf("		return t.%s\n", fieldName)
	b.Printf("	}\n\n")

	b.Printf("	others := make([]%s, len(t.%s))\n", subKind, fieldName)
	b.Printf("	// check that the type of the array is clonable\n")
	b.Printf("	if _, ok := t.%s[0].(gog.Cloner); ok {\n", fieldName)
	b.Printf("		for k, v := range t.%s {\n", fieldName)
	b.Printf("			t := v.(gog.Cloner)\n")
	b.Printf("			others[k] = t.Clone().(%s)\n", subKind)
	b.Printf("		}\n")
	b.Printf("		return others\n")
	b.Printf("	}\n\n")

	b.Printf("	for k, v := range t.%s {\n", fieldName)
	b.Printf("		others[k] = t.%s[k]\n", fieldName)
	b.Printf("	}\n")
	b.Printf("	return others\n")
	b.Printf("}\n")
}

func (b *ValueObj) genMap(mapper generator.Struct, fieldName string, m generator.Map) {
	keyKind := m.Key.String()
	valKind := m.Val.String()
	b.Printf("\nfunc (t %s) %s() map[%s]%s {\n", mapper.Name, strings.Title(fieldName), keyKind, valKind)
	b.Printf("	if len(t.%s) == 0 {\n", fieldName)
	b.Printf("		return t.%s\n", fieldName)
	b.Printf("	}\n\n")

	b.Printf("	others := map[%s]%s{}\n", keyKind, valKind)
	b.Printf("	for k, v := range t.%s {\n", fieldName)
	b.Printf("		var key %s\n", keyKind)
	b.Printf("		if t, ok := k.(gog.Cloner); ok {\n")
	b.Printf("			key = t.Clone().(%s)\n", keyKind)
	b.Printf("		} else {\n")
	b.Printf("			key = k\n")
	b.Printf("		}\n")
	b.Printf("		var val %s\n", valKind)
	b.Printf("		if t, ok := v.(gog.Cloner); ok {\n")
	b.Printf("			val = t.Clone().(%s)\n", valKind)
	b.Printf("		} else {\n")
	b.Printf("			val = v\n")
	b.Printf("		}\n")
	b.Printf("		others[key] = val\n")
	b.Printf("	}\n")
	b.Printf("	return others\n")
	b.Printf("}\n")
}

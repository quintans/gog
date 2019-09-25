package main

type Struct struct {
	Tags
	Name    string
	Fields  []Field
	Methods []Method
}

type Field struct {
	Tags
	Name string
	Kind Kind
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
	return f.Kind.Name
}

type Kind struct {
	Name    string
	Pointer bool
	Array   bool
	Args    []Field
	Results []Field
}

type Method struct {
	Name    string
	Args    []Field
	Results []Field
}

func (k Kind) IsFunc() bool {
	return k.Args != nil
}

func (p Kind) String() string {
	var s string
	if p.Array {
		s += "[]"
	}
	if p.Pointer {
		s += "*"
	}
	if p.IsFunc() {
		s += "func("
		for _, v := range p.Args {
			s += v.String() + ","
		}
		s += ") ("
		for _, v := range p.Results {
			s += v.String() + ","
		}
		s += ")"
	} else {
		s += p.Name
	}
	return s
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

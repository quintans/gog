package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const gofilesExt = ".go"

const (
	gogPrefix   = "//gog:"
	requiredTag = "@required"
)

var (
	fileName = flag.String("f", "", "file name to be parsed, overriding the environment variable GOFILE value")
	recur    = flag.Bool("r", false, "scan current dir and sub directories")
)

type Generator interface {
	Imports(Struct) map[string]string
	Generate(Struct) []byte
	Name() string
}

func main() {
	flag.Parse()

	fileToParse := getFileToParse()
	if fileToParse != "" {
		parseGoFileAndGenerateFile(fileToParse)
		return
	}

	if *recur {
		scanCurrentDirAndSubDirs()
		return
	}

	scanCurrentDir()
}

func scanCurrentDir() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		parseGoFileIfTagged(file.Name())
	}
}

func scanCurrentDirAndSubDirs() {
	currentDir := "."
	err := filepath.Walk(currentDir, func(path string, file os.FileInfo, err error) error {
		parseGoFileIfTagged(path)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func parseGoFileIfTagged(name string) {
	if filepath.Ext(name) == gofilesExt && isTagged(name) {
		parseGoFileAndGenerateFile(name)
	}
}

func isTagged(gofile string) bool {
	file, err := os.Open(gofile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// for now we are just handling tagged structs
		if strings.HasPrefix(line, gogPrefix) {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return false
}

func getFileToParse() string {
	if *fileName != "" {
		return *fileName
	}

	return os.Getenv("GOFILE")
}

func parseGoFileAndGenerateFile(gofile string) {
	p := parseGoFile(gofile)

	var name = strings.Split(gofile, ".")[0]
	fileName := fmt.Sprintf("%s_gog.go", name)
	p.generateGoFile(fileName)
}

func parseGoFile(gofile string) *Parser {
	log.Println("Parsing", gofile)

	fs := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fs, gofile, nil, parser.ParseComments)
	die(err, "parsing package: %s", gofile)

	g := NewParser(parsedFile)

	ast.Inspect(parsedFile, g.genImp)
	ast.Inspect(parsedFile, g.genDecl)
	ast.Inspect(parsedFile, g.funcDecl)

	return g
}

func (p *Parser) generateGoFile(filename string) {
	code := p.generateCode()
	err := ioutil.WriteFile(filename, code, 0644)
	die(err, "Writing output")
}

func (p *Parser) generateCode() []byte {
	p.HPrintf("// Code generated by gog; DO NOT EDIT.\n")
	//g.Printf("// Generate at %s\n", time.Now().Format("2006-01-02 15:04:05 -0700"))
	p.HPrintf("package %s\n\n", p.parsedFile.Name.Name)

	for _, mapper := range p.Structs {
		for _, tag := range mapper.Tags {
			gen, ok := p.generators[tag.Name]
			if !ok {
				log.Printf("Could not find generator for %s", tag)
				continue
			}
			imports := gen.Imports(*mapper)
			for path, name := range imports {
				p.Imports[path] = name
			}

			src := gen.Generate(*mapper)
			p.Printf("\n")

			p.Printf("\n // Generated by gog:%s\n\n", gen.Name())
			p.Body.Write(src)
		}
	}

	for path, name := range p.Imports {
		p.HPrintf("import %s\"%s\"\n", name+" ", path)
	}

	return formatCode(p.Flush())
}

func formatCode(source []byte) []byte {
	src, err := format.Source(source)
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return source
	}
	return src
}

func die(err error, msg string, args ...interface{}) {
	if err != nil {
		s := fmt.Sprintf(msg, args...)
		log.Fatal(s+":", err)
	}
}

type Struct struct {
	Name    string
	Fields  []Field
	Tags    []Tag
	Methods []Method
}

type Tag struct {
	Name string
	Args string
}

type Field struct {
	Name     string
	Kind     Kind
	Required bool
}

func (f Field) String() string {
	return f.Name + " " + f.Kind.String()
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

type Parser struct {
	Scribler

	Imports    map[string]string
	Structs    []*Struct
	generators map[string]Generator
	parsedFile *ast.File
}

func NewParser(parsedFile *ast.File) *Parser {
	p := &Parser{
		Imports:    make(map[string]string),
		Structs:    make([]*Struct, 0),
		generators: make(map[string]Generator),
		parsedFile: parsedFile,
	}
	p.Register(&Builder{})
	p.Register(&Getters{})
	p.Register(&AllArgsConstructor{})

	return p
}

func (p *Parser) Register(gen Generator) {
	name := gen.Name()
	// TODO: don't allow if 'name' already exists
	p.generators[name] = gen
	log.Printf("Registering generator: %s\n", name)
}

func (p *Parser) genImp(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.IMPORT {
		// We only care about type declarations.
		return true
	}
	for _, spec := range decl.Specs {
		timport := spec.(*ast.ImportSpec)
		var name string
		if timport.Name != nil {
			name = timport.Name.Name
		}
		p.Imports[timport.Path.Value] = name
	}
	return false
}

func (p *Parser) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.TYPE {
		// We only care about type declarations.
		return true
	}
	for _, spec := range decl.Specs {
		tspec := spec.(*ast.TypeSpec)
		iType, ok := tspec.Type.(*ast.StructType)
		if ok && p.isMarkedForGeneration(decl) {
			aStruct := &Struct{
				Name:    tspec.Name.Name,
				Fields:  make([]Field, 0),
				Methods: make([]Method, 0),
			}
			p.Structs = append(p.Structs, aStruct)
			for _, astField := range iType.Fields.List {
				field := parseField(astField)
				aStruct.Fields = append(aStruct.Fields, field)
			}
			aStruct.Tags = extractTagsFromDoc(decl.Doc)
		}
	}

	return false
}

func (p *Parser) funcDecl(node ast.Node) bool {
	fn, ok := node.(*ast.FuncDecl)
	if !ok {
		return true
	}
	if fn.Recv != nil && len(fn.Recv.List) == 1 {
		field := fn.Recv.List[0]
		var expr ast.Expr
		startExpr, ok := field.Type.(*ast.StarExpr)
		if ok {
			expr = startExpr.X
		} else {
			expr = field.Type.(ast.Expr)
		}
		ident := expr.(*ast.Ident)
		for _, s := range p.Structs {
			if ident.Name == s.Name {
				// add to the list of methods
				kind := parseType(fn.Type)
				method := Method{
					Name:    fn.Name.Name,
					Args:    kind.Args,
					Results: kind.Results,
				}
				s.Methods = append(s.Methods, method)
			}
		}
	}

	return false

}

func (p *Parser) isMarkedForGeneration(decl *ast.GenDecl) bool {
	for _, com := range decl.Doc.List {
		if p.hasValidGenerationPrefix(com.Text) {
			return true
		}
	}
	return false
}

func (p *Parser) hasValidGenerationPrefix(text string) bool {
	for tag := range p.generators {
		if strings.HasPrefix(text, gogPrefix+tag) {
			return true
		}
	}
	return false
}

func extractTagsFromDoc(doc *ast.CommentGroup) []Tag {
	tags := make([]Tag, 0)
	if doc == nil {
		return tags
	}

	docs := make([]string, 0)
	for _, com := range doc.List {
		docs = append(docs, com.Text)
	}

	for _, line := range docs {
		if strings.HasPrefix(line, gogPrefix) {
			tag, arg := splitIntoTagAndArgs(line)
			tags = append(tags, Tag{tag, arg})
		}
	}
	return tags
}

func splitIntoTagAndArgs(line string) (string, string) {
	str := strings.TrimSpace(line)
	offset := len(gogPrefix)
	firstSpace := strings.Index(str, " ")
	if firstSpace == -1 {
		return str[offset:], ""
	}
	return str[offset:firstSpace], str[offset+firstSpace:]
}

func parseField(astField *ast.Field) Field {
	//fmt.Println("====> Comment:", astField.Doc.Text())
	var field Field
	field.Kind = parseType(astField.Type)
	if len(astField.Names) > 0 {
		field.Name = astField.Names[0].Name
	}

	tags := extractTagsFromDoc(astField.Doc)
	_, ok := findTag(tags, requiredTag)
	field.Required = ok
	return field
}

func findTag(tags []Tag, tagName string) (Tag, bool) {
	for _, t := range tags {
		if t.Name == tagName {
			return t, true
		}
	}
	return Tag{}, false
}

func parseType(expr ast.Expr) Kind {
	var kind Kind
	switch n := expr.(type) {
	// if the type is imported
	case *ast.ArrayType:
		kind = parseType(n.Elt)
		kind.Array = true
	case *ast.SelectorExpr:
		pck := n.X.(*ast.Ident)
		kind.Name = pck.Name + "." + n.Sel.Name
	case *ast.StarExpr:
		kind = parseType(n.X)
		kind.Pointer = true
	case *ast.Ident:
		kind.Name = n.Name
	case *ast.FuncType:
		kind.Args = make([]Field, 0)
		kind.Results = make([]Field, 0)
		for _, p := range n.Params.List {
			//fmt.Printf("====> Param: %s, %#v\n", p.Type, p.Type)
			arg := parseField(p)
			kind.Args = append(kind.Args, arg)
		}
		for _, res := range n.Results.List {
			result := parseField(res)
			kind.Results = append(kind.Results, result)
		}
	}
	return kind
}

func UncapFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

type Scribler struct {
	Header bytes.Buffer
	Body   bytes.Buffer
}

func (s *Scribler) HPrintf(format string, args ...interface{}) {
	fmt.Fprintf(&s.Header, format, args...)
}

func (s *Scribler) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&s.Body, format, args...)
}

func (s *Scribler) Flush() []byte {
	head := s.Header.Bytes()
	s.Header.Reset()
	body := s.Body.Bytes()
	s.Body.Reset()
	return append(head, body...)
}

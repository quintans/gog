package main

import (
	"bufio"
	"bytes"
	"errors"
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
const genGogSuffix = "_gog.go"

const (
	gogPrefix  = "//gog:"
	builderTag = gogPrefix + "builder"
	getterTag  = gogPrefix + "getter"
)

var validStructTags = [...]string{builderTag, getterTag}

var (
	fileName = flag.String("f", "", "file name to be parsed, overriding the environment variable GOFILE value")
	recur    = flag.Bool("r", false, "scan current dir and sub directories")
)

func main() {
	flag.Parse()

	fileToParse := getFileToParse()
	if fileToParse != "" {
		parseGoFile(fileToParse)
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
	if filepath.Ext(name) == gofilesExt {
		if isTagged(name) {
			parseGoFile(name)
		}
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

func parseGoFile(gofile string) {
	log.Println("Parsing", gofile)
	var name = strings.Split(gofile, ".")[0]

	fs := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fs, gofile, nil, parser.ParseComments)
	die(err, "parsing package: %s", gofile)

	g := Generator{
		imports: make(map[string]string),
		mappers: make([]*Mapper, 0),
	}

	ast.Inspect(parsedFile, g.genImp)
	ast.Inspect(parsedFile, g.genDecl)

	g.generate(parsedFile)

	src := formatCode(g.final.Bytes())
	err = ioutil.WriteFile(name+genGogSuffix, src, 0644)
	die(err, "Writing output")
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

type Mapper struct {
	docs   []string
	name   string
	fields []Field
}

type Field struct {
	name string
	kind Kind
}

func (f Field) String() string {
	return f.name + " " + f.kind.String()
}

type Kind struct {
	name    string
	pointer bool
	array   bool
	args    []Field
	results []Field
}

func (k Kind) isFunc() bool {
	return k.args != nil
}

func (p Kind) String() string {
	var s string
	if p.array {
		s += "[]"
	}
	if p.pointer {
		s += "*"
	}
	if p.args != nil {
		s += "func("
		for _, v := range p.args {
			s += v.String() + ","
		}
		s += ") ("
		for _, v := range p.results {
			s += v.String() + ","
		}
		s += ")"
	}
	s += p.name
	return s
}

type StructScribler interface {
	Scrible(mapper *Mapper)
}

type Generator struct {
	imports map[string]string
	final   bytes.Buffer
	mappers []*Mapper
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.final, format, args...)
}

func (g *Generator) genImp(node ast.Node) bool {
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
		g.imports[timport.Path.Value] = name
	}
	return false
}

func (g *Generator) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.TYPE {
		// We only care about type declarations.
		return true
	}
	for _, spec := range decl.Specs {
		tspec := spec.(*ast.TypeSpec)
		iType, ok := tspec.Type.(*ast.StructType)
		if ok && isMarkedForGeneration(decl) {
			aStruct := &Mapper{
				docs:   make([]string, 0),
				name:   tspec.Name.Name,
				fields: make([]Field, 0),
			}
			g.mappers = append(g.mappers, aStruct)
			for _, astField := range iType.Fields.List {
				field := parseField(astField)
				aStruct.fields = append(aStruct.fields, field)
			}
			for _, com := range decl.Doc.List {
				aStruct.docs = append(aStruct.docs, com.Text)
			}
		}
	}

	return false
}

func isMarkedForGeneration(decl *ast.GenDecl) bool {
	for _, com := range decl.Doc.List {
		if hasValidGenerationPrefix(com.Text) {
			return true
		}
	}
	return false
}

func hasValidGenerationPrefix(text string) bool {
	for _, tag := range validStructTags {
		if strings.HasPrefix(text, tag) {
			return true
		}
	}
	return false
}

func (g *Generator) generate(parsedFile *ast.File) {
	g.Printf("// Code generated by gog; DO NOT EDIT.\n")
	//g.Printf("// Generate at %s\n", time.Now().Format("2006-01-02 15:04:05 -0700"))
	g.Printf("package %s\n\n", parsedFile.Name.Name)

	for path, name := range g.imports {
		g.Printf("import %s%s\n", name+" ", path)
	}

	for _, mapper := range g.mappers {
		tags := extractStructTags(mapper)

		for _, tag := range tags {
			scribler, err := g.makeStructScribler(tag)
			if err != nil {
				panic(err)
			}
			scribler(mapper)
		}
	}
}

func extractStructTags(mapper *Mapper) []string {
	tags := make([]string, 0)
	for _, line := range mapper.docs {
		if strings.HasPrefix(line, gogPrefix) {
			tag := fetchTag(line)
			tags = append(tags, tag)
		}
	}
	return tags
}

func fetchTag(line string) string {
	firstSpace := strings.Index(line, " ")
	if firstSpace == -1 {
		return line
	}
	return line[:firstSpace]
}

func (g *Generator) makeStructScribler(name string) (scribler func(*Mapper), err error) {
	switch name {
	case builderTag:
		scribler = g.generateBuilder
	case getterTag:
		scribler = g.generateGetters
	default:
		err = errors.New("Unknown scribler name " + name)
	}
	return
}

func (g *Generator) generateBuilder(mapper *Mapper) {
	structName := mapper.name + "Builder"
	g.Printf("\ntype %s struct {", structName)
	for _, field := range mapper.fields {
		g.Printf("\n\t%s", uncapFirst(field.String()))
	}
	g.Printf("\n}\n")

	g.Printf("\nfunc New%s() %s { return %s{} }\n", structName, mapper.name, mapper.name)

	for _, field := range mapper.fields {
		g.Printf("\nfunc (b *%s) %s(%s) *%s {\n", structName, strings.Title(field.name), field.String(), structName)
		g.Printf("	b.%s = %s\n", uncapFirst(field.name), field.name)
		g.Printf("  return b\n")
		g.Printf("}\n")
	}

	g.Printf("\n\nfunc (b *%s) Build() %s {", structName, mapper.name)
	g.Printf("\nreturn %s{\n", mapper.name)
	for _, field := range mapper.fields {
		g.Printf("%s: b.%s,\n", field.name, uncapFirst(field.name))
	}
	g.Printf("}\n}")

	g.Printf("\n\nfunc (src %s) ToBuild() %s {", mapper.name, structName)
	g.Printf("\nreturn %s{\n", structName)
	for _, field := range mapper.fields {
		g.Printf("%s: src.%s,\n", uncapFirst(field.name), field.name)
	}
	g.Printf("}\n}\n")
}

func (g *Generator) generateGetters(mapper *Mapper) {
	for _, field := range mapper.fields {
		g.Printf("\nfunc (t %s) %s() %s {\n", mapper.name, strings.Title(field.name), field.kind.String())
		g.Printf("  return t.%s\n", field.name)
		g.Printf("}\n")
	}
}

func parseField(astField *ast.Field) Field {
	//fmt.Println("====> Comment:", astField.Doc.Text())
	var field Field
	field.kind = parseType(astField.Type)
	if len(astField.Names) > 0 {
		field.name = astField.Names[0].Name
	}
	return field
}

func parseType(expr ast.Expr) Kind {
	var kind Kind
	switch n := expr.(type) {
	// if the type is imported
	case *ast.ArrayType:
		kind = parseType(n.Elt)
		kind.array = true
	case *ast.SelectorExpr:
		pck := n.X.(*ast.Ident)
		kind.name = pck.Name + "." + n.Sel.Name
	case *ast.StarExpr:
		kind = parseType(n.X)
		kind.pointer = true
	case *ast.Ident:
		kind.name = n.Name
	case *ast.FuncType:
		kind.args = make([]Field, 0)
		kind.results = make([]Field, 0)
		for _, p := range n.Params.List {
			//fmt.Printf("====> Param: %s, %#v\n", p.Type, p.Type)
			arg := parseField(p)
			kind.args = append(kind.args, arg)
		}
		for _, res := range n.Results.List {
			result := parseField(res)
			kind.results = append(kind.results, result)
		}
	}
	return kind
}

func uncapFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

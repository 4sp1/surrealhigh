// Types is a tool to automate the creation of methods that satisfy the
// interfaces Doc, DocID, Field, Thing, etc... Given the name of a struct
// type T, types will create types to safely make new type safe surrealdb
// adapters using surrealhigh canonical interface.
//
// # The file is created in the same package and directory as the package that defines T
//
// For example, given this snippet,
//
//	package records
//
//	import "time"
//
//	type Record struct {
//		T time.Time
//		N string
//	}
//
// running this command
//
//	surrealhigh-gen-types -doc=Record -pkg=records [.]
//
// in the same directory will create the file sigh_doc_record.go, in package records,
// containing the toolkit you can use to build surrealhigh queries.
//
// Adapted from https://cs.opensource.google/go/x/tools/+/refs/tags/v0.10.0:cmd/stringer/stringer.go;bpv=0
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"strings"

	"github.com/4sp1/surrealhigh"
	"github.com/4sp1/surrealhigh/templates/jennifer"
	"github.com/rs/zerolog/log"
	"golang.org/x/tools/go/packages"
)

var (
	doc = flag.String("doc", "", "comma-separated list of doc struct names; must be set")
	pkg = flag.String("pkg", "", "destination package")
)

func main() {

	flag.Parse()
	if len(*doc) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	docs := strings.Split(*doc, ",")

	if len(*pkg) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."} // where to look at
	}
	tags := []string{} // build tags

	g := Generator{}
	g.parsePackage(args, tags)

	for _, docName := range docs {
		g.generate(docName)
		for _, f := range g.pkg.files {
			for _, v := range f.values {
				fmt.Println("file", f.file.Name)
				for _, i := range f.file.Imports {
					fmt.Println("import", i.Name, i.Path.Value)
					if i.Name == nil {
						// TODO(malikbenkirane) todo)) auto import
					}
				}
				fmt.Println(v)
				if err := jennifer.NewDoc(
					surrealhigh.Package(*pkg),
					surrealhigh.Table(strings.ToLower(v.structName)),
					v.docFields()...).Write(os.Stdout); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}
	}

}

type Generator struct {
	pkg *Package
}

type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}
	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file: file,
			pkg:  g.pkg,
		}
	}
}

// Value represents a declared struct.
type Value struct {
	structName string
	fields     []Field
}

func (v Value) String() string {
	var fields []string
	for _, field := range v.fields {
		fields = append(fields, field.String())
	}
	return fmt.Sprintf("%s %s", v.structName, strings.Join(fields, " "))
}

func (v Value) docFields() (fields []jennifer.DocField) {
	for _, field := range v.fields {

		field.generateTypeIdents() // prepare fieldIdent, typeQual, and isPointer

		var opts []jennifer.NewFieldOption
		if qual := field.typeQual; qual != "" {
			opts = append(opts, jennifer.NewFieldWithQual(qual))
			log.Trace().Str("qual", qual).Msg("field.typeQual")
		}
		if field.isPointer {
			opts = append(opts, jennifer.NewFieldWithPointer())
		}

		log.Trace().
			Str("fieldName", field.fieldName).
			Str("typeIdent", field.typeIdent).
			Str("typeQual", field.typeQual).
			Bool("isptr", field.isPointer).
			Msg("docFields")

		fields = append(fields, jennifer.NewField(field.fieldName, field.typeIdent, opts...))

	}
	return
}

type Field struct {
	fieldName  string
	typeExpr   ast.Expr
	typeIdents []string

	typeQual  string
	typeIdent string
	isPointer bool
}

func (f Field) String() string {
	if f.typeIdents == nil {
		f.generateTypeIdents()
	}
	return fmt.Sprintf("[ptr:%v]%v(%v)", f.isPointer, f.typeIdents, f.fieldName)
}

func typeIdents(e ast.Expr, star bool) ([]string, bool) {
	switch t := e.(type) {
	case *ast.Ident:
		return []string{t.Name}, star
	case *ast.StarExpr:
		i, star := typeIdents(t.X, true)
		return i, star
	case *ast.SelectorExpr:
		i, star := typeIdents(t.X, star)
		return append(i, t.Sel.Name), star
	default:
		log.Fatal().Msgf("%#v", t)
	}
	return nil, false
}

func (f *Field) generateTypeIdents() {
	f.typeIdents, f.isPointer = typeIdents(f.typeExpr, false)
	if len(f.typeIdents) == 2 {
		f.typeQual = f.typeIdents[0]
		f.typeIdent = f.typeIdents[1]
		return
	}
	if len(f.typeIdents) != 1 {
		log.Fatal().Msgf("%v", f.typeIdents)
	}
	f.typeIdent = f.typeIdents[0]
}

func (g *Generator) generate(structName string) {
	for _, file := range g.pkg.files {
		file.structName = structName
		file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, file.walk)
		}
	}
}

type File struct {
	pkg  *Package
	file *ast.File

	structName string
	values     []Value
}

func (f *File) walk(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.TYPE {
		return true
	}
	for _, spec := range decl.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			continue
		}
		var fields []Field
		for _, field := range st.Fields.List {
			for _, name := range field.Names {
				fields = append(fields, Field{
					fieldName: name.Name,
					typeExpr:  field.Type,
				})
			}
		}
		structName := ts.Name.Name
		if structName != f.structName {
			continue
		}
		f.values = append(f.values, Value{
			structName: structName,
			fields:     fields,
		})
	}
	return false
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *Generator) parsePackage(patterns []string, tags []string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	if len(pkgs) != 1 {
		log.Fatal().Msgf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

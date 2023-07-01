package jennifer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"os"
	"strings"

	"github.com/4sp1/surrealhigh"
	"github.com/rs/zerolog/log"
	"golang.org/x/tools/go/packages"
)

func NewGen(args, tags, docs []string, pkg, out string) error {
	g := Generator{out: os.Stdout}
	if len(out) > 0 {
		f, err := os.Create(out)
		if err != nil {
			return fmt.Errorf("out: %w", err)
		}
		g.out = f
		defer f.Close()
	}
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
				if err := NewDoc(
					surrealhigh.Package(pkg),
					surrealhigh.Table(strings.ToLower(v.structName)),
					v.docFields()...).Write(g.out); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type Generator struct {
	pkg *Package
	out io.Writer
}

type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*FileGen
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*FileGen, len(pkg.Syntax)),
	}
	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &FileGen{
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

func (v Value) docFields() (fields []DocField) {
	for _, field := range v.fields {

		field.generateTypeIdents() // prepare fieldIdent, typeQual, and isPointer

		var opts []NewFieldOption
		if qual := field.typeQual; qual != "" {
			opts = append(opts, NewFieldWithQual(qual))
			log.Trace().Str("qual", qual).Msg("field.typeQual")
		}
		if field.isPointer {
			opts = append(opts, NewFieldWithPointer())
		}
		if field.isArray {
			opts = append(opts, NewFieldWithArray())
		}

		log.Trace().
			Str("fieldName", field.fieldName).
			Str("typeIdent", field.typeIdent).
			Str("typeQual", field.typeQual).
			Bool("isptr", field.isPointer).
			Msg("docFields")

		fields = append(fields, NewField(field.fieldName, field.typeIdent, opts...))

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
	isArray   bool
}

func (f Field) String() string {
	if f.typeIdents == nil {
		f.generateTypeIdents()
	}
	return fmt.Sprintf("[ptr:%v]%v(%v)", f.isPointer, f.typeIdents, f.fieldName)
}

func typeIdents(e ast.Expr, star bool, arr bool) ([]string, bool, bool) {
	switch t := e.(type) {
	case *ast.Ident:
		return []string{t.Name}, star, arr
	case *ast.StarExpr:
		i, star, arr := typeIdents(t.X, true, arr)
		return i, star, arr
	case *ast.SelectorExpr:
		i, star, arr := typeIdents(t.X, star, arr)
		return append(i, t.Sel.Name), star, arr
	case *ast.ArrayType:
		i, star, arr := typeIdents(t.Elt, star, true)
		return i, star, arr
	default:
		log.Fatal().Msgf("%#v", t)
	}
	return nil, false, false
}

func (f *Field) generateTypeIdents() {
	f.typeIdents, f.isPointer, f.isArray = typeIdents(f.typeExpr, false, false)
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

type FileGen struct {
	pkg  *Package
	file *ast.File

	structName string
	values     []Value
}

func (f *FileGen) walk(node ast.Node) bool {
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
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
//	type record struct {
//		t time.Time
//		n string
//	}
//
// running this command
//
//	surrealhigh-gen-types -doc=record -pkg=records [-o out.go] [.]
//
// in the same directory will create the file sigh_doc_record.go, in package records,
// containing the toolkit you can use to build surrealhigh queries.
//
// Adapted from https://cs.opensource.google/go/x/tools/+/refs/tags/v0.10.0:cmd/stringer/stringer.go;bpv=0
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/4sp1/surrealhigh/templates/jennifer"
)

var (
	doc = flag.String("doc", "", "comma-separated list of doc struct names; must be set")
	pkg = flag.String("pkg", "", "destination package")
	out = flag.String("o", "", "destination file .go")
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

	if err := jennifer.NewGen(args, tags, docs, *pkg, *out); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

// Command update regenerates the Go generated-code reference pages from the
// source of a generated config package.
//
// The pages mirror what go doc reports, minus the project-specific parts: type
// aliases, the config table fields of the atlas struct, and everything
// unexported. Hand-written prose lives in the overlay directory, never in the
// generated output.
//
// This command is not part of the module's package graph — its parent
// directory starts with an underscore, so ./... skips it. Run it explicitly:
//
//	go run ./docs/__impl/update
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// target maps one source file of the generated package to one page.
type target struct {
	src   string
	out   string
	title string
}

var targets = []target{
	{"archmage.go", "archmage.mdx", "Archmage"},
	{"atlas.go", "atlas.mdx", "ConfigAtlas"},
	{"atlas_extension.go", "atlas_extension.mdx", "AtlasExtension"},
	{"l10n.go", "l10n.mdx", "L10n"},
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("update: ")

	src := flag.String("src", "internal/conf", "directory of the generated config package")
	out := flag.String("out", "../docs/archmage/src/content/docs/gen-go", "output directory for the .mdx pages")
	overlays := flag.String("overlays", "docs/__impl/update/overlays", "directory of hand-written page preambles")
	flag.Parse()

	if err := run(*src, *out, *overlays); err != nil {
		log.Fatal(err)
	}
}

func run(src, out, overlays string) error {
	if fi, err := os.Stat(out); err != nil || !fi.IsDir() {
		return fmt.Errorf("output directory %s does not exist", out)
	}

	fset := token.NewFileSet()
	files, err := parseDir(fset, src)
	if err != nil {
		return err
	}

	// Collected before doc.NewFromFiles, which drops unexported declarations.
	impls := interfaceChecks(files)

	pkg, err := doc.NewFromFiles(fset, files, "conf")
	if err != nil {
		return err
	}

	r := &renderer{fset: fset, impls: impls, printer: pkg.Printer(), parser: pkg.Parser()}
	r.printer.DocLinkURL = func(*comment.DocLink) string { return "" }

	sections := group(fset, pkg)
	if err := cleanPages(out); err != nil {
		return err
	}

	for _, t := range targets {
		preamble, err := readOverlay(overlays, t.src)
		if err != nil {
			return err
		}
		page := r.render(t, sections[t.src], preamble)
		dst := filepath.Join(out, t.out)
		if err := os.WriteFile(dst, []byte(page), 0o644); err != nil {
			return err
		}
		fmt.Printf("%s -> %s\n", filepath.Join(src, t.src), dst)
	}
	return nil
}

func parseDir(fset *token.FileSet, dir string) ([]*ast.File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no Go source found in %s", dir)
	}
	return files, nil
}

// section holds the documented declarations of a single source file.
type section struct {
	vars  []*doc.Value
	funcs []*doc.Func
	types []*doc.Type
}

func group(fset *token.FileSet, pkg *doc.Package) map[string]*section {
	m := make(map[string]*section)
	at := func(n ast.Node) *section {
		name := filepath.Base(fset.Position(n.Pos()).Filename)
		s := m[name]
		if s == nil {
			s = &section{}
			m[name] = s
		}
		return s
	}

	for _, v := range pkg.Vars {
		s := at(v.Decl)
		s.vars = append(s.vars, v)
	}
	for _, fn := range pkg.Funcs {
		s := at(fn.Decl)
		s.funcs = append(s.funcs, fn)
	}
	for _, t := range pkg.Types {
		if isAlias(t) {
			continue
		}
		s := at(t.Decl)
		s.types = append(s.types, t)
	}
	return m
}

// isAlias reports whether t is declared with `=`. Aliases merely re-export SDK
// types under the config package, so they carry no documentation of their own.
func isAlias(t *doc.Type) bool {
	for _, spec := range t.Decl.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == t.Name {
			return ts.Assign.IsValid()
		}
	}
	return false
}

// interfaceChecks collects compliance declarations of the form
//
//	var _ archmage.Atlas = (*ConfigAtlas)(nil)
//
// keyed by the concrete type. go/doc drops them because the variable is
// unexported, yet they are the only place the implemented interface is stated.
func interfaceChecks(files []*ast.File) map[string][]string {
	m := make(map[string][]string)
	for _, f := range files {
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || vs.Type == nil || len(vs.Names) != 1 || len(vs.Values) != 1 {
					continue
				}
				if vs.Names[0].Name != "_" {
					continue
				}
				if name := concreteType(vs.Values[0]); name != "" {
					m[name] = append(m[name], types.ExprString(vs.Type))
				}
			}
		}
	}
	return m
}

// concreteType extracts T from `(*T)(nil)` and `T{}`.
func concreteType(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.CallExpr:
		if p, ok := v.Fun.(*ast.ParenExpr); ok {
			if s, ok := p.X.(*ast.StarExpr); ok {
				if id, ok := s.X.(*ast.Ident); ok {
					return id.Name
				}
			}
		}
	case *ast.CompositeLit:
		if id, ok := v.Type.(*ast.Ident); ok {
			return id.Name
		}
	}
	return ""
}

type renderer struct {
	fset    *token.FileSet
	impls   map[string][]string
	printer *comment.Printer
	parser  *comment.Parser
}

func (r *renderer) render(t target, s *section, preamble string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: '%s'\n---\n", t.title)
	if preamble != "" {
		b.WriteString("\n" + preamble)
	}
	if s == nil {
		return b.String()
	}

	if len(s.vars) > 0 {
		b.WriteString("\n## Variables\n")
		for _, v := range s.vars {
			r.writeDoc(&b, v.Doc)
			v.Decl.Doc = nil
			b.WriteString("\n" + r.codeBlock(r.node(v.Decl)))
		}
	}
	if len(s.funcs) > 0 {
		b.WriteString("\n## Functions\n")
		for _, fn := range s.funcs {
			r.writeFunc(&b, "###", fn)
		}
	}
	if len(s.types) > 0 {
		b.WriteString("\n## Types\n")
		for _, ty := range s.types {
			fmt.Fprintf(&b, "\n### %s\n", ty.Name)
			r.writeDoc(&b, ty.Doc)
			for _, iface := range r.impls[ty.Name] {
				fmt.Fprintf(&b, "\nImplements `%s`.\n", iface)
			}
			b.WriteString("\n" + r.codeBlock(r.typeDecl(ty)))
			for _, fn := range ty.Funcs {
				r.writeFunc(&b, "####", fn)
			}
			for _, fn := range ty.Methods {
				r.writeFunc(&b, "####", fn)
			}
		}
	}
	return b.String()
}

func (r *renderer) writeFunc(b *strings.Builder, heading string, fn *doc.Func) {
	fmt.Fprintf(b, "\n%s %s\n", heading, fn.Name)
	r.writeDoc(b, fn.Doc)

	sig := *fn.Decl
	sig.Doc = nil
	sig.Body = nil
	b.WriteString("\n" + r.codeBlock(r.node(&sig)))
}

func (r *renderer) writeDoc(b *strings.Builder, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	md := string(r.printer.Markdown(r.parser.Parse(text)))
	b.WriteString("\n" + strings.TrimRight(md, "\n") + "\n")
}

// typeDecl renders a type declaration, replacing the config table fields of the
// atlas struct with a placeholder comment. Those fields are named after the
// tables of one particular config repo, so they say nothing about the API.
func (r *renderer) typeDecl(ty *doc.Type) string {
	decl := ty.Decl
	decl.Doc = nil

	dropped := false
	for _, spec := range decl.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		ts.Doc = nil
		if st, ok := ts.Type.(*ast.StructType); ok && dropTableFields(st) {
			dropped = true
		}
	}

	code := r.node(decl)
	if dropped {
		code = strings.TrimRight(code, "\n")
		code = strings.TrimRight(strings.TrimSuffix(code, "}"), " \t\n")
		code += "\n\n    // One field per config type.\n    // ...\n}"
	}
	return code
}

// dropTableFields keeps embedded fields and fields whose type comes from
// another package, and reports whether anything else was removed.
func dropTableFields(st *ast.StructType) bool {
	// go/doc sets Incomplete after removing unexported fields, which makes the
	// printer emit "// Has unexported fields." — noise for a generated struct.
	st.Incomplete = false

	fields := st.Fields
	if fields == nil {
		return false
	}

	kept := fields.List[:0]
	dropped := false
	for _, f := range fields.List {
		if len(f.Names) == 0 || qualified(f.Type) {
			kept = append(kept, f)
			continue
		}
		dropped = true
	}
	fields.List = kept
	return dropped
}

// qualified reports whether the type expression names a type from another
// package, such as *archmage.VersionInfo.
func qualified(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if se, ok := n.(*ast.SelectorExpr); ok {
			if _, ok := se.X.(*ast.Ident); ok {
				found = true
			}
		}
		return !found
	})
	return found
}

func (r *renderer) node(n ast.Node) string {
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}
	if err := cfg.Fprint(&buf, r.fset, n); err != nil {
		log.Fatalf("printing %T: %v", n, err)
	}
	return buf.String()
}

func (r *renderer) codeBlock(code string) string {
	return "```go\n" + strings.TrimRight(code, "\n") + "\n```\n"
}

func readOverlay(dir, srcFile string) (string, error) {
	name := strings.TrimSuffix(srcFile, ".go") + ".md"
	b, err := os.ReadFile(filepath.Join(dir, name))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(b), "\n") + "\n", nil
}

func cleanPages(dir string) error {
	matches, err := filepath.Glob(filepath.Join(dir, "*.mdx"))
	if err != nil {
		return err
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			return err
		}
	}
	return nil
}

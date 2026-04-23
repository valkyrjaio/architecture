// Package analysis walks a Go implementation of ComponentProviderContract and
// extracts the types returned from each Get*Providers method, resolving each
// to the source file that declares it.
//
// The Go port is assumed to use an interface like:
//
//	type ComponentProvider interface {
//	    GetComponentProviders(app Application) []ComponentProvider
//	    GetContainerProviders(app Application) []ServiceProvider
//	    // ...
//	}
//
// and each implementation returns a literal slice of typed nils or zero values
// (a common Go pattern for type lists), e.g.
//
//	return []ComponentProvider{(*foo.Provider)(nil), (*bar.Provider)(nil)}
//
// We extract those type references and resolve each to a file via the
// packages loader.
package analysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Methods we care about, keyed by the output bucket name.
var methods = map[string]string{
	"componentProviders": "GetComponentProviders",
	"containerProviders": "GetContainerProviders",
	"eventProviders":     "GetEventProviders",
	"cliProviders":       "GetCliProviders",
	"httpProviders":      "GetHttpProviders",
}

// ProviderRef is a single resolved (or unresolved) provider reference.
type ProviderRef struct {
	FQN  string // e.g. "github.com/you/mod/foo.Provider"
	File string // absolute path, or "" if unresolved
}

// Result is the analysis output for one provider file.
type Result struct {
	TypeName  string                   // FQN of the provider type in this file
	Providers map[string][]ProviderRef // keyed by methods[*]
}

// Walker caches a loaded package set so multiple analyze() calls share work.
type Walker struct {
	pkgs []*packages.Package
	fset *token.FileSet
}

// NewWalker loads the given patterns (e.g. "./...") with full type info. The
// loader handles modules, vendor, and GOPATH — we don't reimplement any of it.
func NewWalker(patterns ...string) (*Walker, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports |
			packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package load had errors")
	}
	return &Walker{pkgs: pkgs, fset: cfg.Fset}, nil
}

// Analyze finds the provider declaration in the given file and extracts the
// class references returned from each Get*Providers method.
func (w *Walker) Analyze(filePath string) (*Result, error) {
	pkg, file := w.findFile(filePath)
	if file == nil {
		return nil, fmt.Errorf("file not found in loaded packages: %s", filePath)
	}

	result := &Result{Providers: make(map[string][]ProviderRef, len(methods))}
	for key := range methods {
		result.Providers[key] = nil
	}

	// Locate the top-level type declaration in this file — we'll use it for
	// the TypeName field and to restrict method lookup to this file's type.
	typeName, methodsByName := collectMethods(file)
	if typeName != "" {
		result.TypeName = pkg.PkgPath + "." + typeName
	}

	for key, methodName := range methods {
		fn, ok := methodsByName[methodName]
		if !ok {
			continue
		}
		refs := w.extractTypeRefs(pkg, fn)
		result.Providers[key] = refs
	}

	return result, nil
}

// findFile locates the *ast.File for a given path across loaded packages.
func (w *Walker) findFile(filePath string) (*packages.Package, *ast.File) {
	for _, pkg := range w.pkgs {
		for i, f := range pkg.CompiledGoFiles {
			if f == filePath {
				return pkg, pkg.Syntax[i]
			}
		}
	}
	return nil, nil
}

// collectMethods walks the file and returns the first concrete type name
// declared in it plus a map of methods defined on that type.
func collectMethods(file *ast.File) (string, map[string]*ast.FuncDecl) {
	methodsByName := map[string]*ast.FuncDecl{}
	var typeName string

	// First pass: find the first non-interface type declaration.
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts := spec.(*ast.TypeSpec)
			if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
				continue
			}
			typeName = ts.Name.Name
			break
		}
		if typeName != "" {
			break
		}
	}

	// Second pass: gather methods on that type.
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		recv := receiverTypeName(fn.Recv.List[0].Type)
		if typeName == "" || recv == typeName {
			methodsByName[fn.Name.Name] = fn
		}
	}
	return typeName, methodsByName
}

func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	}
	return ""
}

// extractTypeRefs walks the method body looking for composite literals inside
// return statements and pulls out the type of each element.
func (w *Walker) extractTypeRefs(pkg *packages.Package, fn *ast.FuncDecl) []ProviderRef {
	var refs []ProviderRef
	if fn.Body == nil {
		return refs
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, expr := range ret.Results {
			cl, ok := expr.(*ast.CompositeLit)
			if !ok {
				continue
			}
			for _, elt := range cl.Elts {
				if ref := w.refFromExpr(pkg, elt); ref != nil {
					refs = append(refs, *ref)
				}
			}
		}
		return true
	})
	return refs
}

// refFromExpr handles the common "typed nil" pattern used to build type lists
// in Go: (*pkg.Type)(nil). Also handles bare identifiers and zero-value
// composite literals like pkg.Type{}.
func (w *Walker) refFromExpr(pkg *packages.Package, expr ast.Expr) *ProviderRef {
	// (*pkg.Type)(nil) parses as a CallExpr whose Fun is a ParenExpr wrapping
	// a StarExpr. Peel off those layers to find the underlying type expr.
	if call, ok := expr.(*ast.CallExpr); ok {
		if paren, ok := call.Fun.(*ast.ParenExpr); ok {
			if star, ok := paren.X.(*ast.StarExpr); ok {
				return w.refFromTypeExpr(pkg, star.X)
			}
		}
	}
	// pkg.Type{} -> composite literal
	if cl, ok := expr.(*ast.CompositeLit); ok && cl.Type != nil {
		return w.refFromTypeExpr(pkg, cl.Type)
	}
	// Bare identifier (same-package type)
	return w.refFromTypeExpr(pkg, expr)
}

func (w *Walker) refFromTypeExpr(pkg *packages.Package, expr ast.Expr) *ProviderRef {
	tv, ok := pkg.TypesInfo.Types[expr]
	if !ok || tv.Type == nil {
		return nil
	}
	named, ok := tv.Type.(*types.Named)
	if !ok {
		// Might be a pointer — unwrap once.
		if ptr, ok := tv.Type.(*types.Pointer); ok {
			named, ok = ptr.Elem().(*types.Named)
			if !ok {
				return nil
			}
		} else {
			return nil
		}
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return nil
	}
	fqn := obj.Pkg().Path() + "." + obj.Name()
	file := w.fset.Position(obj.Pos()).Filename
	return &ProviderRef{FQN: fqn, File: file}
}

// Consume `strings` to keep the import; useful if you extend type-name
// formatting (e.g. trimming vendor prefixes).
var _ = strings.TrimPrefix

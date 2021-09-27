package tracegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
)

const (
	tracingAnnotation = "@trace"
	unknownType       = "Unknown"
)

// Run generates OpenTelemetry tracing decorators and writes result to the
// specified io.Writer.
func Run(source string, w io.Writer) error {
	g := generator{
		source:  source,
		imports: make(map[string]string),
	}

	raw, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}
	g.raw = string(raw)

	fset := token.NewFileSet()
	g.fin, err = parser.ParseFile(fset, source, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse source file: %w", err)
	}

	g.run()

	err = g.fout.Render(w)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	return nil
}

type generator struct {
	source string
	raw    string
	fin    *ast.File
	fout   *jen.File

	baseName    string
	wrapperName string
	imports     map[string]string // maps a package name to a fully qualified path
}

func (g *generator) run() {
	ast.Inspect(g.fin, func(node ast.Node) bool {
		switch t := node.(type) {
		case *ast.File:
			g.fout = jen.NewFile(t.Name.Name)
			g.fout.PackageComment("// Code generated by tracegen. DO NOT EDIT.")
			g.fout.PackageComment("// Source: " + g.source)
			g.handleImports(t.Imports)
		case *ast.TypeSpec:
			if t.Name.IsExported() {
				if it, ok := t.Type.(*ast.InterfaceType); ok {
					g.handleInterface(t, it)
				}
			}
		}

		return true
	})
}

func (g generator) handleImports(specs []*ast.ImportSpec) {
	for _, spec := range specs {
		g.handleImport(spec)
	}
}

func (g *generator) handleImport(spec *ast.ImportSpec) {
	path := strings.Trim(spec.Path.Value, `"`)
	name := path

	if i := strings.LastIndex(path, "/"); i != -1 {
		name = path[i+1:]
	}

	if spec.Name != nil {
		name = spec.Name.Name
	}

	g.imports[name] = path
	g.fout.ImportAlias(path, name)
}

func (g *generator) handleInterface(t *ast.TypeSpec, it *ast.InterfaceType) {
	if len(it.Methods.List) == 0 {
		return
	}

	g.baseName = t.Name.Name
	g.wrapperName = "Traced" + g.baseName
	g.fout.Type().Id(g.wrapperName).Struct(
		jen.Id("base").Id(t.Name.Name),
		jen.Id("tracer").Qual("go.opentelemetry.io/otel/trace", "Tracer"),
	)

	g.fout.Func().
		Id("New"+g.wrapperName).
		Params(
			jen.Id("_base").Id(t.Name.Name),
			jen.Id("_tracer").Qual("go.opentelemetry.io/otel/trace", "Tracer"),
		).
		Params(jen.Op("*").Id(g.wrapperName)).
		Block(
			jen.Return(jen.Op("&").Id(g.wrapperName).Values(jen.Dict{
				jen.Id("base"):   jen.Id("_base"),
				jen.Id("tracer"): jen.Id("_tracer"),
			})),
		)

	for _, f := range it.Methods.List {
		if ft, ok := f.Type.(*ast.FuncType); ok {
			g.handleMethod(f, ft)
		}
	}
}

func (g *generator) handleMethod(f *ast.Field, ft *ast.FuncType) {
	g.fout.Func().
		Params(jen.Id("_d").Op("*").Id(g.wrapperName)).
		Id(f.Names[0].Name).
		ParamsFunc(func(group *jen.Group) {
			for _, param := range ft.Params.List {
				paramName := param.Names[0].Name

				group.Id(paramName).Do(g.addType(param))
			}
		}).
		ParamsFunc(func(group *jen.Group) {
			if ft.Results == nil {
				return
			}

			for _, result := range ft.Results.List {
				group.Do(g.addType(result))
			}
		}).
		BlockFunc(func(group *jen.Group) {
			g.generateMethodBody(group, f, ft)
		})
}

func (g *generator) addType(param *ast.Field) func(s *jen.Statement) {
	return func(s *jen.Statement) {
		paramType := g.formatFieldType(param)

		if strings.HasPrefix(paramType, "...") {
			paramType = paramType[3:]
			s.Op("...")
		}

	loop:
		for {
			switch {
			case strings.HasPrefix(paramType, "*"):
				paramType = paramType[1:]
				s.Op("*")
			case strings.HasPrefix(paramType, "[]"):
				paramType = paramType[2:]
				s.Op("[]")
			default:
				break loop
			}
		}

		i := strings.Index(paramType, ".")
		if i == -1 {
			s.Id(paramType)
			return
		}

		s.Qual(g.imports[paramType[:i]], paramType[i+1:])
	}
}

func (g *generator) generateMethodBody(
	group *jen.Group,
	f *ast.Field,
	ft *ast.FuncType,
) {
	methodName := f.Names[0].Name

	group.Var().Id("_span").Qual("go.opentelemetry.io/otel/trace", "Span")
	group.List(
		jen.Id("ctx"), jen.Id("_span")).
		Op("=").
		Id("_d").Dot("tracer").Dot("Start").Call(jen.Id("ctx"), jen.Lit(g.baseName+"."+methodName+"()"))
	group.Defer().Id("_span").Dot("End").Call()

	if f.Doc != nil {
		g.generateTraceAttributes(f, group)
	}

	params := make([]jen.Code, len(ft.Params.List))
	for i, param := range ft.Params.List {
		if _, ok := param.Type.(*ast.Ellipsis); ok {
			params[i] = jen.Id(param.Names[0].Name + "...")
		} else {
			params[i] = jen.Id(param.Names[0].Name)
		}
	}

	if ft.Results == nil {
		group.Id("_d").Dot("base").Dot(methodName).Call(params...)
		return
	}

	vars := make([]jen.Code, len(ft.Results.List))
	for i := range ft.Results.List {
		vars[i] = jen.Id("_var" + strconv.Itoa(i))
	}

	group.List(vars...).Op(":=").Id("_d").Dot("base").Dot(methodName).Call(params...)

	for i, result := range ft.Results.List {
		if g.formatFieldType(result) == "error" {
			v := "_var" + strconv.Itoa(i)
			group.If(jen.Id(v).Op("!=").Nil()).BlockFunc(func(group *jen.Group) {
				group.Id("_span").Dot("RecordError").Call(jen.Id(v))
				group.Id("_span").Dot("SetStatus").Call(
					jen.Qual("go.opentelemetry.io/otel/codes", "Error"),
					jen.Id(v).Dot("Error").Call())
			})
		}
	}

	group.Return(vars...)
}

func (g *generator) generateTraceAttributes(f *ast.Field, group *jen.Group) {
	for _, comment := range f.Doc.List {
		if idx := strings.Index(comment.Text, tracingAnnotation); idx != -1 {
			annotations := strings.Split(comment.Text[idx+len(tracingAnnotation):], ",")
			for _, annotation := range annotations {
				g.handleAnnotation(group, annotation)
			}
		}
	}
}

func (g *generator) handleAnnotation(group *jen.Group, annotation string) {
	typ, param := g.parseAnnotation(annotation)

	if typ == unknownType {
		group.If(
			jen.List(jen.Id("data"), jen.Id("err")).Op(":=").Qual("encoding/json", "Marshal").Call(jen.Id(param)),
			jen.Id("err").Op("==").Nil(),
		).BlockFunc(func(group *jen.Group) {
			group.Id("_span").Dot("SetAttributes").Call(
				jen.Qual("go.opentelemetry.io/otel/attribute", "String").Call(jen.Lit(param), jen.String().Parens(jen.Id("data"))),
			)
		}).Else().BlockFunc(func(group *jen.Group) {
			group.Id("_span").Dot("SetAttributes").Call(
				jen.Qual("go.opentelemetry.io/otel/attribute", "String").Call(jen.Lit(param), jen.Qual("fmt", "Sprint").Call(jen.Id(param))),
			)
		})
	} else {
		group.Id("_span").Dot("SetAttributes").Call(
			jen.Qual("go.opentelemetry.io/otel/attribute", typ).Call(jen.Lit(param), jen.Id(param)),
		)
	}
}

func (g *generator) parseAnnotation(annotation string) (typ, param string) {
	typ = unknownType
	param = strings.TrimSpace(annotation)
	if i := strings.Index(param, ":"); i != -1 {
		typ = param[:i]
		param = param[i+1:]
	}
	return typ, param
}

func (g *generator) formatFieldType(f *ast.Field) string {
	return g.raw[f.Type.Pos()-1 : f.Type.End()-1]
}

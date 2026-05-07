//go:build ignore

// gen_skill_required_fields.go reads all *Request structs in
// internal/features/deploy/types/init.go, extracts fields with
// validate:"required", and regenerates the required-fields block
// inside skills/api-catalog/SKILL.md.
//
// Run:  go run api/scripts/gen_skill_required_fields.go
// Or:   make gen-skill
//
// The generated block is bracketed by:
//
//	<!-- BEGIN REQUIRED FIELDS (auto-generated) -->
//	<!-- END REQUIRED FIELDS (auto-generated) -->
//
// so it is safe to run repeatedly.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const (
	beginMarker = "<!-- BEGIN REQUIRED FIELDS (auto-generated) -->"
	endMarker   = "<!-- END REQUIRED FIELDS (auto-generated) -->"
)

func main() {
	_, file, _, _ := runtime.Caller(0)
	// file = .../api/scripts/gen_skill_required_fields.go
	apiRoot := filepath.Join(filepath.Dir(file), "..")

	typesFile := filepath.Join(apiRoot, "internal", "features", "deploy", "types", "init.go")
	skillFile := filepath.Join(apiRoot, "skills", "api-catalog", "SKILL.md")

	structs, err := parseRequiredFields(typesFile)
	if err != nil {
		fatalf("parse types: %v", err)
	}

	block := buildBlock(structs)

	if err := patchSkill(skillFile, block); err != nil {
		fatalf("patch skill: %v", err)
	}

	fmt.Printf("✓ Updated required fields in %s (%d structs)\n", skillFile, len(structs))
}

// structFields holds a request struct name and its required fields with metadata.
type structField struct {
	Name        string // JSON field name
	Type        string // Go type (string, int, etc.)
	Description string // from `description:` tag
	Rules       string // extra hints (e.g. for repository)
}

type requestStruct struct {
	Name   string
	Fields []structField
}

// parseRequiredFields uses go/ast to extract every *Request struct's required fields.
func parseRequiredFields(path string) ([]requestStruct, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	var result []requestStruct
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			name := typeSpec.Name.Name
			if !strings.HasSuffix(name, "Request") {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			var fields []structField
			for _, field := range structType.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
				validateTag := tag.Get("validate")
				if !strings.Contains(validateTag, "required") {
					continue
				}

				jsonName := strings.SplitN(tag.Get("json"), ",", 2)[0]
				if jsonName == "" || jsonName == "-" {
					if len(field.Names) > 0 {
						jsonName = field.Names[0].Name
					}
				}

				typeName := typeString(field.Type)
				desc := tag.Get("description")
				rules := ""

				// Add known special-case hints
				if jsonName == "repository" {
					rules = "MUST be numeric GitHub repo ID — call GET /api/v1/github-connector/repositories and use the integer `id`. NEVER pass owner/repo slug or a URL."
				}
				if jsonName == "environment" {
					rules = `Valid values: "production", "staging", "development"`
				}
				if jsonName == "build_pack" {
					rules = `Valid values: "dockerfile", "nixpacks", "static"`
				}

				fields = append(fields, structField{
					Name:        jsonName,
					Type:        typeName,
					Description: desc,
					Rules:       rules,
				})
			}

			if len(fields) > 0 {
				result = append(result, requestStruct{Name: name, Fields: fields})
			}
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", t.X, t.Sel.Name)
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", typeString(t.Key), typeString(t.Value))
	default:
		return "any"
	}
}

func buildBlock(structs []requestStruct) string {
	var b bytes.Buffer
	b.WriteString(beginMarker + "\n")
	b.WriteString("## REQUIRED FIELDS — auto-generated from Go struct tags (do not edit manually)\n")
	b.WriteString("<!-- Run `go run api/scripts/gen_skill_required_fields.go` to regenerate -->\n\n")
	b.WriteString("These fields have `validate:\"required\"` and WILL cause HTTP 400 if omitted.\n\n")

	for _, s := range structs {
		b.WriteString(fmt.Sprintf("### %s\n", s.Name))
		for _, f := range s.Fields {
			line := fmt.Sprintf("- `%s` (%s)", f.Name, f.Type)
			if f.Description != "" {
				line += ": " + f.Description
			}
			if f.Rules != "" {
				line += " **→ " + f.Rules + "**"
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(endMarker + "\n")
	return b.String()
}

func patchSkill(skillPath, block string) error {
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return err
	}
	content := string(data)

	re := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(beginMarker) + `.*?` + regexp.QuoteMeta(endMarker) + `\n?`)
	if re.MatchString(content) {
		content = re.ReplaceAllString(content, block)
	} else {
		content += "\n" + block
	}

	return os.WriteFile(skillPath, []byte(content), 0o644)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

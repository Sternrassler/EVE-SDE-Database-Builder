// tools/add-tomap-methods.go
// Post-Processing Tool: Adds ToMap() methods to generated structs
// ADR Reference: ADR-003 (Custom Post-Processing)
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

const (
	generatedMarker = "// Code generated"
	toMapMethodName = "ToMap"
)

// Config holds the configuration for the tool
type Config struct {
	InputFiles  []string
	DryRun      bool
	Verbose     bool
	ForceUpdate bool
}

// FieldInfo contains parsed information about a struct field
type FieldInfo struct {
	Name     string
	JSONTag  string
	Type     string
	IsPtr    bool
	IsMap    bool
	IsSlice  bool
	IsStruct bool
}

func main() {
	cfg := parseFlags()

	if len(cfg.InputFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		fmt.Fprintln(os.Stderr, "Usage: add-tomap-methods [options] <file1.go> [file2.go ...]")
		os.Exit(1)
	}

	successCount := 0
	errorCount := 0

	for _, inputFile := range cfg.InputFiles {
		if cfg.Verbose {
			fmt.Printf("Processing: %s\n", inputFile)
		}

		err := processFile(inputFile, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", inputFile, err)
			errorCount++
			continue
		}

		successCount++
	}

	fmt.Printf("Processed %d file(s) successfully, %d error(s)\n", successCount, errorCount)

	if errorCount > 0 {
		os.Exit(1)
	}
}

func parseFlags() *Config {
	cfg := &Config{}

	dryRun := flag.Bool("dry-run", false, "Print output to stdout instead of modifying files")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	force := flag.Bool("force", false, "Force update even if ToMap methods already exist")

	flag.Parse()

	cfg.DryRun = *dryRun
	cfg.Verbose = *verbose
	cfg.ForceUpdate = *force
	cfg.InputFiles = flag.Args()

	return cfg
}

func processFile(filename string, cfg *Config) error {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check if file is generated code (safety check)
	if !cfg.ForceUpdate && !strings.Contains(string(content), generatedMarker) {
		if cfg.Verbose {
			fmt.Printf("Skipping %s: not generated code (missing '%s' marker)\n", filename, generatedMarker)
		}
		return nil
	}

	// Parse the Go source file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Find all struct declarations and check if ToMap already exists
	structs := extractStructs(file)
	if len(structs) == 0 {
		if cfg.Verbose {
			fmt.Printf("Skipping %s: no structs found\n", filename)
		}
		return nil
	}

	// Check if ToMap methods already exist
	existingMethods := extractExistingToMapMethods(file)
	if !cfg.ForceUpdate && len(existingMethods) > 0 {
		if cfg.Verbose {
			fmt.Printf("Skipping %s: ToMap methods already exist for %d struct(s)\n", filename, len(existingMethods))
		}
		return nil
	}

	// Generate ToMap methods for each struct
	var buf bytes.Buffer
	buf.Write(content)

	for structName, fields := range structs {
		// Skip if method already exists (unless force update)
		if !cfg.ForceUpdate && existingMethods[structName] {
			continue
		}

		method := generateToMapMethod(structName, fields)
		buf.WriteString("\n\n")
		buf.WriteString(method)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated code: %w", err)
	}

	// Write output
	if cfg.DryRun {
		fmt.Println(string(formatted))
	} else {
		err = os.WriteFile(filename, formatted, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		if cfg.Verbose {
			fmt.Printf("Added ToMap methods to %s (%d struct(s))\n", filename, len(structs))
		}
	}

	return nil
}

// extractStructs finds all struct type declarations in the file
func extractStructs(file *ast.File) map[string][]FieldInfo {
	structs := make(map[string][]FieldInfo)

	ast.Inspect(file, func(n ast.Node) bool {
		// Look for type declarations
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// Check if it's a struct type
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structName := typeSpec.Name.Name
		fields := extractFields(structType)
		structs[structName] = fields

		return true
	})

	return structs
}

// extractFields extracts field information from a struct
func extractFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo

	for _, field := range structType.Fields.List {
		// Skip embedded fields
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name

		// Skip unexported fields
		if !ast.IsExported(fieldName) {
			continue
		}

		// Extract JSON tag
		jsonTag := ""
		if field.Tag != nil {
			tag := field.Tag.Value
			jsonTag = extractJSONTag(tag)
		}

		// Skip fields with json:"-" tag
		if jsonTag == "-" {
			continue
		}

		// Analyze field type
		typeInfo := analyzeType(field.Type)

		fields = append(fields, FieldInfo{
			Name:     fieldName,
			JSONTag:  jsonTag,
			Type:     typeInfo.Name,
			IsPtr:    typeInfo.IsPtr,
			IsMap:    typeInfo.IsMap,
			IsSlice:  typeInfo.IsSlice,
			IsStruct: typeInfo.IsStruct,
		})
	}

	return fields
}

// TypeInfo holds analyzed type information
type TypeInfo struct {
	Name     string
	IsPtr    bool
	IsMap    bool
	IsSlice  bool
	IsStruct bool
}

// analyzeType analyzes a field type to determine its characteristics
func analyzeType(expr ast.Expr) TypeInfo {
	info := TypeInfo{}

	switch t := expr.(type) {
	case *ast.StarExpr:
		// Pointer type
		info.IsPtr = true
		innerInfo := analyzeType(t.X)
		info.Name = "*" + innerInfo.Name
		info.IsMap = innerInfo.IsMap
		info.IsSlice = innerInfo.IsSlice
		info.IsStruct = innerInfo.IsStruct

	case *ast.MapType:
		// Map type
		info.IsMap = true
		info.Name = "map"

	case *ast.ArrayType:
		// Slice/array type
		info.IsSlice = true
		info.Name = "slice"

	case *ast.Ident:
		// Simple identifier (int, string, bool, or custom type)
		info.Name = t.Name
		// Check if it's likely a struct (capitalized name, not a builtin)
		if ast.IsExported(t.Name) && !isBuiltinType(t.Name) {
			info.IsStruct = true
		}

	case *ast.SelectorExpr:
		// Qualified identifier (e.g., time.Time, pkg.Type)
		info.Name = fmt.Sprintf("%v.%v", t.X, t.Sel)
		info.IsStruct = true

	default:
		info.Name = "interface{}"
	}

	return info
}

// isBuiltinType checks if a type name is a Go builtin type
func isBuiltinType(name string) bool {
	builtins := map[string]bool{
		"bool":       true,
		"string":     true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uintptr":    true,
		"byte":       true,
		"rune":       true,
		"float32":    true,
		"float64":    true,
		"complex64":  true,
		"complex128": true,
	}
	return builtins[name]
}

// extractJSONTag extracts the JSON tag value from a struct tag
func extractJSONTag(tag string) string {
	// Remove backticks
	tag = strings.Trim(tag, "`")

	// Find json:"..." part
	for _, part := range strings.Split(tag, " ") {
		if strings.HasPrefix(part, "json:") {
			value := strings.TrimPrefix(part, "json:")
			value = strings.Trim(value, `"`)

			// Extract field name (before comma)
			if idx := strings.Index(value, ","); idx != -1 {
				return value[:idx]
			}
			return value
		}
	}

	return ""
}

// extractExistingToMapMethods finds existing ToMap methods
func extractExistingToMapMethods(file *ast.File) map[string]bool {
	methods := make(map[string]bool)

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Check if it's a method named ToMap
		if funcDecl.Name.Name != toMapMethodName {
			return true
		}

		// Check if it has a receiver
		if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			return true
		}

		// Extract receiver type name
		recv := funcDecl.Recv.List[0]
		var recvTypeName string

		switch t := recv.Type.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				recvTypeName = ident.Name
			}
		case *ast.Ident:
			recvTypeName = t.Name
		}

		if recvTypeName != "" {
			methods[recvTypeName] = true
		}

		return true
	})

	return methods
}

// generateToMapMethod generates a ToMap() method for a struct
func generateToMapMethod(structName string, fields []FieldInfo) string {
	var buf bytes.Buffer

	// Generate method signature
	fmt.Fprintf(&buf, "// ToMap converts %s to map[string]interface{} for database operations\n", structName)
	fmt.Fprintf(&buf, "func (t *%s) ToMap() map[string]interface{} {\n", structName)
	fmt.Fprintf(&buf, "\tm := make(map[string]interface{})\n\n")

	// Generate map entries for each field
	for _, field := range fields {
		mapKey := field.JSONTag
		if mapKey == "" {
			mapKey = field.Name
		}

		fieldAccess := fmt.Sprintf("t.%s", field.Name)

		// Handle different field types
		if field.IsPtr {
			// Pointer field - check for nil
			fmt.Fprintf(&buf, "\tif %s != nil {\n", fieldAccess)
			fmt.Fprintf(&buf, "\t\tm[\"%s\"] = *%s\n", mapKey, fieldAccess)
			fmt.Fprintf(&buf, "\t}\n")
		} else if field.IsStruct {
			// Nested struct - check if it has ToMap method, otherwise use reflection
			fmt.Fprintf(&buf, "\t// Nested struct field: %s\n", field.Name)
			fmt.Fprintf(&buf, "\tif toMapper, ok := interface{}(%s).(interface{ ToMap() map[string]interface{} }); ok {\n", fieldAccess)
			fmt.Fprintf(&buf, "\t\tm[\"%s\"] = toMapper.ToMap()\n", mapKey)
			fmt.Fprintf(&buf, "\t} else {\n")
			fmt.Fprintf(&buf, "\t\tm[\"%s\"] = %s\n", mapKey, fieldAccess)
			fmt.Fprintf(&buf, "\t}\n")
		} else {
			// Simple field
			fmt.Fprintf(&buf, "\tm[\"%s\"] = %s\n", mapKey, fieldAccess)
		}
	}

	fmt.Fprintf(&buf, "\n\treturn m\n")
	fmt.Fprintf(&buf, "}")

	return buf.String()
}

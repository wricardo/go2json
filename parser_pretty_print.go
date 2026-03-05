package codesurgeon

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Knetic/govaluate"
)

// omitNullsFromJSON recursively removes null and empty values from JSON data
func omitNullsFromJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			cleaned := omitNullsFromJSON(val)
			// Only include non-nil values
			if cleaned != nil {
				result[key] = cleaned
			}
		}
		return result
	case []interface{}:
		if len(v) == 0 {
			return nil // Omit empty slices
		}
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			cleaned := omitNullsFromJSON(item)
			if cleaned != nil {
				result = append(result, cleaned)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case nil:
		return nil
	default:
		return v
	}
}

func PrettyPrint(
	parsed []*ParsedInfo,
	mode string,
	ignoreRules []string,
	plainStructs, fieldsPlainStructs, structsWithMethod, fieldsStructsWithMethod, methods, functions, tags bool,
	comments bool,
	omitNulls bool,
) string {
	var sb strings.Builder

	switch mode {
	case "json", "":
		return prettyPrintJSON(parsed, omitNulls)
	case "grepindex":
		return prettyPrintGrepIndex(parsed, ignoreRules, &sb)
	case "llm":
		return prettyPrintLLM(parsed, ignoreRules, plainStructs, fieldsPlainStructs, structsWithMethod, fieldsStructsWithMethod, methods, functions, comments, &sb)
	default:
		return "Invalid mode: " + mode
	}
}

func prettyPrintJSON(parsed []*ParsedInfo, omitNulls bool) string {
	var data interface{} = parsed

	if omitNulls {
		// Marshal to JSON then unmarshal to get a map structure we can filter
		jsonBytes, err := json.MarshalIndent(parsed, "", "  ")
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}

		var jsonData interface{}
		if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
			return fmt.Sprintf("Error: %v", err)
		}

		// Filter out nulls and empty values
		data = omitNullsFromJSON(jsonData)
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(pretty)
}

func prettyPrintGrepIndex(parsed []*ParsedInfo, ignoreRules []string, sb *strings.Builder) string {
	for _, p := range parsed {
		if len(p.Packages) == 0 {
			continue
		}
		posType, posValue := getPosition(p)
		if shouldIgnorePosition(p, ignoreRules, posType) || len(p.Packages) == 0 {
			continue
		}

		pos_ := formatAndPrint(sb, fmt.Sprintf("%s> %s\n", posType, posValue))

		for _, pkg := range p.Packages {
			pkg_ := formatAndPrint(sb, fmt.Sprintf("package> %s %s", pkg.Package, pos_))
			processStructsForGrepIndex(pkg, pkg_, sb)
			processFunctionsForGrepIndex(pkg, pkg_, sb)
			processVariablesForGrepIndex(pkg, posType, posValue, sb)
			processConstantsForGrepIndex(pkg, pkg_, sb)
			processInterfacesForGrepIndex(pkg, pkg_, sb)
		}
	}
	return sb.String()
}

func prettyPrintLLM(
	parsed []*ParsedInfo,
	ignoreRules []string,
	plainStructs, fieldsPlainStructs, structsWithMethod, fieldsStructsWithMethod, methods, functions, comments bool,
	sb *strings.Builder,
) string {
	for _, p := range parsed {
		if len(p.Packages) == 0 {
			continue
		}
		for _, pkg := range p.Packages {
			if p.File != "" {
				fmt.Fprintf(sb, "// file: %s\n", p.File)
			} else if p.Directory != "" {
				fmt.Fprintf(sb, "// directory: %s\n", p.Directory)
			}
			if pkg.Package != "" {
				fmt.Fprintf(sb, "package %s\n", pkg.Package)
			}

			for _, s := range pkg.Structs {
				if shouldIgnoreStruct(s, pkg.Package, ignoreRules) || !shouldIncludeStruct(s, plainStructs, structsWithMethod) {
					continue
				}
				if comments && len(s.Docs) > 0 {
					for _, d := range s.Docs {
						fmt.Fprintf(sb, "// %s\n", d)
					}
				}
				hasFields := (fieldsStructsWithMethod && len(s.Methods) > 0) || (fieldsPlainStructs && len(s.Methods) == 0)
				hasMethods := methods && len(s.Methods) > 0
				filteredFields := filterFields(s.Fields, ignoreRules)
				filteredMethods := filterMethods(s.Methods, ignoreRules)
				if !hasFields {
					filteredFields = nil
				}
				if !hasMethods {
					filteredMethods = nil
				}
				if len(filteredFields) == 0 && len(filteredMethods) == 0 {
					fmt.Fprintf(sb, "type %s struct{}\n", s.Name)
					continue
				}
				fmt.Fprintf(sb, "type %s struct{\n", s.Name)
				printFieldsToon(filteredFields, sb)
				printMethodsToon(filteredMethods, sb)
				sb.WriteString("}\n")
			}

			for _, iface := range pkg.Interfaces {
				if comments && len(iface.Docs) > 0 {
					for _, d := range iface.Docs {
						fmt.Fprintf(sb, "// %s\n", d)
					}
				}
				if len(iface.Methods) == 0 {
					fmt.Fprintf(sb, "type %s interface{}\n", iface.Name)
				} else {
					fmt.Fprintf(sb, "type %s interface{\n", iface.Name)
					for _, m := range iface.Methods {
						fmt.Fprintf(sb, "%s\n", m.Signature)
					}
					sb.WriteString("}\n")
				}
			}

			for _, td := range pkg.TypeDefs {
				if comments && len(td.Docs) > 0 {
					for _, d := range td.Docs {
						fmt.Fprintf(sb, "// %s\n", d)
					}
				}
				fmt.Fprintf(sb, "%s\n", td.Definition)
			}

			if functions {
				for _, f := range pkg.Functions {
					if shouldIgnoreFunction(f, ignoreRules) {
						continue
					}
					if comments && len(f.Docs) > 0 {
						for _, d := range f.Docs {
							fmt.Fprintf(sb, "// %s\n", d)
						}
					}
					fmt.Fprintf(sb, "%s\n", f.Definition)
				}
			}

			printVarsToon(pkg.Variables, comments, sb)
			printConstsToon(pkg.Constants, comments, sb)
		}
	}
	return sb.String()
}

// filterFields returns fields that pass ignore rules.
func filterFields(fields []Field, ignoreRules []string) []Field {
	var result []Field
	for _, f := range fields {
		if !shouldIgnoreField(f, ignoreRules) {
			result = append(result, f)
		}
	}
	return result
}

// filterMethods returns methods that pass ignore rules.
func filterMethods(methods []Method, ignoreRules []string) []Method {
	var result []Method
	for _, m := range methods {
		if !shouldIgnoreMethod(m, ignoreRules) {
			result = append(result, m)
		}
	}
	return result
}

// fieldGroup represents a group of consecutive fields sharing the same type.
type fieldGroup struct {
	names []string
	typ   string
}

// printFieldsToon prints struct fields in Go syntax with consecutive same-type grouping and alignment.
func printFieldsToon(fields []Field, sb *strings.Builder) {
	if len(fields) == 0 {
		return
	}

	// Group consecutive fields with the same type
	var groups []fieldGroup
	current := fieldGroup{names: []string{fields[0].Name}, typ: fields[0].Type}
	for i := 1; i < len(fields); i++ {
		if fields[i].Type == current.typ {
			current.names = append(current.names, fields[i].Name)
		} else {
			groups = append(groups, current)
			current = fieldGroup{names: []string{fields[i].Name}, typ: fields[i].Type}
		}
	}
	groups = append(groups, current)

	for _, g := range groups {
		fmt.Fprintf(sb, "%s %s\n", strings.Join(g.names, ","), g.typ)
	}
}

// printMethodsToon prints methods inside a struct block using toon format.
func printMethodsToon(methods []Method, sb *strings.Builder) {
	for _, m := range methods {
		prefix := ""
		if strings.HasPrefix(m.Receiver, "*") {
			prefix = "*"
		}
		fmt.Fprintf(sb, "%s%s\n", prefix, m.Signature)
	}
}

func printVarsToon(vars []Variable, comments bool, sb *strings.Builder) {
	// Group consecutive variables with same type
	type varGroup struct {
		names []string
		typ   string
		docs  [][]string
	}
	var groups []varGroup
	for _, v := range vars {
		if len(groups) > 0 && groups[len(groups)-1].typ == v.Type {
			groups[len(groups)-1].names = append(groups[len(groups)-1].names, v.Name)
			groups[len(groups)-1].docs = append(groups[len(groups)-1].docs, v.Docs)
		} else {
			groups = append(groups, varGroup{names: []string{v.Name}, typ: v.Type, docs: [][]string{v.Docs}})
		}
	}
	for _, g := range groups {
		if comments {
			for _, docs := range g.docs {
				for _, d := range docs {
					fmt.Fprintf(sb, "// %s\n", d)
				}
			}
		}
		fmt.Fprintf(sb, "var %s %s\n", strings.Join(g.names, ", "), g.typ)
	}
}

func printConstsToon(consts []Constant, comments bool, sb *strings.Builder) {
	for _, c := range consts {
		if comments && len(c.Docs) > 0 {
			for _, d := range c.Docs {
				fmt.Fprintf(sb, "// %s\n", d)
			}
		}
		if c.Value != "" {
			fmt.Fprintf(sb, "const %s = %s\n", c.Name, c.Value)
		} else {
			fmt.Fprintf(sb, "const %s\n", c.Name)
		}
	}
}

// Helper Functions

func getPosition(p *ParsedInfo) (string, string) {
	if p.File != "" {
		return "file", p.File
	}
	return "directory", p.Directory
}

func shouldIgnorePosition(p *ParsedInfo, ignoreRules []string, posType string) bool {
	if len(ignoreRules) == 0 {
		return false
	}
	posValue := p.Directory
	if posType == "file" {
		posValue = p.File
	}
	for _, ignore := range ignoreRules {
		expression, err := govaluate.NewEvaluableExpression(ignore)
		if err != nil {
			continue
		}
		parameters := map[string]interface{}{posType: posValue}
		result, _ := expression.Evaluate(parameters)
		if result == true {
			return true
		}
	}
	return false
}

func formatAndPrint(sb *strings.Builder, text string) string {
	fmt.Fprint(sb, text)
	return strings.Replace(text, ">", ":", 1)
}

func processStructsForGrepIndex(pkg Package, pkg_ string, sb *strings.Builder) {
	for _, s := range pkg.Structs {
		hasMethods := "has_methods"
		if len(s.Methods) == 0 {
			hasMethods = "no_methods"
		}
		struct_ := formatAndPrint(sb, fmt.Sprintf("struct> %s %s %s", s.Name, hasMethods, pkg_))

		for _, f := range s.Fields {
			formatAndPrint(sb, fmt.Sprintf("field> %s (%s) %s", f.Name, f.Type, struct_))
		}
		for _, m := range s.Methods {
			method_ := formatAndPrint(sb, fmt.Sprintf("method> %s %s", m.Name, struct_))
			paramsString := processParamsForGrepIndex(m.Params, method_, sb)
			for _, r := range m.Returns {
				formatAndPrint(sb, fmt.Sprintf("return> %s %s %s", r.Type, paramsString, method_))
			}
		}
	}
}

func processParamsForGrepIndex(params []Param, method_ string, sb *strings.Builder) string {
	var paramsSlice []string
	for _, p := range params {
		formatAndPrint(sb, fmt.Sprintf("param> %s (%s) %s", p.Name, p.Type, method_))
		paramsSlice = append(paramsSlice, fmt.Sprintf("param %s (%s)", p.Name, p.Type))
	}
	if len(paramsSlice) == 0 {
		return "no_params"
	}
	return strings.Join(paramsSlice, " ")
}

func processFunctionsForGrepIndex(pkg Package, pkg_ string, sb *strings.Builder) {
	for _, f := range pkg.Functions {
		function_ := formatAndPrint(sb, fmt.Sprintf("function> %s %s", f.Name, pkg_))
		for _, p := range f.Params {
			formatAndPrint(sb, fmt.Sprintf("param> %s (%s) %s", p.Name, p.Type, function_))
		}
	}
}

func processVariablesForGrepIndex(pkg Package, posType, posValue string, sb *strings.Builder) {
	for _, v := range pkg.Variables {
		formatAndPrint(sb, fmt.Sprintf("variable> %s (%s) package %s %s:%s", v.Name, v.Type, pkg.Package, posType, posValue))
	}
}

func processConstantsForGrepIndex(pkg Package, pkg_ string, sb *strings.Builder) {
	for _, c := range pkg.Constants {
		formatAndPrint(sb, fmt.Sprintf("constant> %s %s", c.Name, pkg_))
	}
}

func processInterfacesForGrepIndex(pkg Package, pkg_ string, sb *strings.Builder) {
	for _, i := range pkg.Interfaces {
		interface_ := formatAndPrint(sb, fmt.Sprintf("interface> %s %s", i.Name, pkg_))
		for _, m := range i.Methods {
			method_ := formatAndPrint(sb, fmt.Sprintf("method> %s %s", m.Name, interface_))
			for _, p := range m.Params {
				formatAndPrint(sb, fmt.Sprintf("param> %s (%s) %s", p.Name, p.Type, method_))
			}
		}
	}
}

func shouldIgnoreStruct(s Struct, pkgName string, ignoreRules []string) bool {
	if len(ignoreRules) == 0 {
		return false
	}
	structParams := map[string]interface{}{
		"package":            pkgName,
		"struct_name":        s.Name,
		"struct_len_fields":  len(s.Fields),
		"struct_len_methods": len(s.Methods),
	}
	for _, ignore := range ignoreRules {
		expression, err := govaluate.NewEvaluableExpression(ignore)
		if err != nil {
			continue
		}
		result, _ := expression.Evaluate(structParams)
		if result == true {
			return true
		}
	}
	return false
}

func shouldIncludeStruct(s Struct, plainStructs, structsWithMethod bool) bool {
	if !plainStructs && len(s.Methods) == 0 {
		return false
	}
	if !structsWithMethod && len(s.Methods) > 0 {
		return false
	}
	return true
}

func shouldIgnoreField(f Field, ignoreRules []string) bool {
	if len(ignoreRules) == 0 {
		return false
	}
	fieldParams := map[string]interface{}{
		"field_name": f.Name,
		"field_type": f.Type,
	}
	for _, ignore := range ignoreRules {
		expression, err := govaluate.NewEvaluableExpression(ignore)
		if err != nil {
			continue
		}
		result, _ := expression.Evaluate(fieldParams)
		if result == true {
			return true
		}
	}
	return false
}

func shouldIgnoreMethod(m Method, ignoreRules []string) bool {
	if len(ignoreRules) == 0 {
		return false
	}
	methodParams := map[string]interface{}{
		"method_name":        m.Name,
		"method_len_params":  len(m.Params),
		"method_len_returns": len(m.Returns),
	}
	for _, ignore := range ignoreRules {
		expression, err := govaluate.NewEvaluableExpression(ignore)
		if err != nil {
			continue
		}
		result, _ := expression.Evaluate(methodParams)
		if result == true {
			return true
		}
	}
	return false
}

func shouldIgnoreFunction(f Function, ignoreRules []string) bool {
	if len(ignoreRules) == 0 {
		return false
	}
	funcParams := map[string]interface{}{
		"function_name":        f.Name,
		"function_len_params":  len(f.Params),
		"function_len_returns": len(f.Returns),
	}
	for _, ignore := range ignoreRules {
		expression, err := govaluate.NewEvaluableExpression(ignore)
		if err != nil {
			continue
		}
		result, _ := expression.Evaluate(funcParams)
		if result == true {
			return true
		}
	}
	return false
}

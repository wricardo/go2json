package go2json

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/printer"
	"go/token"
	"strings"
)

// extractParsedInfo extracts parsed information from the AST.
// moduleName is the name of the module being parsed, as seen on the go.mod file.
func extractParsedInfo(packages map[string]*ast.Package, moduleName string, relModPath string) (*ParsedInfo, error) {
	output := &ParsedInfo{
		Modules: make([]Module, 0, len(packages)),
	}

	// Construct full name, avoiding trailing slash for root module
	trimmedPath := strings.TrimLeft(relModPath, "./")
	fullName := moduleName
	if trimmedPath != "" {
		fullName = moduleName + "/" + trimmedPath
	}

	m := &Module{
		RootModuleName:    moduleName,
		RelativeDirectory: relModPath,
		FullName:          fullName,
		Packages:          make([]Package, 0, len(packages)),
	}

	for _, pkg := range packages {
		outPkg := Package{
			Structs:   make([]Struct, 0),
			Functions: make([]Function, 0),
			Variables: make([]Variable, 0),
			Constants: make([]Constant, 0),
			Imports:   make([]Import, 0),
			TypeDefs:  make([]TypeDef, 0),
		}

		docPkg := doc.New(pkg, "", doc.AllDecls|doc.AllMethods|doc.PreserveAST)
		outPkg.Package = pkg.Name

		imports, err := extractImports(pkg)
		if err != nil {
			return nil, err
		}
		outPkg.Imports = imports

		structs, err := extractStructs(docPkg, outPkg)
		if err != nil {
			return nil, err
		}
		outPkg.Structs = append(outPkg.Structs, structs...)

		interfaces, err := extractInterfaces(docPkg, outPkg)
		if err != nil {
			return nil, err
		}
		outPkg.Interfaces = append(outPkg.Interfaces, interfaces...)

		typeDefs, err := extractTypeDefs(docPkg, outPkg)
		if err != nil {
			return nil, err
		}
		outPkg.TypeDefs = append(outPkg.TypeDefs, typeDefs...)

		structMap := make(map[string]*Struct)
		for i := range outPkg.Structs {
			structMap[outPkg.Structs[i].Name] = &outPkg.Structs[i]
		}

		functions, methods, err := extractFunctionsAndMethods(pkg, outPkg)
		if err != nil {
			return nil, err
		}
		outPkg.Functions = append(outPkg.Functions, functions...)

		for _, method := range methods {
			receiverName := strings.TrimPrefix(method.Receiver, "*")
			if structPtr, ok := structMap[receiverName]; ok {
				method.PtrStruct = structPtr
				structPtr.Methods = append(structPtr.Methods, method)
			}
		}

		constants, variables, err := extractConstantsVariables(pkg, outPkg)
		if err != nil {
			return nil, err
		}
		outPkg.Constants = append(outPkg.Constants, constants...)
		outPkg.Variables = append(outPkg.Variables, variables...)

		m.Packages = append(m.Packages, outPkg)
	}
	output.Modules = append(output.Modules, *m)
	output.Packages = m.Packages

	return output, nil
}

// extractStructs extracts structs from the provided documentation package.
func extractStructs(docPkg *doc.Package, ourPkg Package) ([]Struct, error) {
	var structs []Struct
	for _, t := range docPkg.Types {
		if t == nil || t.Decl == nil {
			return nil, errors.New("t or t.Decl is nil")
		}

		for _, spec := range t.Decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				return nil, errors.New("not a *ast.TypeSpec")
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if ok {
				parsedStruct := Struct{
					Name:       t.Name,
					Fields:     make([]Field, 0, len(structType.Fields.List)),
					Docs:       getDocsForStruct(t.Doc),
					Methods:    make([]Method, 0),
					IsExported: ast.IsExported(t.Name),
					PtrPackage: &ourPkg,
				}

				var defBuf bytes.Buffer
				defBuf.WriteString(fmt.Sprintf("type %s struct {\n", t.Name))
				for _, field := range structType.Fields.List {
					defBuf.WriteString("\t")
					for i, name := range field.Names {
						if i > 0 {
							defBuf.WriteString(", ")
						}
						defBuf.WriteString(name.Name)
					}
					if len(field.Names) > 0 {
						defBuf.WriteString(" ")
					}
					defBuf.WriteString(exprToString(field.Type))
					if field.Tag != nil {
						defBuf.WriteString(" ")
						defBuf.WriteString(field.Tag.Value)
					}
					if field.Comment != nil && len(field.Comment.List) > 0 {
						defBuf.WriteString(" ")
						defBuf.WriteString(field.Comment.List[0].Text)
					}
					defBuf.WriteString("\n")
				}
				defBuf.WriteString("}")
				parsedStruct.Definition = defBuf.String()

				for _, fvalue := range structType.Fields.List {
					name := ""
					if len(fvalue.Names) > 0 {
						name = fvalue.Names[0].Obj.Name
					}

					field := Field{
						Name:        name,
						Type:        "",
						Tag:         "",
						TypeDetails: TypeDetails{},
					}

					if len(field.Name) > 0 {
						field.Private = strings.ToLower(string(field.Name[0])) == string(field.Name[0])
					}

					if fvalue.Doc != nil {
						field.Docs = getDocsForFieldAst(fvalue.Doc)
					}

					if fvalue.Comment != nil {
						field.Comment = cleanDocText(fvalue.Comment.Text())
					}

					if fvalue.Tag != nil {
						field.Tag = strings.Trim(fvalue.Tag.Value, "`")
					}

					typeDetails, err := getFullType(fvalue.Type, ourPkg)
					if err != nil {
						return nil, err
					}
					field.TypeDetails = *typeDetails
					field.Pointer = typeDetails.IsPointer
					field.Slice = typeDetails.IsSlice
					field.Type = typeDetails.TypeName

					parsedStruct.Fields = append(parsedStruct.Fields, field)
				}

				structs = append(structs, parsedStruct)
			}
		}
	}
	return structs, nil
}

// extractInterfaces extracts interfaces from the provided documentation package.
func extractInterfaces(docPkg *doc.Package, pkg Package) ([]Interface, error) {
	var interfaces []Interface
	for _, t := range docPkg.Types {
		if t == nil || t.Decl == nil {
			return nil, errors.New("t or t.Decl is nil")
		}

		for _, spec := range t.Decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				return nil, errors.New("not a *ast.TypeSpec")
			}

			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if ok {
				parsedInterface := Interface{
					Name:       t.Name,
					Methods:    make([]Method, 0),
					Docs:       getDocsForStruct(t.Doc),
					IsExported: ast.IsExported(t.Name),
					PtrPackage: &pkg,
				}

				var defBuf bytes.Buffer
				defBuf.WriteString(fmt.Sprintf("type %s interface {\n", t.Name))

				for _, m := range interfaceType.Methods.List {
					if funcType, ok := m.Type.(*ast.FuncType); ok {
						methodDef := m.Names[0].Name + "(" + formatParams(funcType.Params, pkg) + ")"
						if funcType.Results != nil && len(funcType.Results.List) > 0 {
							returnStr := formatParams(funcType.Results, pkg)
							if len(funcType.Results.List) > 1 {
								methodDef += " (" + returnStr + ")"
							} else {
								methodDef += " " + returnStr
							}
						}

						methodName := m.Names[0].Name
						method := Method{
							Name:    methodName,
							Params:  extractParams(funcType.Params, pkg),
							Returns: extractParams(funcType.Results, pkg),
							Docs:    getDocsForFieldAst(m.Doc),
							Signature: fmt.Sprintf("%s(%s) (%s)", methodName,
								formatParams(funcType.Params, pkg), formatParams(funcType.Results, pkg)),
							Definition:  methodDef,
							IsExported:  ast.IsExported(methodName),
							IsTest:      isTestFunction(methodName),
							IsBenchmark: isBenchmarkFunction(methodName),
						}

						defBuf.WriteString("\t")
						defBuf.WriteString(m.Names[0].Name)
						defBuf.WriteString("(")
						defBuf.WriteString(formatParams(funcType.Params, pkg))
						defBuf.WriteString(")")

						if funcType.Results != nil && len(funcType.Results.List) > 0 {
							defBuf.WriteString(" ")
							returnStr := formatParams(funcType.Results, pkg)
							if len(funcType.Results.List) > 1 {
								defBuf.WriteString("(")
								defBuf.WriteString(returnStr)
								defBuf.WriteString(")")
							} else {
								defBuf.WriteString(returnStr)
							}
						}
						defBuf.WriteString("\n")

						parsedInterface.Methods = append(parsedInterface.Methods, method)
					}
				}

				defBuf.WriteString("}")
				parsedInterface.Definition = defBuf.String()

				interfaces = append(interfaces, parsedInterface)
			}
		}
	}
	return interfaces, nil
}

// extractTypeDefs extracts named type declarations that are neither structs nor interfaces.
func extractTypeDefs(docPkg *doc.Package, pkg Package) ([]TypeDef, error) {
	var typeDefs []TypeDef
	for _, t := range docPkg.Types {
		if t == nil || t.Decl == nil {
			continue
		}
		for _, spec := range t.Decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
				continue
			}
			if _, isIface := typeSpec.Type.(*ast.InterfaceType); isIface {
				continue
			}

			underlying := exprToString(typeSpec.Type)
			def := fmt.Sprintf("type %s %s", t.Name, underlying)

			typeDefs = append(typeDefs, TypeDef{
				Name:       t.Name,
				Underlying: underlying,
				Docs:       getDocsForStruct(t.Doc),
				Definition: def,
				IsExported: ast.IsExported(t.Name),
				PtrPackage: &pkg,
			})
		}
	}
	return typeDefs, nil
}

// extractImports extracts unique imports from the provided package.
func extractImports(pkg *ast.Package) ([]Import, error) {
	importMap := make(map[string]Import)
	for _, file := range pkg.Files {
		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, "\"")
			var importName string
			if importSpec.Name != nil {
				importName = importSpec.Name.Name
			}
			importMap[importPath] = Import{
				Name: importName,
				Path: importPath,
			}
		}
	}

	var imports []Import
	for _, imp := range importMap {
		imports = append(imports, imp)
	}
	return imports, nil
}

// extractConstantsVariables extracts constants and variables from the provided package.
func extractConstantsVariables(pkg *ast.Package, ourPkg Package) ([]Constant, []Variable, error) {
	var constants []Constant
	var variables []Variable

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			switch genDecl.Tok {
			case token.CONST:
				for _, spec := range genDecl.Specs {
					valSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for i, name := range valSpec.Names {
						constant := Constant{
							Name:  name.Name,
							Value: "",
							Docs:  getDocsForFieldAst(valSpec.Doc),
						}
						if i < len(valSpec.Values) {
							constant.Value = exprToString(valSpec.Values[i])
						}
						constants = append(constants, constant)
					}
				}
			case token.VAR:
				for _, spec := range genDecl.Specs {
					valSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range valSpec.Names {
						varType := ""
						if valSpec.Type != nil {
							tmp, _ := getFullType(valSpec.Type, ourPkg)
							varType = tmp.TypeName
						}
						variable := Variable{
							Name: name.Name,
							Type: varType,
							Docs: getDocsForFieldAst(valSpec.Doc),
						}
						variables = append(variables, variable)
					}
				}
			}
		}
	}

	return constants, variables, nil
}

func extractParams(fieldList *ast.FieldList, pkg Package) []Param {
	if fieldList == nil {
		return nil
	}
	params := make([]Param, 0, len(fieldList.List))
	for _, field := range fieldList.List {
		paramType, err := getFullType(field.Type, pkg)
		if err != nil {
			continue
		}
		for _, name := range field.Names {
			params = append(params, Param{Name: name.Name, Type: paramType.TypeName, TypeDetails: *paramType})
		}
		if len(field.Names) == 0 {
			params = append(params, Param{Name: "", Type: paramType.TypeName, TypeDetails: *paramType})
		}
	}
	return params
}

func formatParams(fields *ast.FieldList, pkg Package) string {
	if fields == nil {
		return ""
	}
	paramStrings := []string{}
	for _, param := range extractParams(fields, pkg) {
		if param.Name != "" {
			paramStrings = append(paramStrings, fmt.Sprintf("%s %s", param.Name, param.Type))
		} else {
			paramStrings = append(paramStrings, param.Type)
		}
	}
	return strings.Join(paramStrings, ", ")
}

// parseFunctionDecl extracts function details from an *ast.FuncDecl node.
func parseFunctionDecl(funcDecl *ast.FuncDecl, docs string, pkg Package) (Function, error) {
	funcName := funcDecl.Name.Name
	function := Function{
		Name:        funcName,
		Docs:        getDocsForField([]string{docs}),
		IsExported:  ast.IsExported(funcName),
		IsTest:      isTestFunction(funcName),
		IsBenchmark: isBenchmarkFunction(funcName),
	}

	params := []Param{}
	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			paramType, err := getFullType(param.Type, pkg)
			if err != nil {
				return function, err
			}

			if len(param.Names) > 0 {
				for _, name := range param.Names {
					p := &Param{
						Name:        name.Name,
						Type:        paramType.TypeName,
						TypeDetails: *paramType,
						PtrFunc:     &function,
					}
					params = append(params, *p)
				}
			} else {
				p := &Param{
					Type:        paramType.TypeName,
					TypeDetails: *paramType,
					PtrFunc:     &function,
				}
				params = append(params, *p)
			}
		}
	}
	function.Params = params

	returns := []Param{}
	if funcDecl.Type.Results != nil {
		for _, result := range funcDecl.Type.Results.List {
			returnType, err := getFullType(result.Type, pkg)
			if err != nil {
				return function, err
			}

			if len(result.Names) > 0 {
				for _, name := range result.Names {
					returns = append(returns, Param{
						Name:    name.Name,
						Type:    returnType.TypeName,
						PtrFunc: &function,
					})
				}
			} else {
				returns = append(returns, Param{
					Type:    returnType.TypeName,
					PtrFunc: &function,
				})
			}
		}
	}
	function.Returns = returns

	var bodyBuf bytes.Buffer
	if funcDecl.Body != nil {
		err := format.Node(&bodyBuf, token.NewFileSet(), funcDecl.Body)
		if err != nil {
			return function, err
		}
		function.Body = bodyBuf.String()
	}

	paramStrings := []string{}
	for _, param := range function.Params {
		if param.Name != "" {
			paramStrings = append(paramStrings, param.Name+" "+param.Type)
		} else {
			paramStrings = append(paramStrings, param.Type)
		}
	}

	returnStrings := []string{}
	for _, ret := range function.Returns {
		if ret.Name != "" {
			returnStrings = append(returnStrings, ret.Name+" "+ret.Type)
		} else {
			returnStrings = append(returnStrings, ret.Type)
		}
	}

	signature := fmt.Sprintf("%s(%s)", function.Name, strings.Join(paramStrings, ", "))
	if len(returnStrings) > 0 {
		returnsStr := strings.Join(returnStrings, ", ")
		if len(function.Returns) > 1 {
			returnsStr = "(" + returnsStr + ")"
		}
		signature += fmt.Sprintf(" %s", returnsStr)
	}
	function.Signature = signature

	definition := fmt.Sprintf("func %s(%s)", function.Name, strings.Join(paramStrings, ", "))
	if len(returnStrings) > 0 {
		returnsStr := strings.Join(returnStrings, ", ")
		if len(function.Returns) > 1 {
			returnsStr = "(" + returnsStr + ")"
		}
		definition += fmt.Sprintf(" %s", returnsStr)
	}
	function.Definition = definition

	return function, nil
}

// parseMethodDecl extracts method details from an *ast.FuncDecl node.
func parseMethodDecl(funcDecl *ast.FuncDecl, docs string, ourPkg Package) (Method, error) {
	receiverType, _ := getFullType(funcDecl.Recv.List[0].Type, ourPkg)
	methodName := funcDecl.Name.Name
	method := Method{
		Name:        methodName,
		Receiver:    receiverType.TypeName,
		Docs:        getDocsForField([]string{docs}),
		IsExported:  ast.IsExported(methodName),
		IsTest:      isTestFunction(methodName),
		IsBenchmark: isBenchmarkFunction(methodName),
	}

	params := []Param{}
	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			paramType, err := getFullType(param.Type, ourPkg)
			if err != nil {
				return method, err
			}
			if (paramType.PackageName == nil || *paramType.PackageName == "") && ourPkg.Package != "" {
				paramType.PackageName = &ourPkg.Package
			}

			if len(param.Names) > 0 {
				for _, name := range param.Names {
					p := &Param{
						Name:        name.Name,
						Type:        paramType.TypeName,
						TypeDetails: *paramType,
						PtrMethod:   &method,
					}
					params = append(params, *p)
				}
			} else {
				p := &Param{
					Name:        "",
					Type:        paramType.TypeName,
					TypeDetails: *paramType,
					PtrMethod:   &method,
				}
				params = append(params, *p)
			}
		}
	}
	method.Params = params

	returns := []Param{}
	if funcDecl.Type.Results != nil {
		for _, result := range funcDecl.Type.Results.List {
			returnType, err := getFullType(result.Type, ourPkg)
			if err != nil {
				return method, err
			}

			if len(result.Names) > 0 {
				for _, name := range result.Names {
					p := &Param{
						Name:        name.Name,
						Type:        returnType.TypeName,
						TypeDetails: *returnType,
						PtrMethod:   &method,
					}
					returns = append(returns, *p)
				}
			} else {
				p := &Param{
					Name:        "",
					Type:        returnType.TypeName,
					TypeDetails: *returnType,
					PtrMethod:   &method,
				}
				returns = append(returns, *p)
			}
		}
	}
	method.Returns = returns

	var bodyBuf bytes.Buffer
	if funcDecl.Body != nil {
		err := format.Node(&bodyBuf, token.NewFileSet(), funcDecl.Body)
		if err != nil {
			return method, err
		}
		method.Body = bodyBuf.String()
	}

	paramStrings := []string{}
	for _, param := range method.Params {
		if param.Name != "" {
			paramStrings = append(paramStrings, param.Name+" "+param.Type)
		} else {
			paramStrings = append(paramStrings, param.Type)
		}
	}

	returnStrings := []string{}
	for _, ret := range method.Returns {
		if ret.Name != "" {
			returnStrings = append(returnStrings, ret.Name+" "+ret.Type)
		} else {
			returnStrings = append(returnStrings, ret.Type)
		}
	}

	signature := fmt.Sprintf("%s(%s)", method.Name, strings.Join(paramStrings, ", "))
	if len(returnStrings) > 0 {
		returnsStr := strings.Join(returnStrings, ", ")
		if len(method.Returns) > 1 {
			returnsStr = "(" + returnsStr + ")"
		}
		signature += fmt.Sprintf(" %s", returnsStr)
	}
	method.Signature = signature

	definition := fmt.Sprintf("func (%s) %s(%s)", method.Receiver, method.Name, strings.Join(paramStrings, ", "))
	if len(returnStrings) > 0 {
		returnsStr := strings.Join(returnStrings, ", ")
		if len(method.Returns) > 1 {
			returnsStr = "(" + returnsStr + ")"
		}
		definition += fmt.Sprintf(" %s", returnsStr)
	}
	method.Definition = definition

	return method, nil
}

// extractFunctionsAndMethods traverses the AST to extract all functions and methods.
func extractFunctionsAndMethods(pkg *ast.Package, ourPkg Package) ([]Function, []Method, error) {
	var functions []Function
	var methods []Method

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Recv == nil {
					function, err := parseFunctionDecl(funcDecl, funcDecl.Doc.Text(), ourPkg)
					if err != nil {
						return nil, nil, err
					}
					functions = append(functions, function)
				} else {
					method, err := parseMethodDecl(funcDecl, funcDecl.Doc.Text(), ourPkg)
					if err != nil {
						return nil, nil, err
					}
					methods = append(methods, method)
				}
			}
		}
	}

	return functions, methods, nil
}

// isTestFunction checks if a function name follows the Go test function naming convention.
func isTestFunction(name string) bool {
	return strings.HasPrefix(name, "Test") && len(name) > 4 && strings.ToUpper(name[4:5]) == name[4:5]
}

// isBenchmarkFunction checks if a function name follows the Go benchmark function naming convention.
func isBenchmarkFunction(name string) bool {
	return strings.HasPrefix(name, "Benchmark") && len(name) > 9 && strings.ToUpper(name[9:10]) == name[9:10]
}

// resolveFullImportPath resolves a short package name to its full import path
// by looking through the package's imports list.
func resolveFullImportPath(shortPackageName string, pkg Package) string {
	for _, imp := range pkg.Imports {
		importName := imp.Name
		if importName == "" {
			parts := strings.Split(imp.Path, "/")
			if len(parts) > 0 {
				importName = parts[len(parts)-1]
			}
		}
		if importName == shortPackageName {
			return imp.Path
		}
	}
	return shortPackageName
}

// getFullType processes an AST expression and returns comprehensive TypeDetails.
func getFullType(expr ast.Expr, pkg Package) (*TypeDetails, error) {
	coallesce := func(a, b *string) *string {
		if a != nil && *a != "" {
			return a
		}
		return b
	}
	tr := TypeDetails{
		IsPointer:  false,
		IsSlice:    false,
		IsMap:      false,
		IsBuiltin:  false,
		IsExternal: false,
	}

	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string", "int", "int32", "int64", "float32", "float64", "bool", "byte", "rune", "error":
			tr.IsBuiltin = true
		default:
			if pkg.Package != "" {
				tr.PackageName = &pkg.Package
			}
			tr.IsBuiltin = false
		}
		tr.TypeName = t.Name
		tr.Type = &t.Name

	case *ast.SelectorExpr:
		tr.IsExternal = true
		xFullType, err := getFullType(t.X, pkg)
		if err != nil {
			return nil, err
		}
		tr.TypeName = xFullType.TypeName + "." + t.Sel.Name

		shortPackageName := xFullType.TypeName
		fullImportPath := resolveFullImportPath(shortPackageName, pkg)

		tr.PackageName = &shortPackageName
		tr.Package = &fullImportPath
		tr.Type = &t.Sel.Name
		tr.TypeReferences = []TypeReference{
			{
				Package:     &fullImportPath,
				PackageName: &shortPackageName,
				Name:        t.Sel.Name,
			},
		}

	case *ast.StarExpr:
		tr.IsPointer = true
		innerFullType, err := getFullType(t.X, pkg)
		if err != nil {
			return nil, err
		}
		tr.IsSlice = tr.IsSlice || innerFullType.IsSlice
		tr.IsMap = tr.IsMap || innerFullType.IsMap
		tr.TypeName = "*" + innerFullType.TypeName
		tr.TypeReferences = append(tr.TypeReferences, innerFullType.TypeReferences...)
		tr.Package = coallesce(tr.Package, innerFullType.Package)
		tr.PackageName = coallesce(tr.PackageName, innerFullType.PackageName)
		tr.Type = coallesce(tr.Type, innerFullType.Type)

	case *ast.IndexExpr:
		xFullType, err := getFullType(t.X, pkg)
		if err != nil {
			return nil, err
		}
		indexFullType, err := getFullType(t.Index, pkg)
		if err != nil {
			return nil, err
		}
		tr.TypeName = fmt.Sprintf("%s[%s]", xFullType.TypeName, indexFullType.TypeName)
		tr.Type = &xFullType.TypeName
		tr.Package = coallesce(tr.Package, indexFullType.Package)
		tr.PackageName = coallesce(tr.PackageName, indexFullType.PackageName)

	case *ast.IndexListExpr:
		xFullType, err := getFullType(t.X, pkg)
		if err != nil {
			return nil, err
		}
		indices := []string{}
		for _, index := range t.Indices {
			indexFullType, err := getFullType(index, pkg)
			if err != nil {
				return nil, err
			}
			indices = append(indices, indexFullType.TypeName)
		}
		tr.TypeName = fmt.Sprintf("%s[%s]", xFullType.TypeName, strings.Join(indices, ", "))
		tr.Type = &xFullType.TypeName
		tr.Package = coallesce(tr.Package, xFullType.Package)
		tr.PackageName = coallesce(tr.PackageName, xFullType.PackageName)

	case *ast.ArrayType:
		tr.IsSlice = true
		eltFullType, err := getFullType(t.Elt, pkg)
		if err != nil {
			return nil, err
		}
		tr.TypeName = "[]" + eltFullType.TypeName
		tr.Type = &eltFullType.TypeName
		tr.Package = coallesce(tr.Package, eltFullType.Package)
		tr.PackageName = coallesce(tr.PackageName, eltFullType.PackageName)

	case *ast.ChanType:
		dir := ""
		switch t.Dir {
		case ast.RECV:
			dir = "<-chan "
		case ast.SEND:
			dir = "chan<- "
		default:
			dir = "chan "
		}
		valueFullType, err := getFullType(t.Value, pkg)
		if err != nil {
			return nil, err
		}
		tr.TypeName = fmt.Sprintf("%s%s", dir, valueFullType.TypeName)
		tr.Type = &valueFullType.TypeName
		tr.Package = coallesce(tr.Package, valueFullType.Package)
		tr.PackageName = coallesce(tr.PackageName, valueFullType.PackageName)

	case *ast.MapType:
		keyFullType, err := getFullType(t.Key, pkg)
		if err != nil {
			return nil, err
		}
		valueFullType, err := getFullType(t.Value, pkg)
		if err != nil {
			return nil, err
		}
		tr.TypeName = fmt.Sprintf("map[%s]%s", keyFullType.TypeName, valueFullType.TypeName)
		tr.IsMap = true
		tr.Type = &valueFullType.TypeName
		tr.Package = coallesce(tr.Package, valueFullType.Package)
		tr.PackageName = coallesce(tr.PackageName, valueFullType.PackageName)

	case *ast.FuncType:
		params, err := fieldListToString(t.Params, pkg)
		if err != nil {
			return nil, err
		}
		results, err := fieldListToString(t.Results, pkg)
		if err != nil {
			return nil, err
		}
		tr.TypeName = fmt.Sprintf("func(%s) (%s)", params, results)
		tr.Type = &tr.TypeName

	case *ast.InterfaceType:
		methodsStr, err := fieldListToString(t.Methods, pkg)
		if err != nil {
			return nil, err
		}
		tr.TypeName = fmt.Sprintf("interface{%s}", methodsStr)
		tr.Type = &tr.TypeName

	case *ast.StructType:
		fieldsStr, err := fieldListToString(t.Fields, pkg)
		if err != nil {
			return nil, err
		}
		tr.TypeName = fmt.Sprintf("struct{%s}", fieldsStr)
		tr.Type = &tr.TypeName

	case *ast.Ellipsis:
		eltFullType, err := getFullType(t.Elt, pkg)
		if err != nil {
			return nil, err
		}
		if eltFullType == nil {
			tr.TypeName = "..."
		} else {
			if eltFullType.PackageName == nil {
				tr.TypeName = "..." + eltFullType.TypeName
			} else {
				tr.TypeName = "..." + justTypeString(eltFullType.TypeName, eltFullType.IsPointer, eltFullType.IsSlice, &ast.Ident{Name: *eltFullType.PackageName})
			}
		}
		tr.Type = &eltFullType.TypeName

	case *ast.ParenExpr:
		return getFullType(t.X, pkg)

	default:
		return nil, fmt.Errorf("unsupported type: %T", expr)
	}

	return &tr, nil
}

// fieldListToString converts an *ast.FieldList to its string representation.
func fieldListToString(fl *ast.FieldList, pkg Package) (string, error) {
	if fl == nil {
		return "", nil
	}
	fields := []string{}
	for _, field := range fl.List {
		typeRef, err := getFullType(field.Type, pkg)
		if err != nil {
			return "", err
		}
		if len(field.Names) == 0 {
			fields = append(fields, typeRef.TypeName)
		} else {
			for _, name := range field.Names {
				fields = append(fields, fmt.Sprintf("%s %s", name.Name, typeRef.TypeName))
			}
		}
	}
	return strings.Join(fields, ", "), nil
}

func justTypeString(typeName string, isPointer, isSlice bool, packageName *ast.Ident) string {
	if isPointer {
		typeName = "*" + typeName
	}
	if isSlice {
		typeName = "[]" + typeName
	}
	if packageName != nil {
		typeName = packageName.Name + "." + typeName
	}
	return typeName
}

func exprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, token.NewFileSet(), expr)
	if err != nil {
		return "<err>"
	}
	return buf.String()
}

func getDocsForStruct(doc string) []string {
	trimmed := strings.Trim(doc, "\n")
	if trimmed == "" {
		return []string{}
	}
	tmp := strings.Split(trimmed, "\n")
	docs := make([]string, 0, len(tmp))
	for _, v := range tmp {
		clean := cleanDocText(v)
		if clean == "" {
			continue
		}
		docs = append(docs, clean)
	}
	return docs
}

func getDocsForFieldAst(cg *ast.CommentGroup) []string {
	if cg == nil {
		return []string{}
	}
	docs := make([]string, 0, len(cg.List))
	for _, v := range cg.List {
		docs = append(docs, cleanDocText(v.Text))
	}
	return docs
}

func getDocsForField(list []string) []string {
	docs := make([]string, 0, len(list))
	for _, v := range list {
		clean := cleanDocText(v)
		if clean == "" {
			continue
		}
		docs = append(docs, clean)
	}
	return docs
}

func cleanDocText(doc string) string {
	reverseString := func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}

	if strings.HasPrefix(doc, "// ") {
		doc = strings.Replace(doc, "// ", "", 1)
	} else if strings.HasPrefix(doc, "//") {
		doc = strings.Replace(doc, "//", "", 1)
	} else if strings.HasPrefix(doc, "/*") {
		doc = strings.Replace(doc, "/*", "", 1)
	}
	if strings.HasSuffix(doc, "*/") {
		doc = reverseString(strings.Replace(reverseString(doc), "/*", "", 1))
	}
	return strings.Trim(strings.Trim(doc, " "), "\n")
}

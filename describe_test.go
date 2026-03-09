package go2json

import (
	"strings"
	"testing"
)

func parseExamples(t *testing.T) []*ParsedInfo {
	t.Helper()
	parsed, err := ParseDirectoryRecursive("./examples/structparser")
	if err != nil {
		t.Fatalf("ParseDirectoryRecursive failed: %v", err)
	}
	return parsed
}

func TestDescribeType_Depth0(t *testing.T) {
	parsed := parseExamples(t)
	result, err := DescribeType("FirstStruct", parsed, 0)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}

	structs := collectStructNames(result)
	if len(structs) != 1 {
		t.Fatalf("expected 1 struct, got %d: %v", len(structs), structs)
	}
	if structs[0] != "FirstStruct" {
		t.Fatalf("expected FirstStruct, got %s", structs[0])
	}

	// Depth 0 should not include referenced types
	typeDefs := collectTypeDefNames(result)
	if len(typeDefs) != 0 {
		t.Errorf("depth 0 should not include typedefs, got %v", typeDefs)
	}
}

func TestDescribeType_Depth1(t *testing.T) {
	parsed := parseExamples(t)
	result, err := DescribeType("FirstStruct", parsed, 1)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}

	structs := collectStructNames(result)
	names := make(map[string]bool)
	for _, n := range structs {
		names[n] = true
	}

	// FirstStruct has a SecondStruct field and TakesThirdStruct method
	if !names["FirstStruct"] {
		t.Error("missing FirstStruct")
	}
	if !names["SecondStruct"] {
		t.Error("missing SecondStruct")
	}
	if !names["ThirdStruct"] {
		t.Error("missing ThirdStruct (referenced by TakesThirdStruct method)")
	}
}

func TestDescribeType_Depth2_TransitiveRefs(t *testing.T) {
	// UnimplementedZivoAPIHandler.SendSms takes/returns *FirstStruct
	// At depth=1 we get FirstStruct; at depth=2 we also get SecondStruct+ThirdStruct
	parsed := parseExamples(t)
	result, err := DescribeType("UnimplementedZivoAPIHandler", parsed, 2)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}

	names := make(map[string]bool)
	for _, n := range collectStructNames(result) {
		names[n] = true
	}
	if !names["UnimplementedZivoAPIHandler"] {
		t.Error("missing UnimplementedZivoAPIHandler")
	}
	if !names["FirstStruct"] {
		t.Error("missing FirstStruct (depth 1)")
	}
	if !names["SecondStruct"] {
		t.Error("missing SecondStruct (depth 2, via FirstStruct)")
	}
}

func TestDescribeType_NotFound(t *testing.T) {
	parsed := parseExamples(t)
	_, err := DescribeType("NonExistentType", parsed, 1)
	if err == nil {
		t.Fatal("expected error for non-existent type")
	}
	if !strings.Contains(err.Error(), "NonExistentType") {
		t.Errorf("error should mention the type name, got: %v", err)
	}
}

func TestDescribeType_NoCycle(t *testing.T) {
	// UnimplementedZivoAPIHandler -> FirstStruct -> SecondStruct (and FirstStruct via SendSms)
	// BFS should terminate without infinite loop
	parsed := parseExamples(t)
	result, err := DescribeType("UnimplementedZivoAPIHandler", parsed, 10)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestDescribeType_TypeDefs(t *testing.T) {
	parsed := parseExamples(t)
	result, err := DescribeType("FirstStruct", parsed, 1)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}

	typeDefs := collectTypeDefNames(result)
	names := make(map[string]bool)
	for _, n := range typeDefs {
		names[n] = true
	}

	if !names["SomeFunc"] {
		t.Error("missing SomeFunc typedef (referenced by Func field)")
	}
	if !names["SpecialString"] {
		t.Error("missing SpecialString typedef (referenced by SpecialString field)")
	}
}

func TestDescribeType_PrettyPrintLLM(t *testing.T) {
	parsed := parseExamples(t)
	result, err := DescribeType("FirstStruct", parsed, 1)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}
	output := PrettyPrint(result, "llm", nil, true, true, true, true, true, true, true, false, false)
	if output == "" {
		t.Fatal("PrettyPrint returned empty string")
	}

	// Struct output
	if !strings.Contains(output, "type FirstStruct struct{") {
		t.Error("output should contain FirstStruct struct definition")
	}
	if !strings.Contains(output, "type SecondStruct struct{") {
		t.Error("output should contain SecondStruct struct definition")
	}

	// TypeDef output
	if !strings.Contains(output, "type SomeFunc func(a string) error") {
		t.Error("output should contain SomeFunc typedef definition")
	}
	if !strings.Contains(output, "type SpecialString string") {
		t.Error("output should contain SpecialString typedef definition")
	}

	// Methods
	if !strings.Contains(output, "MyTestMethod") {
		t.Error("output should contain method MyTestMethod")
	}
}

func TestDescribeType_PrettyPrintJSON(t *testing.T) {
	parsed := parseExamples(t)
	result, err := DescribeType("FirstStruct", parsed, 1)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}
	output := PrettyPrint(result, "json", nil, true, true, true, true, true, true, true, false, false)
	if !strings.Contains(output, `"name": "FirstStruct"`) {
		t.Error("JSON output should contain FirstStruct")
	}
	if !strings.Contains(output, `"name": "SomeFunc"`) {
		t.Error("JSON output should contain SomeFunc typedef")
	}
}

func TestDescribeType_PrettyPrintJSONOmitNulls(t *testing.T) {
	parsed := parseExamples(t)
	result, err := DescribeType("SecondStruct", parsed, 0)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}
	output := PrettyPrint(result, "json", nil, true, true, true, true, true, true, true, false, true)
	if strings.Contains(output, `"methods": null`) {
		t.Error("omit-nulls should remove null methods")
	}
	if !strings.Contains(output, `"name": "SecondStruct"`) {
		t.Error("should still contain SecondStruct")
	}
}

// --- BuildTypeIndex tests ---

func TestBuildTypeIndex_Structs(t *testing.T) {
	parsed := parseExamples(t)
	index := BuildTypeIndex(parsed)

	entry, ok := index["FirstStruct"]
	if !ok {
		t.Fatal("FirstStruct not in index")
	}
	if entry.Kind != "struct" {
		t.Errorf("expected kind struct, got %s", entry.Kind)
	}
	if entry.Struct == nil {
		t.Fatal("Struct pointer is nil")
	}
	if entry.Struct.Name != "FirstStruct" {
		t.Errorf("expected name FirstStruct, got %s", entry.Struct.Name)
	}
}

func TestBuildTypeIndex_TypeDefs(t *testing.T) {
	parsed := parseExamples(t)
	index := BuildTypeIndex(parsed)

	entry, ok := index["SomeFunc"]
	if !ok {
		t.Fatal("SomeFunc not in index")
	}
	if entry.Kind != "typedef" {
		t.Errorf("expected kind typedef, got %s", entry.Kind)
	}
	if entry.TypeDef == nil {
		t.Fatal("TypeDef pointer is nil")
	}
	if entry.TypeDef.Underlying != "func(a string) error" {
		t.Errorf("unexpected underlying type: %s", entry.TypeDef.Underlying)
	}
}

func TestBuildTypeIndex_FirstMatchWins(t *testing.T) {
	// Create two parsed infos with same type name in different packages
	pkg1 := Package{
		Package: "pkg1",
		Structs: []Struct{{Name: "Foo", Fields: []Field{{Name: "X", Type: "int"}}}},
	}
	pkg2 := Package{
		Package: "pkg2",
		Structs: []Struct{{Name: "Foo", Fields: []Field{{Name: "Y", Type: "string"}}}},
	}
	parsed := []*ParsedInfo{
		{Packages: []Package{pkg1}},
		{Packages: []Package{pkg2}},
	}
	index := BuildTypeIndex(parsed)

	entry := index["Foo"]
	if entry.Package.Package != "pkg1" {
		t.Errorf("expected first match (pkg1), got %s", entry.Package.Package)
	}
}

func TestBuildTypeIndex_Empty(t *testing.T) {
	index := BuildTypeIndex(nil)
	if len(index) != 0 {
		t.Errorf("expected empty index, got %d entries", len(index))
	}
}

// --- extractTypeNames tests ---

func TestExtractTypeNames_TypeReferences(t *testing.T) {
	pkg := strPtr("github.com/example")
	pkgName := strPtr("example")
	td := TypeDetails{
		TypeReferences: []TypeReference{
			{Package: pkg, PackageName: pkgName, Name: "Foo"},
			{Package: pkg, PackageName: pkgName, Name: "Bar"},
			{Name: "string"}, // builtin, should be excluded
		},
	}
	names := extractTypeNames(td)
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d: %v", len(names), names)
	}
	if names[0] != "Foo" || names[1] != "Bar" {
		t.Errorf("expected [Foo Bar], got %v", names)
	}
}

func TestExtractTypeNames_SingleType(t *testing.T) {
	typeName := "MyStruct"
	td := TypeDetails{Type: &typeName}
	names := extractTypeNames(td)
	if len(names) != 1 || names[0] != "MyStruct" {
		t.Errorf("expected [MyStruct], got %v", names)
	}
}

func TestExtractTypeNames_Builtin(t *testing.T) {
	typeName := "string"
	td := TypeDetails{Type: &typeName}
	names := extractTypeNames(td)
	if len(names) != 0 {
		t.Errorf("expected empty for builtin, got %v", names)
	}
}

func TestExtractTypeNames_Nil(t *testing.T) {
	td := TypeDetails{}
	names := extractTypeNames(td)
	if names != nil {
		t.Errorf("expected nil for empty TypeDetails, got %v", names)
	}
}

func TestExtractTypeNames_AllBuiltinRefs(t *testing.T) {
	td := TypeDetails{
		TypeReferences: []TypeReference{
			{Name: "int"},
			{Name: "error"},
		},
	}
	names := extractTypeNames(td)
	if len(names) != 0 {
		t.Errorf("expected empty for all-builtin refs, got %v", names)
	}
}

// --- referencedTypes tests ---

func TestReferencedTypes_Struct(t *testing.T) {
	fooType := "Foo"
	barType := "Bar"
	entry := typeEntry{
		Kind: "struct",
		Struct: &Struct{
			Fields: []Field{
				{Name: "A", TypeDetails: TypeDetails{Type: &fooType}},
				{Name: "B", TypeDetails: TypeDetails{Type: &barType}},
			},
		},
	}
	refs := referencedTypes(entry)
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d: %v", len(refs), refs)
	}
	refSet := make(map[string]bool)
	for _, r := range refs {
		refSet[r] = true
	}
	if !refSet["Foo"] || !refSet["Bar"] {
		t.Errorf("expected Foo and Bar, got %v", refs)
	}
}

func TestReferencedTypes_StructMethods(t *testing.T) {
	bazType := "Baz"
	entry := typeEntry{
		Kind: "struct",
		Struct: &Struct{
			Methods: []Method{
				{
					Params:  []Param{{TypeDetails: TypeDetails{Type: &bazType}}},
					Returns: []Param{},
				},
			},
		},
	}
	refs := referencedTypes(entry)
	if len(refs) != 1 || refs[0] != "Baz" {
		t.Errorf("expected [Baz], got %v", refs)
	}
}

func TestReferencedTypes_StructMethodReturns(t *testing.T) {
	retType := "Result"
	entry := typeEntry{
		Kind: "struct",
		Struct: &Struct{
			Methods: []Method{
				{
					Params:  []Param{},
					Returns: []Param{{TypeDetails: TypeDetails{Type: &retType}}},
				},
			},
		},
	}
	refs := referencedTypes(entry)
	if len(refs) != 1 || refs[0] != "Result" {
		t.Errorf("expected [Result], got %v", refs)
	}
}

func TestReferencedTypes_Interface(t *testing.T) {
	argType := "Request"
	retType := "Response"
	entry := typeEntry{
		Kind: "interface",
		Interface: &Interface{
			Methods: []Method{
				{
					Params:  []Param{{TypeDetails: TypeDetails{Type: &argType}}},
					Returns: []Param{{TypeDetails: TypeDetails{Type: &retType}}},
				},
			},
		},
	}
	refs := referencedTypes(entry)
	refSet := make(map[string]bool)
	for _, r := range refs {
		refSet[r] = true
	}
	if !refSet["Request"] || !refSet["Response"] {
		t.Errorf("expected Request and Response, got %v", refs)
	}
}

func TestReferencedTypes_TypeDef(t *testing.T) {
	entry := typeEntry{
		Kind:    "typedef",
		TypeDef: &TypeDef{Name: "MyFunc", Underlying: "func() error"},
	}
	refs := referencedTypes(entry)
	if len(refs) != 0 {
		t.Errorf("typedefs should return no refs, got %v", refs)
	}
}

func TestReferencedTypes_Deduplication(t *testing.T) {
	fooType := "Foo"
	entry := typeEntry{
		Kind: "struct",
		Struct: &Struct{
			Fields: []Field{
				{Name: "A", TypeDetails: TypeDetails{Type: &fooType}},
				{Name: "B", TypeDetails: TypeDetails{Type: &fooType}},
			},
		},
	}
	refs := referencedTypes(entry)
	if len(refs) != 1 {
		t.Errorf("expected deduplicated to 1 ref, got %d: %v", len(refs), refs)
	}
}

// --- extractTypeDefs tests ---

func TestExtractTypeDefs_FromParse(t *testing.T) {
	parsed, err := ParseDirectory("./examples/structparser")
	if err != nil {
		t.Fatalf("ParseDirectory failed: %v", err)
	}

	var allTypeDefs []TypeDef
	for _, pkg := range parsed.Packages {
		allTypeDefs = append(allTypeDefs, pkg.TypeDefs...)
	}

	names := make(map[string]bool)
	for _, td := range allTypeDefs {
		names[td.Name] = true
	}

	if !names["SomeFunc"] {
		t.Error("expected SomeFunc in TypeDefs")
	}
	if !names["SpecialString"] {
		t.Error("expected SpecialString in TypeDefs")
	}

	// Verify SomeFunc definition
	for _, td := range allTypeDefs {
		if td.Name == "SomeFunc" {
			if td.Underlying != "func(a string) error" {
				t.Errorf("SomeFunc underlying: expected 'func(a string) error', got %q", td.Underlying)
			}
			if td.Definition != "type SomeFunc func(a string) error" {
				t.Errorf("SomeFunc definition: expected 'type SomeFunc func(a string) error', got %q", td.Definition)
			}
			if !td.IsExported {
				t.Error("SomeFunc should be exported")
			}
		}
		if td.Name == "SpecialString" {
			if td.Underlying != "string" {
				t.Errorf("SpecialString underlying: expected 'string', got %q", td.Underlying)
			}
		}
	}
}

func TestExtractTypeDefs_ExcludesStructsAndInterfaces(t *testing.T) {
	parsed, err := ParseDirectory("./examples/structparser")
	if err != nil {
		t.Fatalf("ParseDirectory failed: %v", err)
	}

	for _, pkg := range parsed.Packages {
		for _, td := range pkg.TypeDefs {
			// None of the known struct names should appear as TypeDefs
			if td.Name == "FirstStruct" || td.Name == "SecondStruct" || td.Name == "ThirdStruct" ||
				td.Name == "SimpleStruct" || td.Name == "CommentsAndDocs" || td.Name == "privateStruct" {
				t.Errorf("struct %q should not be in TypeDefs", td.Name)
			}
		}
	}
}

// --- printFieldsToon tests ---

func TestPrintFieldsToon_GroupsConsecutiveSameType(t *testing.T) {
	fields := []Field{
		{Name: "A", Type: "int"},
		{Name: "B", Type: "int"},
		{Name: "C", Type: "string"},
	}
	var sb strings.Builder
	printFieldsToon(fields, &sb)
	output := sb.String()

	if !strings.Contains(output, "A,B int") {
		t.Errorf("expected grouped 'A,B int', got:\n%s", output)
	}
	if !strings.Contains(output, "C string") {
		t.Errorf("expected 'C string', got:\n%s", output)
	}
}

func TestPrintFieldsToon_NoGroupingDifferentTypes(t *testing.T) {
	fields := []Field{
		{Name: "X", Type: "int"},
		{Name: "Y", Type: "string"},
		{Name: "Z", Type: "bool"},
	}
	var sb strings.Builder
	printFieldsToon(fields, &sb)
	output := sb.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d:\n%s", len(lines), output)
	}
}

func TestPrintFieldsToon_Empty(t *testing.T) {
	var sb strings.Builder
	printFieldsToon(nil, &sb)
	if sb.Len() != 0 {
		t.Error("expected empty output for nil fields")
	}
}

func TestPrintFieldsToon_SingleField(t *testing.T) {
	fields := []Field{{Name: "Solo", Type: "float64"}}
	var sb strings.Builder
	printFieldsToon(fields, &sb)
	if !strings.Contains(sb.String(), "Solo float64") {
		t.Errorf("expected 'Solo float64', got: %s", sb.String())
	}
}

// --- printMethodsToon tests ---

func TestPrintMethodsToon_PointerReceiver(t *testing.T) {
	methods := []Method{
		{Receiver: "*MyStruct", Name: "DoThing", Signature: "DoThing(x int) error"},
	}
	var sb strings.Builder
	printMethodsToon(methods, &sb)
	output := sb.String()
	if !strings.Contains(output, "*DoThing(x int) error") {
		t.Errorf("expected pointer prefix '*', got:\n%s", output)
	}
}

func TestPrintMethodsToon_ValueReceiver(t *testing.T) {
	methods := []Method{
		{Receiver: "MyStruct", Name: "GetName", Signature: "GetName() string"},
	}
	var sb strings.Builder
	printMethodsToon(methods, &sb)
	output := sb.String()
	if strings.HasPrefix(output, "*") {
		t.Errorf("value receiver should not have * prefix, got:\n%s", output)
	}
	if !strings.Contains(output, "GetName() string") {
		t.Errorf("expected 'GetName() string', got:\n%s", output)
	}
}

func TestPrintMethodsToon_Empty(t *testing.T) {
	var sb strings.Builder
	printMethodsToon(nil, &sb)
	if sb.Len() != 0 {
		t.Error("expected empty output for nil methods")
	}
}

// --- filterFields / filterMethods tests ---

func TestFilterFields_NoRules(t *testing.T) {
	fields := []Field{{Name: "A", Type: "int"}, {Name: "B", Type: "string"}}
	result := filterFields(fields, nil)
	if len(result) != 2 {
		t.Errorf("expected 2 fields with no rules, got %d", len(result))
	}
}

func TestFilterMethods_NoRules(t *testing.T) {
	methods := []Method{{Name: "A"}, {Name: "B"}}
	result := filterMethods(methods, nil)
	if len(result) != 2 {
		t.Errorf("expected 2 methods with no rules, got %d", len(result))
	}
}

// --- printVarsToon tests ---

func TestPrintVarsToon_GroupsSameType(t *testing.T) {
	vars := []Variable{
		{Name: "a", Type: "int"},
		{Name: "b", Type: "int"},
		{Name: "c", Type: "string"},
	}
	var sb strings.Builder
	printVarsToon(vars, false, &sb)
	output := sb.String()
	if !strings.Contains(output, "var a, b int") {
		t.Errorf("expected grouped 'var a, b int', got:\n%s", output)
	}
	if !strings.Contains(output, "var c string") {
		t.Errorf("expected 'var c string', got:\n%s", output)
	}
}

func TestPrintVarsToon_WithComments(t *testing.T) {
	vars := []Variable{
		{Name: "x", Type: "int", Docs: []string{"x doc"}},
	}
	var sb strings.Builder
	printVarsToon(vars, true, &sb)
	output := sb.String()
	if !strings.Contains(output, "// x doc") {
		t.Errorf("expected comment, got:\n%s", output)
	}
}

// --- printConstsToon tests ---

func TestPrintConstsToon_WithValue(t *testing.T) {
	consts := []Constant{
		{Name: "MaxSize", Value: "100"},
	}
	var sb strings.Builder
	printConstsToon(consts, false, &sb)
	output := sb.String()
	if !strings.Contains(output, `const MaxSize = 100`) {
		t.Errorf("expected 'const MaxSize = 100', got:\n%s", output)
	}
}

func TestPrintConstsToon_WithoutValue(t *testing.T) {
	consts := []Constant{
		{Name: "Iota"},
	}
	var sb strings.Builder
	printConstsToon(consts, false, &sb)
	output := sb.String()
	if !strings.Contains(output, "const Iota") {
		t.Errorf("expected 'const Iota', got:\n%s", output)
	}
	if strings.Contains(output, "=") {
		t.Error("should not contain '=' for empty value")
	}
}

func TestPrintConstsToon_WithComments(t *testing.T) {
	consts := []Constant{
		{Name: "Pi", Value: "3.14", Docs: []string{"mathematical constant"}},
	}
	var sb strings.Builder
	printConstsToon(consts, true, &sb)
	output := sb.String()
	if !strings.Contains(output, "// mathematical constant") {
		t.Errorf("expected comment, got:\n%s", output)
	}
}

// --- DescribeType with interface ---

func TestDescribeType_Interface(t *testing.T) {
	iface := Interface{
		Name: "MyIface",
		Methods: []Method{
			{
				Name:   "Do",
				Params: []Param{{TypeDetails: TypeDetails{Type: strPtr("Config")}}},
			},
		},
	}
	config := Struct{Name: "Config", Fields: []Field{{Name: "Val", Type: "string"}}}
	pkg := Package{
		Package:    "test",
		Interfaces: []Interface{iface},
		Structs:    []Struct{config},
	}
	parsed := []*ParsedInfo{{Packages: []Package{pkg}}}

	result, err := DescribeType("MyIface", parsed, 1)
	if err != nil {
		t.Fatalf("DescribeType failed: %v", err)
	}

	ifaceNames := collectInterfaceNames(result)
	structNames := collectStructNames(result)
	if len(ifaceNames) != 1 || ifaceNames[0] != "MyIface" {
		t.Errorf("expected [MyIface], got %v", ifaceNames)
	}
	if len(structNames) != 1 || structNames[0] != "Config" {
		t.Errorf("expected [Config], got %v", structNames)
	}
}

// helpers

func strPtr(s string) *string {
	return &s
}

func collectTypeDefNames(parsed []*ParsedInfo) []string {
	var names []string
	for _, p := range parsed {
		for _, pkg := range p.Packages {
			for _, td := range pkg.TypeDefs {
				names = append(names, td.Name)
			}
		}
	}
	return names
}

func collectStructNames(parsed []*ParsedInfo) []string {
	var names []string
	for _, p := range parsed {
		for _, pkg := range p.Packages {
			for _, s := range pkg.Structs {
				names = append(names, s.Name)
			}
		}
	}
	return names
}

func collectInterfaceNames(parsed []*ParsedInfo) []string {
	var names []string
	for _, p := range parsed {
		for _, pkg := range p.Packages {
			for _, i := range pkg.Interfaces {
				names = append(names, i.Name)
			}
		}
	}
	return names
}

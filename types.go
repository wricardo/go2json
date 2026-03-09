package go2json

// ParsedInfo holds parsed information about Go packages.
type ParsedInfo struct {
	Modules   []Module  `json:"modules"`
	Packages  []Package `json:"packages"`  // Deprecated: use Modules instead
	Directory string    `json:"directory"` // if information was parsed from a directory. It's either a directory or a file
	File      string    `json:"file"`      // if information was parsed from a single file. It's either a directory or a file
}

// Module represents a Go module with its packages.
type Module struct {
	RootModuleName    string    `json:"root_module_name"`   // Name of the module as seen on the go.mod file of the project
	RelativeDirectory string    `json:"relative_directory"` // Relative (to go.mod) directory of the module / or /cmd/something
	FullName          string    `json:"full_name"`          // Full name of the module, including the relative directory should be RootModuleName/RelativeDirectory
	Packages          []Package `json:"packages"`
}

// Package represents a Go package with its components such as imports, structs, functions, etc.
type Package struct {
	Package    string      `json:"package"`     // Name of the package as seen in the package declaration (e.g., "main")
	ModuleName string      `json:"module_name"` // Name of the module as seen in the go.mod file
	Imports    []Import    `json:"imports,omitempty"`
	Structs    []Struct    `json:"structs,omitempty"`
	Functions  []Function  `json:"functions,omitempty"`
	Variables  []Variable  `json:"variables,omitempty"`
	Constants  []Constant  `json:"constants,omitempty"`
	Interfaces []Interface `json:"interfaces,omitempty"`
	TypeDefs   []TypeDef   `json:"type_defs,omitempty"`

	PtrModule *Module `json:"-"` // Pointer to the module that this package belongs to
}

// Import represents a Go import declaration.
type Import struct {
	Name string `json:"name,omitempty"` // the alias of the package as it's being imported
	Path string `json:"path"`

	PtrPackage *Package `json:"-"` // Pointer to the package that this import belongs to
}

// Struct represents a Go struct and its fields and methods.
type Struct struct {
	Name       string   `json:"name"`
	Fields     []Field  `json:"fields,omitempty"`
	Methods    []Method `json:"methods,omitempty"`
	Docs       []string `json:"docs,omitempty"`
	Definition string   `json:"definition,omitempty"` // Full Go code definition of the struct
	IsExported bool     `json:"is_exported"`          // Whether the struct is exported (public)

	PtrPackage *Package `json:"-"` // Pointer to the package that this struct belongs to
}

// Interface represents a Go interface and its methods.
type Interface struct {
	Name       string   `json:"name"`
	Methods    []Method `json:"methods,omitempty"`
	Docs       []string `json:"docs,omitempty"`
	Definition string   `json:"definition,omitempty"` // Full Go code definition of the interface
	IsExported bool     `json:"is_exported"`          // Whether the interface is exported (public)

	PtrPackage *Package `json:"-"` // Pointer to the package that this interface belongs to
}

// Method represents a method in a Go struct or interface.
type Method struct {
	Receiver    string   `json:"receiver,omitempty"` // Receiver type (e.g., "*MyStruct" or "MyStruct")
	Name        string   `json:"name"`
	Params      []Param  `json:"params,omitempty"`
	Returns     []Param  `json:"returns,omitempty"`
	Docs        []string `json:"docs,omitempty"`
	Signature   string   `json:"signature"`
	Body        string   `json:"body,omitempty"`       // New field for method body
	Definition  string   `json:"definition,omitempty"` // Full Go code definition of the method
	IsExported  bool     `json:"is_exported"`          // Whether the method is exported (public)
	IsTest      bool     `json:"is_test"`              // Whether the method is a test method
	IsBenchmark bool     `json:"is_benchmark"`         // Whether the method is a benchmark method

	PtrStruct *Struct `json:"-"` // Pointer to the struct that this method belongs to
}

// Function represents a Go function with its parameters, return types, and documentation.
type Function struct {
	Name        string   `json:"name"`
	Params      []Param  `json:"params,omitempty"`
	Returns     []Param  `json:"returns,omitempty"`
	Docs        []string `json:"docs,omitempty"`
	Signature   string   `json:"signature"`
	Body        string   `json:"body,omitempty"`       // New field for function body
	Definition  string   `json:"definition,omitempty"` // Full Go code definition of the function
	IsExported  bool     `json:"is_exported"`          // Whether the function is exported (public)
	IsTest      bool     `json:"is_test"`              // Whether the function is a test function (TestXxx)
	IsBenchmark bool     `json:"is_benchmark"`         // Whether the function is a benchmark function (BenchmarkXxx)
}

// Param represents a parameter or return value in a Go function or method.
type Param struct {
	Name        string      `json:"name"` // Name of the parameter or return value
	Type        string      `json:"type"` // Type (e.g., "int", "*string")
	TypeDetails TypeDetails `json:"type_details"`

	PtrMethod *Method   `json:"-"` // Pointer to the method that this parameter belongs to
	PtrFunc   *Function `json:"-"` // Pointer to the function that this parameter belongs to
}

// Field represents a field in a Go struct.
type Field struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	TypeDetails TypeDetails `json:"type_details"`
	Tag         string      `json:"tag,omitempty"`
	Private     bool        `json:"private"`
	Pointer     bool        `json:"pointer"`
	Slice       bool        `json:"slice"`
	Docs        []string    `json:"docs,omitempty"`
	Comment     string      `json:"comment,omitempty"`

	PtrStruct *Struct `json:"-"` // Pointer to the struct that this field belongs to
}

// TypeDef represents a named type declaration that is neither a struct nor an interface
// (e.g., "type SomeFunc func(a string) error" or "type SpecialString string").
type TypeDef struct {
	Name       string   `json:"name"`
	Underlying string   `json:"underlying"` // The underlying type expression as a string
	Docs       []string `json:"docs,omitempty"`
	Definition string   `json:"definition,omitempty"` // Full Go code definition
	IsExported bool     `json:"is_exported"`

	PtrPackage *Package `json:"-"`
}

// TypeDetails holds comprehensive metadata about a Go type expression.
type TypeDetails struct {
	Package     *string // in the cases of external types. like pbhire.Person this would be "github.com/x/y/pbhire"
	PackageName *string // in the cases of external types. like pbhire.Person this would be "pbhire"
	Type        *string // in the cases of external types. like pbhire.Person this would be "Person"

	TypeName       string
	IsPointer      bool
	IsSlice        bool
	IsMap          bool
	IsBuiltin      bool // if string, int, etc
	IsExternal     bool // if the type is from another package
	TypeReferences []TypeReference
}

// TypeReference holds metadata about a referenced type.
type TypeReference struct {
	Package     *string
	PackageName *string
	Name        string // the name of the type, example TypeReference or Person or int32 or string
}

// Variable represents a global variable in a Go package.
type Variable struct {
	Name string   `json:"name"`
	Type string   `json:"type"`
	Docs []string `json:"docs,omitempty"`
}

// Constant represents a constant in a Go package.
type Constant struct {
	Name  string   `json:"name"`
	Value string   `json:"value"`
	Docs  []string `json:"docs,omitempty"`
}

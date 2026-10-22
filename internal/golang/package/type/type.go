package _type

type ModuleChange int

const (
	CreateModuleChange ModuleChange = iota
	UpdateModuleChange
	DeleteModuleChange
)

type (
	Import              string // `"fmt"` or `alias "github.com/x/y"`
	VariableDefinition  string // `var x = 1` / `const y = "z"`
	FuncDefinition      string // full func incl. receiver + body
	TypeDefinition      string // full type decl (non-interface)
	InterfaceDefinition string // full `type X interface { ... }`
)

type FeatureMap map[string]string // free-form tags: "build":"linux", "generated":"true", ...

type File interface {
	Name() string
	Description() string // file-level doc comment
	Package() Package
	Imports() []Import
	Variables() []VariableDefinition
	Funcs() []FuncDefinition
	Types() []TypeDefinition
	Interfaces() []InterfaceDefinition
}

type Package interface {
	Name() string
	Path() string         // import path
	Description() string  // package doc comment from ak.md
	Features() FeatureMap // package features list from ak.md
	Files() []File
}

type Match struct {
	Type       TypeDefinition
	Name       string
	PointerRcv bool
}

type Reader interface {
	// Read loads the package at an import path or directory.
	Read(path string) (Package, error)
	// Implementations returns the types in the package at path that satisfy iface.
	Implementations(path string, iface InterfaceDefinition) ([]Match, error)
}

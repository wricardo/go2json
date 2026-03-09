package go2json

import "fmt"

// builtinTypes is the set of Go builtin types excluded from graph traversal.
var builtinTypes = map[string]bool{
	"string": true, "int": true, "int8": true, "int16": true,
	"int32": true, "int64": true, "uint": true, "uint8": true,
	"uint16": true, "uint32": true, "uint64": true, "uintptr": true,
	"float32": true, "float64": true, "complex64": true, "complex128": true,
	"bool": true, "byte": true, "rune": true, "error": true, "any": true,
}

type typeEntry struct {
	Kind      string // "struct", "interface", or "typedef"
	Struct    *Struct
	Interface *Interface
	TypeDef   *TypeDef
	Package   *Package
}

// BuildTypeIndex indexes every struct, interface, and named type by unqualified
// name across all parsed packages.
func BuildTypeIndex(parsed []*ParsedInfo) map[string]typeEntry {
	index := make(map[string]typeEntry)
	for _, p := range parsed {
		for i := range p.Packages {
			pkg := &p.Packages[i]
			for j := range pkg.Structs {
				s := &pkg.Structs[j]
				if _, exists := index[s.Name]; !exists {
					index[s.Name] = typeEntry{
						Kind:    "struct",
						Struct:  s,
						Package: pkg,
					}
				}
			}
			for j := range pkg.Interfaces {
				iface := &pkg.Interfaces[j]
				if _, exists := index[iface.Name]; !exists {
					index[iface.Name] = typeEntry{
						Kind:      "interface",
						Interface: iface,
						Package:   pkg,
					}
				}
			}
			for j := range pkg.TypeDefs {
				td := &pkg.TypeDefs[j]
				if _, exists := index[td.Name]; !exists {
					index[td.Name] = typeEntry{
						Kind:    "typedef",
						TypeDef: td,
						Package: pkg,
					}
				}
			}
		}
	}
	return index
}

// extractTypeNames returns the non-builtin type names referenced by a TypeDetails.
func extractTypeNames(td TypeDetails) []string {
	if len(td.TypeReferences) > 0 {
		var names []string
		for _, ref := range td.TypeReferences {
			if !builtinTypes[ref.Name] {
				names = append(names, ref.Name)
			}
		}
		return names
	}
	if td.Type != nil && !builtinTypes[*td.Type] {
		return []string{*td.Type}
	}
	return nil
}

// referencedTypes collects all non-builtin type names referenced by a type entry
// (from struct fields, method params, and method returns).
func referencedTypes(entry typeEntry) []string {
	var refs []string
	seen := make(map[string]bool)
	add := func(names []string) {
		for _, n := range names {
			if !seen[n] {
				seen[n] = true
				refs = append(refs, n)
			}
		}
	}

	addFromMethods := func(methods []Method) {
		for _, m := range methods {
			for _, p := range m.Params {
				add(extractTypeNames(p.TypeDetails))
			}
			for _, r := range m.Returns {
				add(extractTypeNames(r.TypeDetails))
			}
		}
	}

	switch entry.Kind {
	case "struct":
		if entry.Struct != nil {
			for _, f := range entry.Struct.Fields {
				add(extractTypeNames(f.TypeDetails))
			}
			addFromMethods(entry.Struct.Methods)
		}
	case "interface":
		if entry.Interface != nil {
			addFromMethods(entry.Interface.Methods)
		}
	case "typedef":
		// TypeDefs are leaf nodes — they don't reference other types in the index
		// (their underlying type is a string, not parsed into TypeDetails)
	}
	return refs
}

type bfsItem struct {
	name  string
	depth int
}

// DescribeType performs a BFS from typeName through field and method type
// references, bounded by maxDepth. It returns synthetic []*ParsedInfo containing
// only the discovered types, grouped by package.
func DescribeType(typeName string, parsed []*ParsedInfo, maxDepth int) ([]*ParsedInfo, error) {
	index := BuildTypeIndex(parsed)

	if _, ok := index[typeName]; !ok {
		return nil, fmt.Errorf("type %q not found", typeName)
	}

	visited := make(map[string]bool)
	queue := []bfsItem{{name: typeName, depth: 0}}
	visited[typeName] = true

	// Collect found entries preserving discovery order
	var found []typeEntry
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		entry, ok := index[item.name]
		if !ok {
			continue
		}
		found = append(found, entry)

		if item.depth < maxDepth {
			for _, ref := range referencedTypes(entry) {
				if !visited[ref] && !builtinTypes[ref] {
					visited[ref] = true
					queue = append(queue, bfsItem{name: ref, depth: item.depth + 1})
				}
			}
		}
	}

	// Group by package pointer → build synthetic ParsedInfo
	pkgOrder := []*Package{}
	pkgSeen := make(map[*Package]bool)
	pkgStructs := make(map[*Package][]Struct)
	pkgInterfaces := make(map[*Package][]Interface)
	pkgTypeDefs := make(map[*Package][]TypeDef)

	for _, entry := range found {
		pkg := entry.Package
		if !pkgSeen[pkg] {
			pkgSeen[pkg] = true
			pkgOrder = append(pkgOrder, pkg)
		}
		switch entry.Kind {
		case "struct":
			pkgStructs[pkg] = append(pkgStructs[pkg], *entry.Struct)
		case "interface":
			pkgInterfaces[pkg] = append(pkgInterfaces[pkg], *entry.Interface)
		case "typedef":
			pkgTypeDefs[pkg] = append(pkgTypeDefs[pkg], *entry.TypeDef)
		}
	}

	var results []*ParsedInfo
	for _, pkg := range pkgOrder {
		synPkg := Package{
			Package:    pkg.Package,
			ModuleName: pkg.ModuleName,
			Structs:    pkgStructs[pkg],
			Interfaces: pkgInterfaces[pkg],
			TypeDefs:   pkgTypeDefs[pkg],
		}
		results = append(results, &ParsedInfo{
			Packages: []Package{synPkg},
		})
	}
	return results, nil
}

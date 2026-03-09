package main

import (
	"fmt"

	"github.com/wricardo/go2json"
)

type INode interface {
	GetID() string
	SetID(id string)
	GetLabel() string
	GetProps() map[string]any
	SetProps(props map[string]any)
}

func main() {
	tmp, err := go2json.ParseFile("other.go")
	if err != nil {
		fmt.Println(err)
		return
	}
	pkg := tmp.Packages[0]
	fragment := make([]go2json.CodeFragment, 0)
	for _, s := range pkg.Structs {
		fragment = append(fragment, go2json.CodeFragment{
			Content: go2json.RenderTemplateNoError(`
			func (n {{.Name}}) GetID() string {
				return n.Id
			}

			func (n {{.Name}}) SetID(id string) {
				n.Id = id
			}	

			func (n {{.Name}}) GetLabel() string {
				return "{{.Name}}"
			}

			func (n {{.Name}}) GetProps() map[string]any {
				return map[string]any{
				}
			}

			func (n {{.Name}}) SetProps(props map[string]any) {
				if val, ok := props["SomeField"]; ok {
					n.SomeField = val.(string)
				}
			}




			`, s),

			Overwrite: false,
		})
	}

	fragments := map[string][]go2json.CodeFragment{
		"other.go": fragment,
	}

	go2json.InsertCodeFragments(fragments)
}

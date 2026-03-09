package main

import (
	"github.com/wricardo/go2json"
)

func main() {
	smsProviders := []struct {
		Id   int
		Name string
	}{
		{1, "Twilio"},
		{2, "Nexmo"},
	}

	fragments := map[string][]go2json.CodeFragment{
		"dynamic.go": {
			go2json.CodeFragment{
				Content: go2json.RenderTemplateNoError(`
					{{- range .}}
					type {{.Name}}Provider struct {
						Id   int
						Name string
					}

					{{- end}}
				`, smsProviders),
				Overwrite: true,
			},
			go2json.CodeFragment{
				Content: go2json.RenderTemplateNoError(`
				{{- range .}}
				func (p *{{.Name}}Provider) SendSMS(to, body string) error {
					fmt.Printf("Sending SMS to %s using %s provider\n", to, p.Name)
					return nil
				}
				{{- end}}
				`, smsProviders),
				Overwrite: false,
			},
			go2json.CodeFragment{
				Content: `type Person struct{
					Name string
				}`,
				Overwrite: false,
			},
		},
	}

	go2json.InsertCodeFragments(fragments)
}

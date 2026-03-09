package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

	changes := []go2json.FileChange{
		{
			PackageName: "main",
			File:        "assets/dynamic.go",
			Fragments: []go2json.CodeFragment{
				{
					Content: go2json.RenderTemplateNoError(`
			import "fmt"
			{{- range .}}
			type {{.Name}}Provider struct {
				Id   int
				Name string
			}

			func (p *{{.Name}}Provider) SendSMS(to, body string) error {
				fmt.Printf("Sending SMS to %s using %s provider\n", to, p.Name)
				return nil
			}
			{{- end}}
			`, smsProviders),
					Overwrite: true,
				},
				{
					Content: `type Person struct{
					Name string
				}`,
					Overwrite: false,
				},
			},
		},
	}
	jsonChanges, _ := json.Marshal(changes)
	fmt.Println(strings.Replace(string(jsonChanges), "\\t", "", -1))

	go2json.ApplyFileChanges(changes)
}

// Get user input for request
func getUserRequest() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your modification request: ")
	request, _ := reader.ReadString('\n')
	return strings.TrimSpace(request)
}

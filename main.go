package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var testData = `
apiVersion: v1
kind: Namespace
metadata:
  labels:
    cloudbees-sidecar-injector: enabled
  name: jenkins-agents

`

type Metadata struct {
	Name      string                 `yaml:"name"`
	Namespace string                 `yaml:"namespace,omitempty"`
	Labels    map[string]interface{} `yaml:"labels,omitempty"`
}

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.
type Doc struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
}

func NewDoc(data map[string]interface{}) Doc {
	doc := Doc{}
	if ver, ok := data["ApiVersion"]; ok {
		doc.ApiVersion = ver.(string)
	}
	if kind, ok := data["Kind"]; ok {
		doc.Kind = kind.(string)
	}
	doc.Metadata = Metadata{}

	metaI, ok := data["metadata"]
	if ok {

		meta, ok := metaI.(map[string]interface{})
		if !ok {
			log.Fatalf("metadata wrong")
		}

		for k, v := range meta {
			if k == "name" {
				doc.Metadata.Name = v.(string)
			}
			if k == "namespace" {
				doc.Metadata.Namespace = v.(string)
			}
			if k == "label" {
				doc.Metadata.Labels = v.(map[string]interface{})
			}
		}
	}

	return doc
}

func SkipDoc(filter *Doc, doc Doc, search string, docStr string) bool {
	log.Printf("filter %v", filter)

	if search != "" && strings.Contains(docStr, search) {
		return false
	}
	log.Printf("filter nil %v", filter)
	if filter != nil {
		if filter.Kind != "" && doc.Kind != filter.Kind {
			return true
		}
		if filter.Metadata.Name != "" && doc.Metadata.Name != filter.Metadata.Name {
			return true
		}
		if filter.Metadata.Namespace != "" && doc.Metadata.Namespace != filter.Metadata.Namespace {
			return true
		}
	}

	return false
}

func ParseDoc(inFile string, filter *Doc, search string) {
	data, err := os.ReadFile(inFile)
	if err != nil {
		log.Fatal(err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var data map[string]interface{}
		err := decoder.Decode(&data)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		doc := NewDoc(data)
		d, err := yaml.Marshal(&data)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if SkipDoc(filter, doc, search, string(d)) {
			continue
		}

		fmt.Printf("--- yaml dump:\n%s\n\n", string(d))
	}
}

func TestExample() {
	doc := Doc{}

	err := yaml.Unmarshal([]byte(testData), &doc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- doc:\n%v\n\n", doc)

	d, err := yaml.Marshal(&doc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal([]byte(testData), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)

	d, err = yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))
}

func main() {

	var fileIn, name, ns, docKind, search string
	flag.StringVar(&fileIn, "file", "", "yaml file")
	flag.StringVar(&name, "name", "", "name to filter")
	flag.StringVar(&ns, "ns", "", "namespace to filter")
	flag.StringVar(&docKind, "kind", "", "kind to filter")
	flag.StringVar(&search, "search", "", "search string to filter")

	flag.Parse()

	if fileIn != "" {
		filter := &Doc{Kind: docKind, Metadata: Metadata{Name: name, Namespace: ns}}
		if docKind == "" && name == "" && ns == "" {
			filter = nil
		}
		ParseDoc(fileIn, filter, search)
	}

}

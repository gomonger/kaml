package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var testData = `
apiVersion: v1
kind: Namespace
metadata:
  name: foobar
`

type Metadata struct {
	Name      string                      `yaml:"name"`
	Namespace string                      `yaml:"namespace,omitempty"`
	Labels    map[interface{}]interface{} `yaml:"labels,omitempty"`
}

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.
type Doc struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
}

func ParseDoc(inFile string, filter *Doc) {
	data, err := os.ReadFile(inFile)
	if err != nil {
		log.Fatal(err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		doc := Doc{}
		err := decoder.Decode(&doc)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		if filter != nil {
			if filter.Kind != "" && doc.Kind != filter.Kind {
				continue
			}
			if filter.Metadata.Name != "" && doc.Metadata.Name != filter.Metadata.Name {
				continue
			}
			if filter.Metadata.Namespace != "" && doc.Metadata.Namespace != filter.Metadata.Namespace {
				continue
			}
		}
		fmt.Println(doc)
		d, err := yaml.Marshal(&doc)
		if err != nil {
			log.Fatalf("error: %v", err)
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

	var fileIn, name, ns, docKind string
	flag.StringVar(&fileIn, "file", "", "yaml file")
	flag.StringVar(&name, "name", "", "name to filter")
	flag.StringVar(&ns, "ns", "", "namespace to filter")
	flag.StringVar(&docKind, "kind", "", "kind to filter")
	flag.Parse()

	if fileIn != "" {
		filter := &Doc{Kind: docKind, Metadata: Metadata{Name: name, Namespace: ns}}
		if docKind == "" && name == "" && ns == "" {
			filter = nil
		}
		ParseDoc(fileIn, filter)
	}

}

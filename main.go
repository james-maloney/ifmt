package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	isXml   bool
	noColor bool
	indent  string
)

func init() {
	flag.BoolVar(&isXml, "xml", false, "format XML")
	flag.BoolVar(&noColor, "no-color", false, "no color")
	flag.StringVar(&indent, "indent", "\t", "indent")
}

func main() {
	flag.Parse()

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		if isXml {
			b := formatXml(s.Bytes())
			fmt.Println(string(b))
		} else {
			sc := &scanner{
				color:  !noColor,
				indent: indent,
			}
			if err := sc.parse(s.Bytes()); err != nil {
				log.Fatal(err)
			}
			fmt.Println(sc.buf.String())
		}
	}

	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
}

func formatXml(b []byte) []byte {
	var mi map[string]interface{}
	if err := xml.Unmarshal(b, &mi); err != nil {
		log.Fatal(err)
	}

	bt, err := xml.MarshalIndent(mi, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	return bt
}

func formatJson(b []byte) []byte {
	var mi map[string]interface{}
	if err := json.Unmarshal(b, &mi); err != nil {
		log.Fatal(err)
	}

	bt, err := json.MarshalIndent(mi, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	return bt
}

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

var isXml bool

func init() {
	flag.BoolVar(&isXml, "xml", false, "format XML")
}

func main() {
	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		if isXml {
			b := formatXml(s.Bytes())
			fmt.Println(string(b))
		} else {
			b := formatJson(s.Bytes())
			s := &scanner{}
			if err := s.parse(b); err != nil {
				log.Fatal(err)
			}
			fmt.Println(s.buf.String())
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

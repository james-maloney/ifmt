package main

import (
	"fmt"

	"testing"
)

func Test(t *testing.T) {
	cases := []struct {
		val []byte
	}{{
		[]byte(`
{
	"foo": "bar"
}
`),
	}}
	for _, cs := range cases {
		s := &scanner{}
		if err := s.parse(cs.val); err != nil {
			t.Fatal(err)
		}
		fmt.Println(s.data)
	}
}

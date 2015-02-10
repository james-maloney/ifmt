package main

import (
	"fmt"

	"testing"
)

func TestValueStart(t *testing.T) {
	cases := []struct {
		value rune
		pass  bool
		state int
	}{
		{'{', true, keyState},
		{'[', true, 0},
		{'"', true, 0},
		{'0', true, 0},
		{'-', true, 0},
		{'1', true, 0},
		{'2', true, 0},
		{'3', true, 0},
		{'4', true, 0},
		{'5', true, 0},
		{'6', true, 0},
		{'7', true, 0},
		{'8', true, 0},
		{'9', true, 0},
		{'t', true, 0},
		{'f', true, 0},
		{'n', true, 0},
		{' ', true, 0},
		{'\t', true, 0},
	}
	for _, cs := range cases {
		s := &scanner{}
		err := valueStart(s, cs.value)
		if cs.pass && err != nil {
			t.Error(err)
		} else {
			if cs.state > 0 {
				if cs.state != s.state[len(s.state)-1] {
					t.Errorf("\tinvalid state: got: %v should: %v", s.state, cs.state)
				}
			}
		}
		if !cs.pass && err == nil {
			t.Error("should not have passed")
		}
	}
}

func TestKeyStart(t *testing.T) {
	cases := []struct {
		name  string
		value rune
		pass  bool
		state int
	}{
		{"key start", '"', true, keyState},
		{"empty space", ' ', true, keyState},
		{"tab rune", '\t', true, keyState},
		{"invalid rune", 'x', false, 0},
	}
	for _, cs := range cases {
		t.Log("case:", cs.name)
		s := &scanner{
			state: []int{objState, keyState},
		}
		err := keyStart(s, cs.value)
		if cs.pass && err != nil {
			t.Error("\t", err)
		} else if cs.pass {
			if cs.state != s.state[len(s.state)-1] {
				t.Errorf("\tinvalid state: got: %v should: %v", s.state, cs.state)
			}
		}
		if !cs.pass && err == nil {
			t.Error("should not have passed")
		}
	}
}

func TestKeyEnd(t *testing.T) {
	s := &scanner{}
	cases := []struct {
		name  string
		value rune
		pass  bool
		state int
		reset bool
	}{
		{"key end", '"', true, objState, true},
		{"escape test", '\\', true, skipState, true},
		{"escape test line two", '"', true, keyState, false},
		{"escape test line three", '"', true, objState, false},
	}
	for _, cs := range cases {
		if cs.reset {
			s.state = []int{objState, keyState}
		}
		t.Log("case:", cs.name)
		err := keyEnd(s, cs.value)
		if cs.pass && err != nil {
			t.Error("\t", err)
		} else if cs.pass {
			if cs.state != s.state[len(s.state)-1] {
				t.Errorf("\tinvalid state: got: %v should: %v", s.state, cs.state)
			}
		}
		if !cs.pass && err == nil {
			t.Error("\tshould not have passed")
		}
	}
}

func TestColon(t *testing.T) {
	s := &scanner{}
	cases := []struct {
		name  string
		value rune
		pass  bool
		state int
	}{
		{"has space", ' ', true, objState},
		{"has \\t", '\t', true, objState},
		{"has ':'", ':', true, objState},
		{"invalid rune", 'x', false, 0},
	}
	for _, cs := range cases {
		s.state = []int{objState}
		t.Log("case:", cs.name)
		err := colon(s, cs.value)
		if cs.pass && err != nil {
			t.Error("\t", err)
		} else if cs.pass {
			if cs.state != s.state[len(s.state)-1] {
				t.Errorf("\tinvalid state: got: %v should: %v", s.state, cs.state)
			}

		}
	}
}

func TestStrValueStart(t *testing.T) {
	s := &scanner{}
	cases := []struct {
		name       string
		value      rune
		pass       bool
		startState int
		state      int
		reset      bool
	}{
		{"value end", '"', true, objState, objState, true},
		{"value end", '"', true, arrState, arrState, true},
		{"escape test", '\\', true, arrState, skipState, true},
		{"escape test line two", '"', true, 0, strValueState, false},
		{"escape test line three", '"', true, 0, arrState, false},
	}
	for _, cs := range cases {
		if cs.reset {
			s.state = []int{cs.startState, strValueState}
		}
		t.Log("case:", cs.name)
		err := strValueStart(s, cs.value)
		if cs.pass && err != nil {
			t.Error("\t", err)
		} else if cs.pass {
			if cs.state != s.state[len(s.state)-1] {
				t.Errorf("\tinvalid state: got: %v should: %v", s.state, cs.state)
			}
		}
		if !cs.pass && err == nil {
			t.Error("\tshould not have passed")
		}
	}
}

func Test(t *testing.T) {
	cases := []struct {
		val []byte
	}{
		{
			[]byte(`
{
	"foo": "bar",
	"baz": "bang",
	"bang": {
		"fizz": "buzz"
	},
	"bust": ["bar", "bang", "bust"],
	"bustTrue": true,
	"bustFalse": false,
	"bustNull": null,
	"nullBoolArray": ["foo", {"foo":"bar"}, false, true, null],
	"num": 0,
	"num2": 0.1e10,
	"num3": -0.100,
	"num4": 10e1,
	"num5": 100.999,
	"numArr": [100, 200, -0.5, 0.5e10],
	"numObj": {
		"num": 100
	}
}
`),
		},
	}
	for _, cs := range cases {
		s := &scanner{}
		if err := s.parse(cs.val); err != nil {
			t.Error(err)
			t.Log(s.state)
			continue
		}
		fmt.Println(s.buf.String())
	}
}

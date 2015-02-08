package main

import (
	"fmt"
	ts "github.com/james-maloney/termstyle"
	"unicode"
)

const (
	objState = iota
	arrState
	strState
	keyState
	boolStart
	zeroStart
	numStart
	skip
)

var (
	objColor  = ts.FG[ts.Blue1]
	arrColor  = ts.FG[ts.RosyBrown]
	keyColor  = ts.FG[ts.System3]
	strColor  = ts.FG[ts.Green3]
	boolColor = ts.FG[ts.Red1]
	numColor  = ts.FG[ts.Cornsilk1]
)

/*
{
	"foo": "bar"
}
*/

type scanner struct {
	state []int
	data  string
	next  func(*scanner, rune) error
}

func (s *scanner) push(state int) {
	s.state = append(s.state, state)
}

func (s *scanner) popState() int {
	var state int
	s.state, state = s.state[:len(s.state)-1], s.state[len(s.state)-1]
	return state
}

func (s *scanner) addColor(color string, value string) {
	s.data += color + string(value)
}

func (s *scanner) parse(data []byte) error {
	s.next = valueStart

	for _, v := range data {
		if err := s.next(s, rune(v)); err != nil {
			return err
		}
	}

	if len(s.state) == 0 {
		return nil
	}

	return fmt.Errorf("JSON is incomplete")
}

func keyStart(s *scanner, r rune) error {
	if unicode.IsSpace(r) {
		s.data += string(r)
		return nil
	}
	if r == '}' {
		s.popState()
		s.data += objColor + string(r) + ts.C
		s.next = valueEnd
		return nil
	}
	if r != '"' {
		return fmt.Errorf("value should have started")
	}
	s.data += keyColor + string(r)
	s.push(keyState)
	s.next = keyEnd
	return nil
}

func keyEnd(s *scanner, r rune) error {
	s.data += string(r)
	if r == '\\' {
		s.push(skip)
		return nil
	}
	if s.state[len(s.state)-1] == skip {
		s.popState()
		return nil
	}
	if r != '"' {
		return nil
	}

	s.data += ts.C
	s.next = colon

	return nil
}

func colon(s *scanner, r rune) error {
	s.data += string(r)
	if unicode.IsSpace(r) {
		return nil
	}
	if r == ':' {
		s.popState()
		s.next = valueStart
		return nil
	}
	return fmt.Errorf("should of seen ':'")
}

func strValue(s *scanner, r rune) error {
	s.data += string(r)
	if r == '\\' {
		s.push(skip)
		return nil
	}
	if s.state[len(s.state)-1] == skip {
		s.popState()
		return nil
	}
	if r == '"' {
		s.data += ts.C
		s.next = valueEnd
	}
	return nil
}

func valueEnd(s *scanner, r rune) error {
	if unicode.IsSpace(r) {
		s.data += string(r)
		return nil
	}
	if r == ',' {
		s.data += string(r)
		s.next = valueStart
		return nil
	}
	if r == '}' {
		s.data += objColor + string(r) + ts.C
		s.popState()
		s.next = valueStart
		return nil
	}
	if len(s.state) == 0 {
		return nil
	}

	return fmt.Errorf("should not be here")
}

func valueStart(s *scanner, r rune) error {
	switch r {
	case '{':
		s.data += objColor + string(r) + ts.C
		s.push(objState)
		s.next = keyStart
		return nil
	case '[':
	case '"':
		s.data += strColor + string(r)
		s.next = strValue
		return nil
	case 'f':
	case 't':
	case '0':
	case '-':
	}

	if r >= '1' || r <= '9' {

	}

	return nil
}

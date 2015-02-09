package main

import (
	"bytes"
	"errors"
	"fmt"
	ts "github.com/james-maloney/termstyle"
	"unicode"
)

var eof = errors.New("eof")

const (
	_ = iota
	objState
	arrState
	keyState
	skipState
	strValueState
)

var (
	objWrap  = ts.FG[ts.Blue1]
	arrWrap  = ts.FG[ts.RosyBrown]
	keyWrap  = ts.FG[ts.System3]
	strWrap  = ts.FG[ts.Green3]
	boolWrap = ts.FG[ts.Red1]
	numWrap  = ts.FG[ts.Cornsilk1]
)

type scanner struct {
	state []int
	next  func(*scanner, rune) error
	ft    []rune // false/true helper state

	buf bytes.Buffer
}

func (s *scanner) pushState(state int) {
	s.state = append(s.state, state)
}

func (s *scanner) popState() int {
	var state int
	s.state, state = s.state[:len(s.state)-1], s.state[len(s.state)-1]
	return state
}

func (s *scanner) parse(data []byte) error {
	s.next = valueStart

	for _, v := range data {
		if err := s.next(s, rune(v)); err != nil {
			if err == eof {
				return nil
			}
			return err
		}
	}

	if len(s.state) == 0 {
		return nil
	}

	return fmt.Errorf("JSON is incomplete")
}

func valueStart(s *scanner, r rune) error {
	switch r {
	case '{':
		s.next = keyStart
		s.pushState(objState)
		s.buf.WriteString(objWrap)
		s.buf.WriteRune(r)
		s.buf.WriteString(ts.C)
		return nil
	case '[':
		return nil
	case '"':
		s.next = strValueStart
		s.pushState(strValueState)
		s.buf.WriteString(strWrap)
		s.buf.WriteRune(r)
		return nil
	case 't', 'f', 'n':
		return nil
	case '0':
		return nil
	case '-':
		return nil
	}
	if r >= '1' && r <= '9' {
		return nil
	}
	if unicode.IsSpace(r) {
		return nil
	}

	return fmt.Errorf("invalid value '%v'", string(r))
}

func keyStart(s *scanner, r rune) error {
	if r == '"' {
		s.next = keyEnd
		s.pushState(keyState)
		s.buf.WriteString(keyWrap)
		s.buf.WriteRune(r)
		return nil
	}
	if unicode.IsSpace(r) {
		s.buf.WriteRune(r)
		return nil
	}

	return fmt.Errorf("invalid string value '%v'", string(r))
}

func keyEnd(s *scanner, r rune) error {
	s.buf.WriteRune(r)
	if s.state[len(s.state)-1] == skipState {
		s.popState()
		return nil
	}
	if r == '\\' {
		s.pushState(skipState)
		return nil
	}
	if r == '"' {
		s.next = colon
		s.popState()
		s.buf.WriteString(ts.C)
		return nil
	}

	return nil
}

func colon(s *scanner, r rune) error {
	s.buf.WriteRune(r)
	if r == ':' {
		s.next = valueStart
		return nil
	}
	if unicode.IsSpace(r) {
		return nil
	}

	return fmt.Errorf("invalid rune '%v' looking for ':'", string(r))
}

func strValueStart(s *scanner, r rune) error {
	s.buf.WriteRune(r)
	if s.state[len(s.state)-1] == skipState {
		s.popState()
		return nil
	}
	if r == '\\' {
		s.pushState(skipState)
		return nil
	}
	if r == '"' {
		s.popState()
		s.next = valueEnd
		s.buf.WriteString(ts.C)
		return nil
	}
	return nil
}

func valueEnd(s *scanner, r rune) error {
	if len(s.state) == 0 {
		s.buf.WriteRune(r)
		return eof
	}

	if unicode.IsSpace(r) {
		s.buf.WriteRune(r)
		return nil
	}

	state := s.state[len(s.state)-1]

	switch state {
	case objState:
		switch r {
		case ',':
			s.next = keyStart
			s.buf.WriteRune(r)
			return nil
		case '}':
			s.popState()
			s.buf.WriteString(objWrap)
			s.buf.WriteRune(r)
			s.buf.WriteString(ts.C)
			return nil
		}
		return fmt.Errorf("invalid end object value '%v'", string(r))
	case arrState:
		switch r {
		case ',':
			s.next = valueStart
			return nil
		case ']':
			s.popState()
			return nil
		}
		return fmt.Errorf("invalid end array value '%v'", string(r))
	}

	return nil
}

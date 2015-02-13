package main

import (
	"bytes"
	"errors"
	"fmt"
	ts "github.com/james-maloney/termstyle"
	"strings"
	"unicode"
)

var eof = errors.New("eof")

const (
	_ = iota
	objState
	arrState
	skipState
	strValueState
)

var (
	objWrap  = ts.FG[ts.System10]
	arrWrap  = ts.FG[ts.System10]
	keyWrap  = ts.FG[ts.Grey74]
	strWrap  = ts.FG[ts.System14]
	boolWrap = ts.FG[ts.System9]
	nullWrap = ts.FG[ts.System9]
	numWrap  = ts.FG[ts.System4]
	errWrap  = ts.FG[ts.System1]
)

type scanner struct {
	state  []int
	next   func(*scanner, rune) error
	ftn    []rune // false/true/null helper state
	color  bool
	indent string
	last   rune

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
		r := rune(v)
		if err := s.next(s, r); err != nil {
			if err == eof {
				return nil
			}
			return err
		}
		s.last = r
	}
	if len(s.state) == 0 {
		return nil
	}

	return fmt.Errorf("JSON is incomplete")
}

func (s *scanner) wrap(str string) {
	if !s.color {
		return
	}
	s.buf.WriteString(str)
}

func (s *scanner) indentNewLine(r rune) {
	if len(s.state) == 0 {
		s.buf.WriteRune('\n')
		return
	}
	if s.buf.Len() == 0 {
		return
	}
	if s.last == '[' && r == ']' {
		return
	}

	s.buf.WriteRune('\n')
	s.buf.WriteString(strings.Repeat(s.indent, len(s.state)))
}

func (s *scanner) valueIndent(r rune) {
	if len(s.state) == 0 {
		return
	}
	if s.state[len(s.state)-1] == arrState {
		s.indentNewLine(r)
	}
}

func negNumStart(s *scanner, r rune) error {
	s.buf.WriteRune(r)
	if r >= '1' && r <= '9' {
		s.next = numStart
		return nil
	}
	if r == '0' {
		s.next = zeroNumStart
		return nil
	}

	return s.genErr("0-9")
}

func zeroNumStart(s *scanner, r rune) error {
	if r == '.' {
		s.buf.WriteRune(r)
		s.next = numStart
		return nil
	}

	s.wrap(ts.C)

	return endValue(s, r)
}

func endValue(s *scanner, r rune) error {
	if unicode.IsSpace(r) {
		return nil
	}

	switch s.state[len(s.state)-1] {
	case objState:
		switch r {
		case ',':
			s.next = keyStart
			s.buf.WriteRune(r)
			return nil
		case '}':
			s.popState()
			s.indentNewLine(r)
			s.wrap(objWrap)
			s.buf.WriteRune(r)
			s.wrap(ts.C)
			return nil
		}

		s.buf.WriteRune(r)
		return s.genErr(",", "}")
	case arrState:
		switch r {
		case ',':
			s.next = valueStart
			s.buf.WriteRune(r)
			return nil
		case ']':
			s.popState()
			s.indentNewLine(r)
			s.wrap(arrWrap)
			s.buf.WriteRune(r)
			s.wrap(ts.C)
			return nil
		}

		s.buf.WriteRune(r)
		return s.genErr(",", "]")
	}

	return fmt.Errorf("invalid num state")
}

func numStart(s *scanner, r rune) error {
	if r >= '0' && r <= '9' {
		s.buf.WriteRune(r)
		return nil
	}
	if r == '.' {
		s.buf.WriteRune(r)
		s.next = decimalStart
		return nil
	}
	if r == 'e' || r == 'E' {
		s.buf.WriteRune(r)
		s.next = eStart
		return nil
	}

	return numEnd(s, r)
}

func decimalStart(s *scanner, r rune) error {
	if r >= '0' && r <= '9' {
		s.buf.WriteRune(r)
		return nil
	}
	if r == 'e' || r == 'E' {
		s.buf.WriteRune(r)
		s.next = eStart
		return nil
	}
	return numEnd(s, r)
}

func eStart(s *scanner, r rune) error {
	s.buf.WriteRune(r)
	if r >= '0' || r <= 9 || r == '+' || r == '-' {
		s.next = numEnd
		return nil
	}

	return s.genErr("0-9", "+", "-")
}

func numEnd(s *scanner, r rune) error {
	if r >= '0' && r <= '9' {
		s.buf.WriteRune(r)
		return nil
	}

	s.wrap(ts.C)

	return endValue(s, r)
}

func valueStart(s *scanner, r rune) error {
	switch r {
	case '{':
		s.next = keyStart
		s.pushState(objState)
		s.wrap(objWrap)
		s.buf.WriteRune(r)
		s.wrap(ts.C)
		return nil
	case '[':
		s.next = valueStart
		s.pushState(arrState)
		s.wrap(arrWrap)
		s.buf.WriteRune(r)
		s.wrap(ts.C)
		return nil
	case ']':
		s.next = valueEnd
		s.popState()
		s.indentNewLine(r)
		s.wrap(arrWrap)
		s.buf.WriteRune(r)
		s.wrap(ts.C)
		return nil
	case '"':
		s.valueIndent(r)
		s.next = strValueStart
		s.pushState(strValueState)
		s.wrap(strWrap)
		s.buf.WriteRune(r)
		return nil
	case 't':
		s.valueIndent(r)
		s.ftn = []rune{'r', 'u', 'e'}
		s.next = boolNullStart
		s.wrap(boolWrap)
		s.buf.WriteRune(r)
		return nil
	case 'f':
		s.valueIndent(r)
		s.ftn = []rune{'a', 'l', 's', 'e'}
		s.next = boolNullStart
		s.wrap(boolWrap)
		s.buf.WriteRune(r)
		return nil
	case 'n':
		s.valueIndent(r)
		s.ftn = []rune{'u', 'l', 'l'}
		s.next = boolNullStart
		s.wrap(nullWrap)
		s.buf.WriteRune(r)
		return nil
	case '0':
		s.valueIndent(r)
		s.next = zeroNumStart
		s.wrap(numWrap)
		s.buf.WriteRune(r)
		return nil
	case '-':
		s.valueIndent(r)
		s.next = numStart
		s.wrap(numWrap)
		s.buf.WriteRune(r)
		return nil
	}
	if r >= '1' && r <= '9' {
		s.valueIndent(r)
		s.next = numStart
		s.wrap(numWrap)
		s.buf.WriteRune(r)
		return nil
	}
	if unicode.IsSpace(r) {
		return nil
	}

	s.buf.WriteRune(r)
	return s.genErr("")
}

func keyStart(s *scanner, r rune) error {
	if r == '"' {
		s.indentNewLine(r)
		s.next = keyEnd
		s.wrap(keyWrap)
		s.buf.WriteRune(r)
		return nil
	}
	if unicode.IsSpace(r) {
		return nil
	}
	if r == '}' {
		s.next = valueEnd
		s.popState()
		s.wrap(objWrap)
		s.buf.WriteRune(r)
		s.wrap(ts.C)
		return nil
	}

	s.buf.WriteRune(r)
	return s.genErr(`"`)
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
		s.wrap(ts.C)
		return nil
	}

	return nil
}

func colon(s *scanner, r rune) error {
	s.buf.WriteRune(r)
	if r == ':' {
		s.buf.WriteRune(' ')
		s.next = valueStart
		return nil
	}
	if unicode.IsSpace(r) {
		return nil
	}

	return s.genErr(":")
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
		s.wrap(ts.C)
		return nil
	}
	return nil
}

func boolNullStart(s *scanner, r rune) error {
	s.buf.WriteRune(r)
	if s.ftn[0] == r {
		s.ftn = s.ftn[1:]
		if len(s.ftn) == 0 {
			s.wrap(ts.C)
			s.next = valueEnd
		}
		return nil
	}

	return s.genErr(string(s.ftn[0]))
}

func valueEnd(s *scanner, r rune) error {
	if len(s.state) == 0 {
		fmt.Println("last rune", string(r))
		s.buf.WriteRune(r)
		return eof
	}

	if unicode.IsSpace(r) {
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
			s.indentNewLine(r)
			s.wrap(objWrap)
			s.buf.WriteRune(r)
			s.wrap(ts.C)
			return nil
		}

		s.buf.WriteRune(r)
		return s.genErr(",", "}")
	case arrState:
		switch r {
		case ',':
			s.next = valueStart
			s.buf.WriteRune(r)
			return nil
		case ']':
			s.popState()
			s.indentNewLine(r)
			s.wrap(arrWrap)
			s.buf.WriteRune(r)
			s.wrap(ts.C)
			return nil
		}

		s.buf.WriteRune(r)
		return s.genErr(",", "]")
	}

	return nil
}

func (s *scanner) genErr(expected ...string) error {
	errFmt := ""
	if len(expected) > 0 {
		errFmt = "%s" + fmt.Sprintf("invalid char, looking for '%s'", strings.Join(expected, " or "))
	} else {
		errFmt = "%sinvalid char"
	}

	str := s.buf.String()
	if len(str) == 0 {
		return fmt.Errorf(errFmt, "")
	}

	if len(str) == 1 {
		e := errWrap + str + ts.C + "<-- "
		return fmt.Errorf(errFmt, e)
	}

	// keep last 100 chars
	if len(str) > 100 {
		str = str[len(str)-100:]
	}

	// pop
	e, r := str[:len(str)-1], str[len(str)-1]

	e += ts.C + errWrap + string(r) + ts.C + "<-- "

	return fmt.Errorf(errFmt, e)
}

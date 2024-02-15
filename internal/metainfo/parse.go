/*
 * Copyright 2024 Eduardo Antunes dos Santos Vieira
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/* The BitTorrent protocol makes use of a small data markup language called
 * 'bencoding' by the official spec. Here, I will call it B-encoding. It
 * determines a standard text representation for strings, integers, lists and
 * dictionaries. In a nutshell:
 * - Strings => <length>:<text>
 * - Integers => i<num>e
 * - Lists => l(<value>)*e
 * - Dicts => d(<key><value>)*e
 * Dictionary keys must be strings, all numbers must be represented in base 10
 * and aren't supposed to be 0-prefixed.
 *
 * This file implements a simple parser for B-encoding.
 */

package metainfo

import (
	"bytes"
	"fmt"
)

// A parse error is defined by its reason, a textual context extracted from
// the encoded text (so that the user may quickly see why the error happened)
// and the position (column) where it happened
type parseError struct {
	reason, context string
	pos             int
}

func (err *parseError) Error() string {
	return fmt.Sprintf(`[!] Parsing error at column %d: %s
Relevant portion of the text: %s`, err.pos, err.reason, err.context)
}

// "Union" type for generic B-encoded values
type value struct {

}

type parser struct {
	enc []byte // B-encoded text
	i   int    // current position in the encoded string
}

func newParser(enc []byte) *parser {
	return &parser{enc: enc, i: 0}
}

// Reports errors with a piece of context from the encoded text itself,
// going from the current parsing position to the endContext parameter. If
// endContext is negative, the whole string after the current position is used
// as context
func (p *parser) err(reason string, endContext int) *parseError {
	if endContext < 0 {
		endContext = len(p.enc)
	}
	return &parseError{
		reason:  reason,
		context: string(p.enc[p.i:endContext]),
		pos:     p.i + 1,
	}
}

// Reports delimiter errors specifically, where the context is the delimiter
// itself and its position is used as the error position, instead of the
// current one. This makes error messages clearer
func (p *parser) errDelim(delim string, start int) *parseError {
	return &parseError{
		reason:  fmt.Sprintf("unmatched '%s' delimiter", delim),
		context: delim,
		pos:     start + 1,
	}
}

// Convert fixed-size ASCII bytes to int
func fromAscii(ascii []byte, n int) (int, bool) {
	var num = 0
	var exp = 1
	for i := n - 1; i >= 0; i-- {
		if ascii[i] < '0' || ascii[i] > '9' {
			return 0, false
		}
		num += int(ascii[i]-'0') * exp
		exp *= 10
	}
	return num, true
}

// Parses strings: <length>:<text>
func (p *parser) parseStr() (string, error) {
	i := bytes.IndexByte(p.enc[p.i:], ':')
	// no ':' in the text => invalid string
	if i < 0 {
		return "", p.err("invalid string", -1)
    }
	n, ok := fromAscii(p.enc[p.i:], i)
	// length isn't properly specified => invalid string
	if !ok {
		return "", p.err("invalid string length specifier", -1)
	}
	p.i += i + 1
	text := string(p.enc[p.i : p.i+n])
	p.i += n // advances the parser to the next token
	return text, nil
}

// Parses integers: i<num>e
func (p *parser) parseInt() (int, error) {
	if p.enc[p.i] != 'i' {
		return 0, p.err("invalid integer", -1)
	}
	end := bytes.IndexByte(p.enc[p.i:], 'e')
	// no terminating 'e' => unbalanced delimiter error
	if end < 0 {
		return 0, p.errDelim("i", p.i)
	}
	num, ok := fromAscii(p.enc[p.i+1:], end-1)
	// invalid integer value => invalid integer
	if !ok {
		return 0, p.err("invalid integer value", p.i+end+1)
	}
	p.i += end + 1 // advances the parser to the next token
	return num, nil
}

// Parses lists: l(<value>*)e
func (p *parser) parseList() ([]any, error) {
	if p.enc[p.i] != 'l' {
		return nil, p.err("invalid list", -1)
	}
	start := p.i
	p.i++ // advances past the 'l'
	vals := make([]any, 0, 64)
	if p.i >= len(p.enc) {
		// missing terminating 'e' => unbalanced delimiter error
		return nil, p.errDelim("l", start)
	}
	for p.enc[p.i] != 'e' {
		v, err := p.parseVal()
		if err != nil {
			// invalid value in the list => invalid list
			return nil, err
		}
		vals = append(vals, v)
		if p.i >= len(p.enc) {
			// missing terminating 'e' => unbalanced delimiter error
			return nil, p.errDelim("l", start)
		}
	}
	p.i++ // advances past the 'e'
	return vals, nil
}

// Parses dicts: d(<key><value>)*e
func (p *parser) parseDict() (map[string]any, error) {
	if p.enc[p.i] != 'd' {
		return nil, p.err("invalid dictionary", -1)
	}
	start := p.i
	p.i++ // advances past the 'd'
	dict := make(map[string]any)
	if p.i >= len(p.enc) {
		// missing terminating 'e' => unbalanced delimiter error
		return nil, p.errDelim("d", start)
	}
	for p.enc[p.i] != 'e' {
		key, err := p.parseStr() // keys must be strings
		if err != nil {
			// invalid key => invalid dict
			return nil, err
		}
		val, err := p.parseVal()
		if err != nil {
			// invalid value => invalid dict
			return nil, err
		}
		dict[key] = val
		if p.i >= len(p.enc) {
			// missing terminating 'e' => unbalanced delimiter error
			return nil, p.errDelim("d", start)
		}
	}
	p.i++ // advances past the 'e'
	return dict, nil
}

// General parsing method (don't use outside of this file)
func (p *parser) parseVal() (any, error) {
	// Check type id character
	switch p.enc[p.i] {
	case 'd':
		// Dict parsing
		return p.parseDict()
	case 'l':
		// List parsing
		return p.parseList()
	case 'i':
		// Integer parsing
		return p.parseInt()
	}
	// Either string parsing or an error
	return p.parseStr()
}

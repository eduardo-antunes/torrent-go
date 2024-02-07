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

/* The torrent protocol makes use of a small data markup language called
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

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// There are few parsing errors in the language of B-encoding; basically just
// invalid string and integer values, and lack of 'e' termination in the
// structures that require them.

type parseErrCode byte

const (
	parseErrInvalidStr = iota
	parseErrInvalidInt
	parseErrUnbalancedDelim
)

// A parsing error is defined by its type, a context that makes it clearer to
// the user what went wrong, and the position it happened in the text, given as
// 1-based column index
type ParseError struct {
	code    parseErrCode
	context string
	pos     int
}

func (err *ParseError) Error() string {
	var msg string
	switch err.code {
	case parseErrInvalidStr:
		msg = fmt.Sprintf("'%s' is not a valid string", err.context)
	case parseErrInvalidInt:
		msg = fmt.Sprintf("'%s' is not a valid integer", err.context)
	case parseErrUnbalancedDelim:
		msg = fmt.Sprintf("delimiter '%s' isn't terminated", err.context)
	default:
		msg = fmt.Sprintf("weird error %d", err.code)
	}
	return fmt.Sprintf("Error at column %d: %s", err.pos, msg)
}

type Parser struct {
	enc string // B-encoded text
	i   int    // current position in the encoded string
}

func NewParser(enc string) *Parser {
	return &Parser{enc: enc, i: 0}
}

// Reports errors with a piece of context from the encoded text itself,
// going from the current parsing position to the endContext parameter. If
// endContext is negative, the whole string after the current position is used
// as context
func (p *Parser) err(code parseErrCode, endContext int) *ParseError {
	if endContext < 0 {
		endContext = len(p.enc)
	}
	return &ParseError{
		code:    code,
		context: p.enc[p.i:endContext],
		pos:     p.i + 1,
	}
}

// Reports delimiter errors specifically, where the context is the delimiter
// itself and its position is used as the error position, instead of the
// current one. This makes error messages clearer
func (p *Parser) errDelim(delim string, start int) *ParseError {
	return &ParseError{
		code:    parseErrUnbalancedDelim,
		context: delim,
		pos:     start + 1,
	}
}

// Parses strings: <length>:<text>
func (p *Parser) ParseStr() (string, error) {
	i := strings.IndexByte(p.enc[p.i:], ':')
	// no ':' in the text => invalid string
	if i < 0 {
		return "", p.err(parseErrInvalidStr, -1)
	}
	n, err := strconv.Atoi(p.enc[p.i : p.i+i])
	// length isn't properly specified => invalid string
	if err != nil {
		return "", p.err(parseErrInvalidStr, -1)
	}
	p.i += i + 1
	text := p.enc[p.i : p.i+n]
	p.i += n // advances the parser to the next token
	return text, nil
}

// Parses integers: i<num>e
func (p *Parser) ParseInt() (int, error) {
	end := strings.IndexByte(p.enc[p.i:], 'e')
	// no terminating 'e' => unbalanced delimiter error
	if end < 0 {
		return 0, p.errDelim("i", p.i)
	}
	num, err := strconv.Atoi(p.enc[p.i+1 : p.i+end])
	// invalid integer value => invalid integer
	if err != nil {
		return 0, p.err(parseErrInvalidInt, p.i+end+1)
	}
	p.i += end + 1 // advances the parser to the next token
	return num, nil
}

// Parses lists: l(<value>*)e
func (p *Parser) ParseList() ([]any, error) {
	start := p.i
	p.i++ // advances past the 'l'
	vals := make([]any, 0, 64)
	if p.i >= len(p.enc) {
		// missing terminating 'e' => unbalanced delimiter error
		return vals, p.errDelim("l", start)
	}
	for p.enc[p.i] != 'e' {
		v, err := p.parse()
		if err != nil {
			// invalid value in the list => invalid list
			return vals, err
		}
		vals = append(vals, v)
		if p.i >= len(p.enc) {
			// missing terminating 'e' => unbalanced delimiter error
			return vals, p.errDelim("l", start)
		}
	}
	p.i++ // advances past the 'e'
	return vals, nil
}

// Parses dicts: d(<key><value>)*e
func (p *Parser) ParseDict() (map[string]any, error) {
	start := p.i
	p.i++ // advances past the 'd'
	dict := make(map[string]any)
	if p.i >= len(p.enc) {
		// missing terminating 'e' => unbalanced delimiter error
		return dict, p.errDelim("d", start)
	}
	for p.enc[p.i] != 'e' {
		key, err := p.ParseStr() // keys must be strings
		if err != nil {
			// invalid key => invalid dict
			return dict, err
		}
		val, err := p.parse()
		if err != nil {
			// invalid value => invalid dict
			return dict, err
		}
		dict[key] = val
		if p.i >= len(p.enc) {
			// missing terminating 'e' => unbalanced delimiter error
			return dict, p.errDelim("d", start)
		}
	}
	p.i++ // advances past the 'e'
	return dict, nil
}

// General parsing method (don't use externally)
func (p *Parser) parse() (any, error) {
	// Check type id character
	switch p.enc[p.i] {
	case 'd':
		// Dict parsing
		return p.ParseDict()
	case 'l':
		// List parsing
		return p.ParseList()
	case 'i':
		// Integer parsing
		return p.ParseInt()
	}
	// Either string parsing or an error
	return p.ParseStr()
}

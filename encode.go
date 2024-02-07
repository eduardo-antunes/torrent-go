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

/* Encode Go data structures into text via BitTorrent's B-encoding. B-encoding
 * is already extensively described in the parse.go file, which effectively
 * does the opposite from what is done here.
 */

package main

import (
	"fmt"
	"strings"
)

// Encodes strings: <length>:<text>
func EncodeStr(text string) string {
	return fmt.Sprintf("%d:%s", len(text), text)
}

// Encodes integers: i<num>e
func EncodeInt(num int) string {
	return fmt.Sprintf("i%de", num)
}

// Encodes lists: l(<value>)*e
func EncodeList(vals []any) string {
	var build strings.Builder
	build.WriteByte('l')
	for _, val := range vals {
		build.WriteString(encodeValue(val))
	}
	build.WriteByte('e')
	return build.String()
}

// Encodes dicts: d(<key><value>)*e
func EncodeDict(dict map[string]any) string {
	var build strings.Builder
	build.WriteByte('d')
	for key, val := range dict {
		build.WriteString(EncodeStr(key))
		build.WriteString(encodeValue(val))
	}
	build.WriteByte('e')
	return build.String()
}

// General encoding function (don't use externally)
func encodeValue(val any) string {
	switch v := val.(type) {
	case int:
		return EncodeInt(v)
	case string:
		return EncodeStr(v)
	case []any:
		return EncodeList(v)
	case map[string]any:
		return EncodeDict(v)
	}
	panic("What?!!!") // should never happen
}

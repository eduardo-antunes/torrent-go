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

// While the parser component of this package translates B-encoded text into
// unstructured dictionaries, this file translates those dictionaries into
// more useful data structures. It also verifies the validity of torrent files.
// Of course, most of the heavy lifting is handled by the excellent
// mapstructure library.

package benc

import (
	"crypto/sha1"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// Torrent metainfo structure, parsed from .torrent files
type Torrent struct {
	InfoHash     [20]byte // hash of the info field (used as torrent ID)
	Announce     string   // announce URL, for the tracker
	CreationDate int64    `mapstructure:"creation date"`
	CreatedBy    string   `mapstructure:"created by"`
	Comment      string

	// File information; two modes of representation
	singleFileMode bool           // is there only one file?
	singleInfo     singleFileInfo // info field in single-file mode
	multiInfo      multiFileInfo  // info field in multi-file mode
}

// Representation of a file in multi-file mode
type file struct {
	Path   []string
	Length uint64
}

// Torrent info field in single-file mode
type singleFileInfo struct {
	Name        string
	Pieces      string
	PieceLength uint64 `mapstructure:"piece length"`
	Length      uint64
}

// Torrent info field in multi-file mode
type multiFileInfo struct {
	Name        string
	Pieces      string
	PieceLength uint64 `mapstructure:"piece length"`
	Files       []file
}

// Parse metainfo structure from the contents of a standard torrent file
func ParseTorrent(fileContents []byte) (*Torrent, error) {
	// Parse the torrent to get a raw dictionary
	p := newParser(fileContents)
	rawMetainfo, err := p.parseDict()
	if err != nil {
		return nil, err // bad torrent file
	}
	// Validate all non-info fields of the raw dictionary, filling the torrent
	// structure with their values or triggering an error
	torrent := new(Torrent)
	if err = mapstructure.Decode(rawMetainfo, torrent); err != nil {
		return nil, fmt.Errorf("[!] Malformed torrent file\n%w", err)
	}
	// Tedious validation code for the info field
	rawInfo, ok := rawMetainfo["info"]
	if !ok {
		return nil, fmt.Errorf("[!] No info field present in the torrent\n")
	}
	info, ok := rawInfo.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("[!] Info field must be a dictionary\n")
	}
	// The info field must be handled specially, as its structure will be
	// different in single-file mode and multi-file mode. We detect the mode
	// by checking for info.files
	if _, filesExists := info["files"]; filesExists {
		// Multi-file mode torrent
		torrent.singleFileMode = false
		if err = mapstructure.Decode(info, &torrent.multiInfo); err != nil {
			return nil, fmt.Errorf("[!] Malformed multi-file info field\n%w", err)
		}
	} else {
		// Single-file mode torrent
		torrent.singleFileMode = true
		if err = mapstructure.Decode(info, &torrent.singleInfo); err != nil {
			return nil, fmt.Errorf("[!] Malformed multi-file info field\n%w", err)
		}
	}
	// Some metainfo fields (notably InfoHash) are not directly present in the
	// torrent file, but must instead be computed from it
	bencInfo := encodeDict(info)
	torrent.InfoHash = sha1.Sum([]byte(bencInfo))
	return torrent, nil
}

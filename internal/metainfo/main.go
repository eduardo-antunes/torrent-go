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

/* While the parser component of this package translates B-encoded text into
 * unstructured dictionaries, this file translates those dictionaries into
 * more useful data structures. It also verifies the validity of torrent files.
 * Of course, most of the heavy lifting is handled by the excellent
 * mapstructure library.
 */

package metainfo

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type metaInfoError string // simple error type for now

func (err metaInfoError) Error() string {
	return fmt.Sprintf("[!] Metainfo parsing error: %s", string(err))
}

// Torrent metainfo structure, parsed from .torrent files
type MetaInfo struct {
	Announce     string
	Info         TorrentInfo
	CreationDate int64  `mapstructure:"creation date"`
	CreatedBy    string `mapstructure:"created by"`
	Comment      string
}

// The info field of the metainfo structure, probably the most important
// component of any torrent file. Single file mode only for now
type TorrentInfo struct {
	Name        string
	Length      int
	PieceLength int `mapstructure:"piece length"`
	Pieces      string
}

// Parse metainfo structure from the contents of a standard torrent file
func ParseMetaInfo(fileContents []byte) (*MetaInfo, error) {
	p := newParser(fileContents)
	rawMetaInfo, err := p.parseDict()
	if err != nil {
		return nil, metaInfoError("not a proper B-encoded dictionary")
	}
	metaInfo := new(MetaInfo)
	err = mapstructure.Decode(rawMetaInfo, metaInfo)
	if err != nil {
		return nil, err
	}
	return metaInfo, nil
}

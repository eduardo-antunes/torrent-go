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

package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/eduardo-antunes/torrent-go/internal/benc"
)


func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <file.torrent>\n", os.Args[0])
		return
	}
	contents, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Could not open torrent file %s\n", os.Args[1])
		return
	}
	torrent, err := benc.ParseMetaInfo(contents)
	if err != nil {
		fmt.Println(err)
		return
	}
    announceUrl, _ := url.Parse(torrent.Announce)
    query := NewTrackerQuery(string(torrent.InfoHash[:]), torrent.Info.Length, 6881)
    TrackerAnnounce(announceUrl, query)
}

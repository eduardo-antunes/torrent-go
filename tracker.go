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
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Parameters for tracker requests
type TrackerQuery struct {
	infoHash   string // torrent ID
	peerId     string // peer ID of this computer
	event      string // one of started, stopped or completed
	port       int    // network port being used
	uploaded   int    // # of bytes uploaded
	downloaded int    // # of bytes downloaded
	left       int    // # of bytes left for completion
}

// Initialize a new tracker query
func NewTrackerQuery(infoHash string, torrentLength, port int) *TrackerQuery {
	return &TrackerQuery{
		infoHash:   infoHash,
		peerId:     generatePeerId(),
		event:      "started",
		port:       port,
		uploaded:   0,
		downloaded: 0,
		left:       torrentLength,
	}
}

func TrackerAnnounce(announce *url.URL, tq *TrackerQuery) error {
	// Very boring, but functional code to build the raw query
	var vals url.Values
	vals.Add("peer_id", url.QueryEscape(tq.peerId))
	vals.Add("info_hash", url.QueryEscape(tq.infoHash))
	vals.Add("downloaded", strconv.Itoa(tq.downloaded))
	vals.Add("uploaded", strconv.Itoa(tq.uploaded))
	vals.Add("port", strconv.Itoa(tq.port))
	vals.Add("left", strconv.Itoa(tq.left))
	vals.Add("event", tq.event)
	announce.RawQuery = vals.Encode()

	resp, err := http.Get(announce.String())
	if err != nil {
		return fmt.Errorf("[!] Could not make HTTP get request: %w", err)
	}
    defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[!] HTTP response with bad status code: %d",
			resp.StatusCode)
	}
    // The text variable is B-encoded
    text, _ := io.ReadAll(resp.Body)
    fmt.Println(string(text))
	return nil
}

// Generates a semi-random peer ID for this computer
func generatePeerId() string {
	// Peer ID = client ID + random bytes
	var build strings.Builder
	clientId := "-TG1000-"
	build.WriteString(clientId)
	for i := 0; i < 20-len(clientId); i++ {
		x := byte(rand.Int() & 0xFF)
		build.WriteByte(x)
	}
	return build.String()
}

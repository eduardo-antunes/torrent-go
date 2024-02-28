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
	"math/rand"
	"net/url"
	"strings"
)

// Parameters for tracker requests
type TrackerRequest struct {
	infoHash   string // torrent ID (hash of info field)
	peerId     string // peer ID (indentifies this computer)
	event      string // one of started, stopped or completed
	port       uint16 // network port being used
	uploaded   uint64 // # of bytes uploaded
	downloaded uint64 // # of bytes downloaded
	left       uint64 // # of bytes left for completion
}

// Create a tracker request object corresponding to the initial, "announce"
// request that is first sent to the tracker
func NewTrackerAnnounce(infoHash string, torrentLength uint64,
	port uint16) *TrackerRequest {
	return &TrackerRequest{
		infoHash:   infoHash,
		peerId:     generatePeerId(),
		event:      "started",
		port:       port,
		uploaded:   0,
		downloaded: 0,
		left:       torrentLength,
	}
}

// Encode the tracker request parameters into a query string
func (req *TrackerRequest) Query() string {
	return fmt.Sprintf("info_hash=%s&peer_id=%s&event=%s&port=%v&uploaded=%v"+
		"&downloaded=%v&left=%v&compact=1", url.QueryEscape(req.infoHash),
		url.QueryEscape(req.peerId), req.event, req.port, req.uploaded,
		req.downloaded, req.left)
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

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

package benc

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type TrackerResponse struct {
	interval    int    // interval, in seconds, between tracker requests
	minInterval int    // minimum value for the interval
	trackerId   string // tracker ID
	Peers       string // connected peers in a compact model
}

func ParseTrackerResponse(responseContents []byte) (*TrackerResponse, error) {
	p := newParser(responseContents)
	rawResp, err := p.parseDict()
	if err != nil {
		return nil, err
	}
	if msg, failedResp := rawResp["failure reason"]; failedResp {
		return nil, fmt.Errorf("[!] Internal tracker error: %s", msg)
	}

	resp := new(TrackerResponse)
	if err = mapstructure.Decode(rawResp, resp); err != nil {
		return nil, fmt.Errorf("[!] Malformed response from tracker\n%w", err)
	}
	return resp, nil
}

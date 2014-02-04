/* Copyright 2014 Ooyala, Inc. All rights reserved.
 *
 * This file is licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is
 * distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and limitations under the License.
 */

package backend

import (
	"net/http"
	"time"
)

const (
	StatusOk          = "OK"
	StatusDegraded    = "DEGRADED"
	StatusCritical    = "CRITICAL"
	StatusMaintenance = "MAINTENANCE"
)

type ServerStatus struct {
	Current string
	Checked time.Time
	Changed time.Time
}

func NewServerStatus() ServerStatus {
	return ServerStatus{
		Current: StatusMaintenance,
		Checked: time.Now(),
		Changed: time.Now(),
	}
}

func (s *ServerStatus) Set(status string) {
	s.Checked = time.Now()
	if s.Current != status {
		s.Current = status
		s.Changed = s.Checked
	}

}

func StatusWeight(s string) uint32 {
	switch s {
	case StatusOk:
		return 0x10000000
	case StatusDegraded:
		return 0x30000000
	case StatusCritical:
		return 0x70000000
	default:
		// "CRITICAL".StatusWeight()
		return 0x70000000
	}
}

func IsValidStatus(s string) bool {
	return s == StatusOk || s == StatusDegraded || s == StatusCritical
}

func (s *ServerStatus) ParseAndSet(res *http.Response) {
	if res.StatusCode != http.StatusOK {
		s.Set(StatusMaintenance)
		return
	}

	hdr := res.Header.Get("Server-Status")
	if IsValidStatus(hdr) {
		s.Set(hdr)
		return
	}

	s.Set(StatusMaintenance)
}

func (s *ServerStatus) Cost(accept string) uint32 {
	cost := StatusWeight(s.Current) &^ StatusWeight(accept)
	return cost + s.SlowStartFactor()
}

const (
	Tstartup = 60   // Startup time in seconds
	Kstartup = 4096 // Maximum slow start cost
)

func (s *ServerStatus) SlowStartFactor() uint32 {
	if !IsValidStatus(s.Current) {
		return 0
	}

	d := time.Now().Unix() - s.Changed.Unix()
	f := uint32(0)
	if d > Tstartup {
		f = 0
	} else if d > 0 {
		k := float64(Kstartup)
		f = uint32(k/float64(d) - k/float64(Tstartup))
	} else {
		// d == 0
		f = Kstartup
	}

	return f
}

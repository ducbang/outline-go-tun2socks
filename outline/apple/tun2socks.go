// Copyright 2019 The Outline Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tun2socks

import (
	"errors"
	"fmt"
	"io"
	"math"
	"runtime/debug"
	"time"

	"github.com/Jigsaw-Code/outline-go-tun2socks/outline"
	"github.com/Jigsaw-Code/outline-ss-server/client"
)

// OutlineTunnel embeds the tun2socks.Tunnel interface so it gets exported by gobind.
type OutlineTunnel interface {
	outline.Tunnel
}

// TunWriter is an interface that allows for outputting packets to the TUN (VPN).
type TunWriter interface {
	io.WriteCloser
}

func init() {
	// Apple VPN extensions have a memory limit of 15MB. Conserve memory by increasing garbage
	// collection frequency and returning memory to the OS every minute.
	debug.SetGCPercent(10)
	// TODO: Check if this is still needed in go 1.13, which returns memory to the OS
	// automatically.
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for range ticker.C {
			debug.FreeOSMemory()
		}
	}()
}

// ConnectShadowsocksTunnel reads packets from a TUN device and routes it to a Shadowsocks proxy server.
// Returns an OutlineTunnel instance that should be used to input packets to the tunnel.
//
// `tunWriter` is used to output packets to the TUN (VPN).
// `host` is  IP address of the Shadowsocks proxy server.
// `port` is the port of the Shadowsocks proxy server.
// `password` is the password of the Shadowsocks proxy.
// `cipher` is the encryption cipher the Shadowsocks proxy.
// `prefix` is the salt prefix to use for TCP connections (optional, use with care).
// `isUDPEnabled` indicates whether the tunnel and/or network enable UDP proxying.
//
// Sets an error if the tunnel fails to connect.
func ConnectShadowsocksTunnel(tunWriter TunWriter, host string, port int, password, cipher string, prefix []byte, isUDPEnabled bool) (OutlineTunnel, error) {
	if tunWriter == nil {
		return nil, errors.New("Must provide a TunWriter")
	} else if port <= 0 || port > math.MaxUint16 {
		return nil, fmt.Errorf("Invalid port number: %v", port)
	}
	ssclient, err := client.NewClient(host, port, password, cipher)
	if err != nil {
		return nil, fmt.Errorf("failed to construct Shadowsocks client: %v", err)
	}
	ssclient.SetTCPSaltGenerator(client.NewPrefixSaltGenerator(prefix))
	return outline.NewTunnel(ssclient, isUDPEnabled, tunWriter)
}

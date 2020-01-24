// Copyright 2019-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"context"
	"fmt"
	streams "github.com/atomix/go-framework/pkg/atomix/stream"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/raft"
	"net"
	"time"
)

const clientTimeout = 15 * time.Second

// newClient returns a new Raft consensus protocol client
func newClient(address raft.ServerAddress, ports map[string]int, r *raft.Raft, fsm *StateMachine, streams *streamManager) *Client {
	return &Client{
		address: address,
		ports:   ports,
		raft:    r,
		state:   fsm,
		streams: streams,
	}
}

// Client is the Raft client
type Client struct {
	address raft.ServerAddress
	ports   map[string]int
	raft    *raft.Raft
	state   *StateMachine
	streams *streamManager
}

func (c *Client) MustLeader() bool {
	return true
}

func (c *Client) IsLeader() bool {
	return c.raft.Leader() == c.address
}

func (c *Client) Leader() string {
	leader := c.raft.Leader()
	if leader == "" {
		return ""
	}

	// Get the IP address of the leader
	leaderIP, _, err := net.SplitHostPort(string(leader))
	if err != nil {
		return ""
	}

	// Match the leader's IP address to the IP of a node
	for host, port := range c.ports {
		addrs, err := net.LookupHost(host)
		if err == nil && len(addrs) > 0 && addrs[0] == leaderIP {
			return fmt.Sprintf("%s:%d", host, port)
		}
	}
	return ""
}

func (c *Client) Write(ctx context.Context, input []byte, stream streams.WriteStream) error {
	streamID, stream := c.streams.addStream(stream)
	entry := &Entry{
		Value:     input,
		StreamID:  streamID,
		Timestamp: time.Now(),
	}
	bytes, err := proto.Marshal(entry)
	if err != nil {
		return err
	}
	future := c.raft.Apply(bytes, clientTimeout)
	if future.Error() != nil {
		stream.Close()
		return future.Error()
	}
	response := future.Response()
	if response != nil {
		return response.(error)
	}
	return nil
}

func (c *Client) Read(ctx context.Context, input []byte, stream streams.WriteStream) error {
	return c.state.Query(input, stream)
}

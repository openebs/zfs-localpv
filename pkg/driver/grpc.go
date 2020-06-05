/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/sirupsen/logrus"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// parseEndpoint should have a valid prefix(unix/tcp) to return a valid endpoint parts
func parseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("Invalid endpoint: %v", ep)
}

// logGRPC logs all the grpc related errors, i.e the final errors
// which are returned to the grpc clients
func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logrus.Infof("GRPC call: %s", info.FullMethod)
	logrus.Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		logrus.Errorf("GRPC error: %v", err)
	} else {
		logrus.Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}

// NonBlockingGRPCServer defines Non blocking GRPC server interfaces
type NonBlockingGRPCServer interface {
	// Start services at the endpoint
	Start()

	// Waits for the service to stop
	Wait()

	// Stops the service gracefully
	Stop()

	// Stops the service forcefully
	ForceStop()
}

// NewNonBlockingGRPCServer returns a new instance of NonBlockingGRPCServer
func NewNonBlockingGRPCServer(ep string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer) NonBlockingGRPCServer {
	return &nonBlockingGRPCServer{
		endpoint:     ep,
		idnty_server: ids,
		ctrl_server:  cs,
		agent_server: ns}
}

// NonBlocking server
// dont block the execution for a task to complete.
// use wait group to wait for all the tasks dispatched.
type nonBlockingGRPCServer struct {
	wg           sync.WaitGroup
	server       *grpc.Server
	endpoint     string
	idnty_server csi.IdentityServer
	ctrl_server  csi.ControllerServer
	agent_server csi.NodeServer
}

// Start grpc server for serving CSI endpoints
func (s *nonBlockingGRPCServer) Start() {

	s.wg.Add(1)

	go s.serve(s.endpoint, s.idnty_server, s.ctrl_server, s.agent_server)

	return
}

// Wait for the service to stop
func (s *nonBlockingGRPCServer) Wait() {
	s.wg.Wait()
}

// Stop the service forcefully
func (s *nonBlockingGRPCServer) Stop() {
	s.server.GracefulStop()
}

// ForceStop the service
func (s *nonBlockingGRPCServer) ForceStop() {
	s.server.Stop()
}

// serve starts serving requests at the provided endpoint based on the type of
// plugin. In this function all the csi related interfaces are provided by
// container-storage-interface
func (s *nonBlockingGRPCServer) serve(endpoint string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer) {

	proto, addr, err := parseEndpoint(endpoint)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	// Clear off the addr if it is already present, this is done to remove stale
	// entries, as this path is shared with the OS and will be the same
	// everytime the plugin restarts, its possible that the last instance leaves
	// a stale entry
	if proto == "unix" {
		addr = "/" + addr
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			logrus.Fatalf("Failed to remove %s, error: %s", addr, err.Error())
		}
	}

	listener, err := net.Listen(proto, addr)
	if err != nil {
		logrus.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logGRPC),
	}
	// Create a new grpc server, all the request from csi client to
	// create/delete/... will hit this server
	server := grpc.NewServer(opts...)
	s.server = server

	if ids != nil {
		csi.RegisterIdentityServer(server, ids)
	}
	if cs != nil {
		csi.RegisterControllerServer(server, cs)
	}
	if ns != nil {
		csi.RegisterNodeServer(server, ns)
	}

	logrus.Infof("Listening for connections on address: %#v", listener.Addr())

	// Start serving requests on the grpc server created
	server.Serve(listener)

}

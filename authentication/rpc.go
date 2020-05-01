package main

import (
	pbauth "common/proto"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

func RunRpcServer() {
	listener, err := net.Listen("tcp", ":5300")

	if err != nil {
		log.Panic("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	pbauth.RegisterAuthRpcServer(grpcServer, &AuthRpcServer{})
	log.Println("AuthRpc service started")
	grpcServer.Serve(listener)
}

type AuthRpcServer struct{}

func (s *AuthRpcServer) Validate(c context.Context, request *pbauth.ValidateRequest) (response *pbauth.ValidateResponse, err error) {
	result, err := doValidate(request.AccessToken)
	if err != nil {
		return nil, err
	}
	response = &pbauth.ValidateResponse{
		Username:    result.Username,
		Permissions: result.Permissions,
	}
	return response, nil
}

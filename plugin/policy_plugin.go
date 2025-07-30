/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"context"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/codes"

	"github.com/oscal-compass/compliance-to-policy-go/v2/api/proto"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

// Plugin must return an RPC server for this plugin type.
var _ proto.PolicyEngineServiceServer = (*pvpService)(nil)

type pvpService struct {
	proto.UnimplementedPolicyEngineServiceServer
	Impl policy.Provider
}

func FromPVP(pe policy.Provider) proto.PolicyEngineServiceServer {
	return &pvpService{
		Impl: pe,
	}
}

func (p *pvpService) Configure(ctx context.Context, request *proto.ConfigureRequest) (*proto.ConfigureResponse, error) {
	if err := p.Impl.Configure(ctx, request.Settings); err != nil {
		return &proto.ConfigureResponse{}, status.Error(codes.Internal, err.Error())
	}

	// policy.Provider.Configure currently only returns an error, so using an empty proto.ConifgureResponse
	return &proto.ConfigureResponse{}, nil
}

func (p *pvpService) Generate(ctx context.Context, request *proto.GenerateRequest) (*proto.GenerateResponse, error) {
	rules := NewPolicyFromProto(request.Rule)
	if err := p.Impl.Generate(ctx, rules); err != nil {
		return &proto.GenerateResponse{}, status.Error(codes.Internal, err.Error())
	}

	// policy.Provider.Generate currently only returns an error, so using an empty proto.GenerateResponse
	return &proto.GenerateResponse{}, nil
}

func (p *pvpService) GetResults(ctx context.Context, request *proto.GetResultsRequest) (*proto.GetResultsResponse, error) {
	rules := NewPolicyFromProto(request.Rule)
	result, err := p.Impl.GetResults(ctx, rules)
	if err != nil {
		return &proto.GetResultsResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &proto.GetResultsResponse{Result: ResultsToProto(result)}, nil
}

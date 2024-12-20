package gapi

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Metadata struct {
	UserAgent string
	ClientIp  string
}

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgent                  = "user-agent"
	xForwardedForHeader        = "x-forwarded-for"
)

func (s *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		if userAgents := md.Get(userAgent); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		if clientIps := md.Get(xForwardedForHeader); len(clientIps) > 0 {
			mtdt.ClientIp = clientIps[0]
		}
	}

	p, ok := peer.FromContext(ctx)
	if ok {
		mtdt.ClientIp = p.Addr.String()
	}

	return mtdt
}

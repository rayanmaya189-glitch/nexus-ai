package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxConcurrent   int
	KeepAliveTime   time.Duration
	KeepAliveTimeout time.Duration
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:            50050,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		MaxConcurrent:   100,
		KeepAliveTime:   30 * time.Second,
		KeepAliveTimeout: 5 * time.Second,
	}
}

func NewGRPCServer(cfg *ServerConfig) (*grpc.Server, error) {
	addr := fmt.Sprintf(":%d", cfg.Port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	server := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    cfg.KeepAliveTime,
			Timeout: cfg.KeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	go func() {
		log.Printf("gRPC server listening on %s", addr)
		if err := server.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	return server, nil
}

type ClientConfig struct {
	Address         string
	Timeout         time.Duration
	KeepAliveTime   time.Duration
	KeepAliveTimeout time.Duration
	MaxRetries      int
}

func DefaultClientConfig(address string) *ClientConfig {
	return &ClientConfig{
		Address:         address,
		Timeout:         10 * time.Second,
		KeepAliveTime:   30 * time.Second,
		KeepAliveTimeout: 5 * time.Second,
		MaxRetries:      3,
	}
}

func NewGRPCClient(cfg *ClientConfig) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.KeepAliveTime,
			Timeout:             cfg.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", cfg.Address, err)
	}

	return conn, nil
}

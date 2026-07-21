package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ServiceClient struct {
	conn    *grpc.ClientConn
	address string
	timeout time.Duration
}

func NewServiceClient(address string, timeout time.Duration) (*ServiceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	return &ServiceClient{
		conn:    conn,
		address: address,
		timeout: timeout,
	}, nil
}

func (c *ServiceClient) Close() error {
	return c.conn.Close()
}

func (c *ServiceClient) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *ServiceClient) Address() string {
	return c.address
}

// ContextWithMetadata creates a context with auth metadata for downstream services
func ContextWithMetadata(ctx context.Context, userID, tenantID int64, email string, roles []string) context.Context {
	md := metadata.Pairs(
		"x-user-id", fmt.Sprintf("%d", userID),
		"x-tenant-id", fmt.Sprintf("%d", tenantID),
		"x-email", email,
	)
	for i, role := range roles {
		md.Append("x-roles", fmt.Sprintf("%d:%s", i, role))
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// ServiceCaller provides a way to call other services with retries
type ServiceCaller struct {
	registry   *ServiceRegistry
	maxRetries int
}

func NewServiceCaller(registry *ServiceRegistry) *ServiceCaller {
	return &ServiceCaller{
		registry:   registry,
		maxRetries: 3,
	}
}

func (sc *ServiceCaller) Call(ctx context.Context, serviceName string, fn func(conn *grpc.ClientConn) error) error {
	var lastErr error

	for attempt := 0; attempt < sc.maxRetries; attempt++ {
		addr, err := sc.registry.GetServiceAddress(serviceName)
		if err != nil {
			return fmt.Errorf("service discovery failed: %w", err)
		}

		client, err := NewServiceClient(addr, 10*time.Second)
		if err != nil {
			lastErr = err
			continue
		}

		err = fn(client.conn)
		client.Close()

		if err == nil {
			return nil
		}

		lastErr = err
		sc.registry.UpdateHealth(serviceName, addr, false)
	}

	return fmt.Errorf("all %d attempts failed: %w", sc.maxRetries, lastErr)
}

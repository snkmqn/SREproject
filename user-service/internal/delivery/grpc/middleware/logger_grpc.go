package middleware

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
)

func ClientInterceptor(serviceName string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		log.Printf("[Client: %s] Calling method: %s", serviceName, method)

		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		if err != nil {
			log.Printf("[Client: %s] Method: %s, Duration: %v, Status: Failed, Error: %v", serviceName, method, duration, err)
		} else {
			log.Printf("[Client: %s] Method: %s, Duration: %v, Status: Success", serviceName, method, duration)
		}

		return err
	}
}

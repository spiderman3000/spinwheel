package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type userIDKey struct{}

// UserIDFromContext retrieves the optional user ID from context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}

// UserIDInterceptor attaches X-User-Id to the context when present.
//
// Anon-first (contract v1): the header is an optional fallback for
// grpcurl/tests — requests without it proceed anonymously instead of
// being rejected. The canonical anon identity is the sw_sid cookie,
// resolved by the HTTP gateway.
func UserIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get("x-user-id"); len(ids) > 0 && ids[0] != "" {
				ctx = context.WithValue(ctx, userIDKey{}, ids[0])
			}
		}
		return handler(ctx, req)
	}
}

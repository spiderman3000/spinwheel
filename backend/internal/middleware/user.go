package middleware

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userIDKey struct{}

// UserIDFromContext retrieves the user ID from context
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}

// UserIDInterceptor extracts X-User-Id from metadata
func UserIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		userIDs := md.Get("x-user-id")
		if len(userIDs) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing x-user-id header")
		}

		// Add user ID to context
		ctx = context.WithValue(ctx, userIDKey{}, userIDs[0])
		
		return handler(ctx, req)
	}
}

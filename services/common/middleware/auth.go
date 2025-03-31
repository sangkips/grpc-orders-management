// services/common/middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/sangkips/order-processing-system/services/common/genproto/auth/auth"
	"github.com/sangkips/order-processing-system/services/common/util"
	"google.golang.org/grpc"
)

type AuthMiddleware struct {
    authClient auth.AuthServiceClient
}

func NewAuthMiddleware(conn *grpc.ClientConn) *AuthMiddleware {
    return &AuthMiddleware{
        authClient: auth.NewAuthServiceClient(conn),
    }
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            util.WriteError(w, http.StatusUnauthorized, errors.New("authorization header required"))
            return
        }

        // Extract the token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            util.WriteError(w, http.StatusUnauthorized, errors.New("invalid authorization header format"))
            return
        }

        token := parts[1]

        // Validate the token
        ctx := r.Context()
        resp, err := m.authClient.ValidateToken(ctx, &auth.ValidateTokenRequest{Token: token})
        if err != nil {
            util.WriteError(w, http.StatusUnauthorized, errors.New("failed to validate token"))
            return
        }

        if !resp.Valid {
            util.WriteError(w, http.StatusUnauthorized, errors.New("invalid token"))
            return
        }

        // Add user info to the request context
        userCtx := context.WithValue(r.Context(), "user", resp.User)
        next.ServeHTTP(w, r.WithContext(userCtx))
    })
}

func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user, ok := r.Context().Value("user").(*auth.UserInfo)
            if !ok {
                util.WriteError(w, http.StatusUnauthorized, errors.New("user not authenticated"))
                return
            }

            hasRole := false
            for _, userRole := range user.Roles {
                if userRole == role {
                    hasRole = true
                    break
                }
            }

            if !hasRole {
                util.WriteError(w, http.StatusForbidden, errors.New("insufficient permissions"))
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
// services/auth-service/service/auth.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sangkips/order-processing-system/services/common/genproto/auth/auth"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
    ID       int64
    Username string
    Password string
    Roles    []string
}

type AuthService struct {
    users       map[string]User
    jwtSecret   []byte
    tokenExpiry time.Duration
}

func NewAuthService(jwtSecret string) *AuthService {
    // In a real app, you'd use a database
    users := map[string]User{
        "admin": {
            ID:       1,
            Username: "admin",
            Password: hashPassword("admin123"),
            Roles:    []string{"admin"},
        },
        "user": {
            ID:       2,
            Username: "user",
            Password: hashPassword("user123"),
            Roles:    []string{"user"},
        },
    }
    
    return &AuthService{
        users:       users,
        jwtSecret:   []byte(jwtSecret),
        tokenExpiry: 24 * time.Hour,
    }
}

func hashPassword(password string) string {
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hashedPassword)
}

func (s *AuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
    user, exists := s.users[req.Username]
    if !exists {
        return nil, errors.New("invalid credentials")
    }
    
    err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
    if err != nil {
        return nil, errors.New("invalid credentials")
    }
    
    // Generate JWT token
    expiresAt := time.Now().Add(s.tokenExpiry)
    token, err := s.generateToken(user, expiresAt)
    if err != nil {
        return nil, err
    }
    
    refreshToken, err := s.generateRefreshToken(user)
    if err != nil {
        return nil, err
    }
    
    return &auth.LoginResponse{
        AccessToken:  token,
        RefreshToken: refreshToken,
        ExpiresAt:    expiresAt.Unix(),
        User: &auth.UserInfo{
            UserId:   user.ID,
            Username: user.Username,
            Roles:    user.Roles,
        },
    }, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
    token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return s.jwtSecret, nil
    })
    
    if err != nil || !token.Valid {
        return &auth.ValidateTokenResponse{Valid: false}, nil
    }
    
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return &auth.ValidateTokenResponse{Valid: false}, nil
    }
    
    username, ok := claims["sub"].(string)
    if !ok {
        return &auth.ValidateTokenResponse{Valid: false}, nil
    }
    
    user, exists := s.users[username]
    if !exists {
        return &auth.ValidateTokenResponse{Valid: false}, nil
    }
    
    return &auth.ValidateTokenResponse{
        Valid: true,
        User: &auth.UserInfo{
            UserId:   user.ID,
            Username: user.Username,
            Roles:    user.Roles,
        },
    }, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
    // Validate refresh token (in a real app, you'd check against stored tokens)
    token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return s.jwtSecret, nil
    })
    
    if err != nil || !token.Valid {
        return nil, errors.New("invalid refresh token")
    }
    
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }
    
    username, ok := claims["sub"].(string)
    if !ok {
        return nil, errors.New("invalid token subject")
    }
    
    user, exists := s.users[username]
    if !exists {
        return nil, errors.New("user not found")
    }
    
    // Generate new tokens
    expiresAt := time.Now().Add(s.tokenExpiry)
    newToken, err := s.generateToken(user, expiresAt)
    if err != nil {
        return nil, err
    }
    
    newRefreshToken, err := s.generateRefreshToken(user)
    if err != nil {
        return nil, err
    }
    
    return &auth.RefreshTokenResponse{
        AccessToken:  newToken,
        RefreshToken: newRefreshToken,
        ExpiresAt:    expiresAt.Unix(),
    }, nil
}

func (s *AuthService) generateToken(user User, expiresAt time.Time) (string, error) {
    claims := jwt.MapClaims{
        "sub":  user.Username,
        "id":   user.ID,
        "exp":  expiresAt.Unix(),
        "iat":  time.Now().Unix(),
        "roles": user.Roles,
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.jwtSecret)
}

func (s *AuthService) generateRefreshToken(user User) (string, error) {
    claims := jwt.MapClaims{
        "sub":  user.Username,
        "id":   user.ID,
        "exp":  time.Now().Add(30 * 24 * time.Hour).Unix(), // 30 days
        "iat":  time.Now().Unix(),
        "type": "refresh",
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.jwtSecret)
}
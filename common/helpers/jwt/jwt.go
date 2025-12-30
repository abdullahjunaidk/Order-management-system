package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID       string   `json:"userId"`
	IsSuperAdmin bool     `json:"isSuperAdmin"`
	CompanyIDs   []string `json:"companyIds"`
	TokenType    string   `json:"type"`
	jwt.RegisteredClaims
}

// generateToken is a helper function to generate JWT tokens.
//
// Parameters:
//   - userID (string): User ID for whom to generate the token.
//   - isSuperAdmin (bool): Whether the user is a super admin.
//   - companyIds ([]string): List of company IDs associated with the user.
//   - tokenType (string): Type of token to generate.
//   - expiry (time.Duration): Token expiry duration.
//   - secretKey (string): Secret key to sign the token.
//
// Returns:
//   - string: JWT token.
//   - error: An error if token generation fails.
// func generateToken(userID string, tokenType string, expiry time.Duration, secretKey string) (string, error) {
// 	iat := time.Now()
// 	exp := time.Now().Add(expiry)
// 	claims := jwt.MapClaims{
// 		"employee_id": userID,
// 		"exp":         exp.Unix(),
// 		"iat":         iat.Unix(),
// 		"type":        tokenType,
// 	}

// 	log.Println("expiry time", expiry)
// 	log.Printf("issued at %s", iat)
// 	log.Printf("expires at %s", exp)

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	signedToken, err := token.SignedString([]byte(secretKey))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to sign token: %w", err)
// 	}

//		return signedToken, nil
//	}
func generateToken(userID string, isSuperAdmin bool, companyIds []string, tokenType string, expiry time.Duration, secretKey string) (string, error) {
	claims := Claims{
		UserID:       userID,
		IsSuperAdmin: isSuperAdmin,
		CompanyIDs:   companyIds,
		TokenType:    tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// verifyToken is a helper function to verify JWT tokens.
//
// Parameters:
//   - tokenString (string): JWT token to verify.
//   - tokenType (string): Type of token to verify.
//   - secretKey (string): Secret key to verify the token.
//
// Returns:
//   - jwt.MapClaims: Claims of the token.
//   - error: An error if token verification fails.
func verifyToken(tokenString string, tokenType string, secretKey string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims["type"] != tokenType {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}

// GenerateAccessToken is a helper function to generate a refresh token.
//
// Parameters:
//   - userID (string): User ID for whom to generate the token.
//   - expiry (time.Duration): Token expiry duration.
//   - secretKey (string): Secret key to sign the token.
//
// Returns:
//   - string: JWT token.
//   - error: An error if token generation fails.
func GenerateAccessToken(userID string, isSuperAdmin bool, companyIds []string, expiry time.Duration, secretKey string) (string, error) {
	return generateToken(userID, isSuperAdmin, companyIds, "access", expiry, secretKey)
}

// GenerateRefreshToken is a helper function to generate a refresh token.
//
// Parameters:
//   - userID (string): User ID for whom to generate the token.
//   - expiry (time.Duration): Token expiry duration.
//   - secretKey (string): Secret key to sign the token.
//
// Returns:
//   - string: JWT token.
//   - error: An error if token generation fails.
func GenerateRefreshToken(userID string, isSuperAdmin bool, companyIds []string, expiry time.Duration, secretKey string) (string, error) {
	return generateToken(userID, isSuperAdmin, companyIds, "refresh", expiry, secretKey)
}

// GeneratePasswordResetToken is a helper function to generate a password reset token.
//
// Parameters:
//   - userID (string): User ID for whom to generate the token.
//   - expiry (time.Duration): Token expiry duration.
//   - secretKey (string): Secret key to sign the token.
//
// Returns:
//   - string: JWT token.
//   - error: An error if token generation fails.
func GeneratePasswordResetToken(userID string, isSuperAdmin bool, companyIds []string, expiry time.Duration, secretKey string) (string, error) {
	return generateToken(userID, isSuperAdmin, companyIds, "password_reset", expiry, secretKey)
}

// GenerateActivationToken is a helper function to generate an activation token.
//
// Parameters:
//   - userID (string): User ID for whom to generate the token.
//   - expiry (time.Duration): Token expiry duration.
//   - secretKey (string): Secret key to sign the token.
//
// Returns:
//   - string: JWT token.
//   - error: An error if token generation fails.
func GenerateActivationToken(userID string, isSuperAdmin bool, companyIds []string, expiry time.Duration, secretKey string) (string, error) {
	return generateToken(userID, isSuperAdmin, companyIds, "activation", expiry, secretKey)
}

// VerifyAccessToken is a helper function to verify an access token.
//
// Parameters:
//   - tokenString (string): JWT token to verify.
//   - secretKey (string): Secret key to verify the token.
//
// Returns:
//   - jwt.MapClaims: Claims of the token.
//   - error: An error if token verification fails.
func VerifyAccessToken(tokenString string, secretKey string) (jwt.MapClaims, error) {
	return verifyToken(tokenString, "access", secretKey)
}

// VerifyRefreshToken is a helper function to verify a refresh token.
//
// Parameters:
//   - tokenString (string): JWT token to verify.
//   - secretKey (string): Secret key to verify the token.
//
// Returns:
//   - jwt.MapClaims: Claims of the token.
//   - error: An error if token verification fails.
func VerifyRefreshToken(tokenString string, secretKey string) (jwt.MapClaims, error) {
	return verifyToken(tokenString, "refresh", secretKey)
}

// VerifyPasswordResetToken is a helper function to verify a password reset token.
//
// Parameters:
//   - tokenString (string): JWT token to verify.
//   - secretKey (string): Secret key to verify the token.
//
// Returns:
//   - jwt.MapClaims: Claims of the token.
//   - error: An error if token verification fails.
func VerifyPasswordResetToken(tokenString string, secretKey string) (jwt.MapClaims, error) {
	return verifyToken(tokenString, "password_reset", secretKey)
}

// VerifyActivationToken is a helper function to verify an activation token.
//
// Parameters:
//   - tokenString (string): JWT token to verify.
//   - secretKey (string): Secret key to verify the token.
//
// Returns:
//   - jwt.MapClaims: Claims of the token.
//   - error: An error if token verification fails.
func VerifyActivationToken(tokenString string, secretKey string) (jwt.MapClaims, error) {
	return verifyToken(tokenString, "activation", secretKey)
}

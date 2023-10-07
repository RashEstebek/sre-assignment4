package utils

import (
	"github.com/golang-jwt/jwt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dreamteam/auth-service/pkg/models"
)

func TestGenerateToken(t *testing.T) {
	issuer := "example.com"
	secretKey := "mysecretkey"
	expirationHours := 24
	jwtWrapper := &JwtWrapper{
		Issuer:          issuer,
		SecretKey:       secretKey,
		ExpirationHours: int64(expirationHours),
	}

	user := models.User{
		Id:    1,
		Email: "test@example.com",
	}

	signedToken, err := jwtWrapper.GenerateToken(user)

	assert.NoError(t, err, "unexpected error while generating token")

	parsedToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	assert.NoError(t, err, "unexpected error while parsing token")
	assert.True(t, parsedToken.Valid, "generated token is invalid")

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok, "failed to parse token claims")
	assert.Equal(t, float64(user.Id), claims["Id"], "unexpected token claim value for 'Id'")
	assert.Equal(t, user.Email, claims["Email"], "unexpected token claim value for 'Email'")

	expirationClaim, ok := claims["exp"].(float64)
	assert.True(t, ok, "failed to parse token expiration claim")
	assert.Greater(t, expirationClaim, float64(time.Now().Unix()), "token has expired")
	assert.LessOrEqual(t, expirationClaim, float64(time.Now().Add(time.Hour*time.Duration(expirationHours)).Unix()), "token expiration claim is too far in the future")
}

func TestValidateToken(t *testing.T) {
	secretKey := "mysecretkey"
	jwtWrapper := &JwtWrapper{
		SecretKey: secretKey,
	}

	validToken, _ := jwtWrapper.GenerateToken(models.User{
		Id:    1,
		Email: "test@example.com",
	})

	claims, err := jwtWrapper.ValidateToken(validToken)

	assert.NoError(t, err, "unexpected error while validating token")
	assert.NotNil(t, claims, "unexpected nil claims")

	expiredClaims := jwt.MapClaims{
		"Id":    2,
		"Email": "expired@example.com",
		"exp":   time.Now().Add(-time.Hour).Unix(),
	}
	expiredTokenString, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte(secretKey))

	claims, err = jwtWrapper.ValidateToken(expiredTokenString)

	assert.Error(t, err, "expected error while validating expired token")
	assert.Nil(t, claims, "unexpected non-nil claims for expired token")
}

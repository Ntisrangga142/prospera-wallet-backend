package pkg

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId int    `json:"id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewJWTClaims(userid int, email string) *Claims {
	return &Claims{
		UserId: userid,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60)),
			Issuer:    os.Getenv("JWT_ISSUER"),
		},
	}
}

func (c *Claims) GenToken() (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("no secret found")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(jwtSecret))
}

func (c *Claims) VerifyToken(token string) error {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("no secret found")
	}
	parsedToken, err := jwt.ParseWithClaims(token, c, func(t *jwt.Token) (any, error) { return []byte(jwtSecret), nil })
	if err != nil {
		return err
	}
	if !parsedToken.Valid {
		return jwt.ErrTokenExpired
	}
	iss, err := parsedToken.Claims.GetIssuer()
	if err != nil {
		return err
	}
	if iss != os.Getenv("JWT_ISSUER") {
		return jwt.ErrTokenInvalidIssuer
	}
	return nil
}

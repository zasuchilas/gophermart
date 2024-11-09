package chisrv

import (
	"errors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/pkg/converters"
	"net/http"
	"time"
)

var tokenAuth *jwtauth.JWTAuth

func InitJWT() {
	tokenAuth = jwtauth.New("HS256", []byte(config.SecretKey), nil /*jwt.WithAcceptableSkew(time.Hour)*/)
}

func makeToken(userID int64) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{
		"userID": userID,
		"exp":    time.Now().Add(100 * 24 * time.Hour).Unix(),
	})
	return tokenString
}

func getUserID(r *http.Request) (int64, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return 0, err
	}
	userID, ok := claims["userID"]
	if !ok {
		return 0, errors.New("userID not found in token")
	}
	return converters.InterfaceToInt64(userID)
}

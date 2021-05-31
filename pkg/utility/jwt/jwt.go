package jwt

import (
	"ceres/pkg/config/auth"
	"fmt"
	"strconv"
	"time"

	JWT "github.com/dgrijalva/jwt-go"
)

// Sign jwt token for current uin
func Sign(uin uint64) (token string) {
	jwt := JWT.New(JWT.SigningMethodHS256)
	claims := make(JWT.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(72)).Unix()
	claims["iat"] = time.Now().Unix()
	claims["comer_uin"] = fmt.Sprintf("%d", uin)
	jwt.Claims = claims
	token, _ = jwt.SignedString(auth.JWTSecret)
	return
}

// Verify jwt token if success then return the uin
func Verify(token string) (uin uint64, err error) {
	auth, err := JWT.Parse(token, func(t *JWT.Token) (interface{}, error) {
		return auth.JWTSecret, nil
	})
	if err != nil {
		return
	}
	claims, _ := auth.Claims.(JWT.MapClaims)
	uinStr, _ := claims["comer_uin"].(string)
	uin, err = strconv.ParseUint(uinStr, 10, 64)
	if err != nil {
		return
	}
	return
}

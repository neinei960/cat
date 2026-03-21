package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/neinei960/cat/server/config"
)

type Claims struct {
	StaffID uint   `json:"staff_id"`
	ShopID  uint   `json:"shop_id"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

type CustomerClaims struct {
	CustomerID uint   `json:"customer_id"`
	ShopID     uint   `json:"shop_id"`
	OpenID     string `json:"openid"`
	jwt.RegisteredClaims
}

func GenerateToken(staffID, shopID uint, role string) (string, error) {
	expireHour := config.AppConfig.JWT.ExpireHour
	if expireHour == 0 {
		expireHour = 72
	}
	claims := Claims{
		StaffID: staffID,
		ShopID:  shopID,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func GenerateCustomerToken(customerID, shopID uint, openID string) (string, error) {
	expireHour := config.AppConfig.JWT.ExpireHour
	if expireHour == 0 {
		expireHour = 72
	}
	claims := CustomerClaims{
		CustomerID: customerID,
		ShopID:     shopID,
		OpenID:     openID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

func ParseCustomerToken(tokenString string) (*CustomerClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomerClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*CustomerClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

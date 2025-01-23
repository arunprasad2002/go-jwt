package helpers

import (
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SingnedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

func GenerateAllTokens(email string, firstName string, lastName string, userType string, uid string) (singedToken string, singedRefreshToken string, err error) {
	claims := &SingnedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		User_type:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	refreshClaims := &SingnedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString([]byte(""))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims).SignedString([]byte(""))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

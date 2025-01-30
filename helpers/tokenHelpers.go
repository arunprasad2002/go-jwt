package helpers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arunprasad2002/go-jwt/database"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SingnedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

type SignedDetails struct {
	Email      string `json:"email"`
	First_name string `json:"first_name"`
	Last_name  string `json:"last_name"`
	Uid        string `json:"uid"`
	User_type  string `json:"user_type"`
	jwt.StandardClaims
}

var SECRET_KEY = "your-secret-key" // Replace with a strong secret key

func GenerateAllTokens(email string, firstName string, lastName string, userType string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		User_type:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(168 * time.Hour).Unix(), // Refresh token expires in 7 days
		},
	}

	// Create access token
	token, tokenErr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if tokenErr != nil {
		log.Panic(tokenErr)
		return "", "", tokenErr
	}

	// Create refresh token
	refreshToken, refreshErr := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if refreshErr != nil {
		log.Panic(refreshErr)
		return "", "", refreshErr
	}

	return token, refreshToken, nil
}

func ValidateToken(singedToken string) (claims *SingnedDetails, msg string) {

	token, err := jwt.ParseWithClaims(
		singedToken,
		&SingnedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(""), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SingnedDetails)

	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}

	return claims, msg
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId *string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}
	return
}

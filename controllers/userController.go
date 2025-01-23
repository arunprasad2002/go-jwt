package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arunprasad2002/go-jwt/database"
	"github.com/arunprasad2002/go-jwt/helpers"
	"github.com/arunprasad2002/go-jwt/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var context, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		defer cancel()

		err := ctx.BindJSON(&user)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validateErr := validate.Struct(user)

		if validateErr != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		userEmailCount, err := userCollection.CountDocuments(context, bson.M{"emal": user.Email})
		defer cancel()

		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		userPhoneCount, err := userCollection.CountDocuments(context, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if userEmailCount > 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "email already exist"})
			return
		}

		if userPhoneCount > 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "phone already exist"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		userID := user.ID.Hex()
		user.User_id = &userID
		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken
		resultInsertionNumber, insertErr := userCollection.InsertOne(context, user)

		if insertErr != nil {
			msg := fmt.Sprintf("User is not created")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		defer cancel()
		ctx.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")
		err := helpers.MatchUserToUid(ctx, userId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		var context, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		err = userCollection.FindOne(context, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, user)
	}
}

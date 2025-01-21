package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/arunprasad2002/go-jwt/database"
	"github.com/arunprasad2002/go-jwt/helpers"
	"github.com/arunprasad2002/go-jwt/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validator = validator.New()

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

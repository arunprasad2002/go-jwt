package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arunprasad2002/go-jwt/database"
	"github.com/arunprasad2002/go-jwt/helpers"
	"github.com/arunprasad2002/go-jwt/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("login or password is incorrect")
		check = false
	}

	return check, msg
}

func SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Only once

		var user models.User

		err := ctx.BindJSON(&user)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validateErr := validate.Struct(user)
		if validateErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validateErr.Error()})
			return
		}

		// Check if email exists
		userEmailCount, err := userCollection.CountDocuments(ctxTimeout, bson.M{"email": user.Email})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Check if phone exists
		userPhoneCount, err := userCollection.CountDocuments(ctxTimeout, bson.M{"phone": user.Phone})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if userEmailCount > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}

		if userPhoneCount > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Phone already exists"})
			return
		}

		// Hash password
		password := HashPassword(*user.Password)
		user.Password = &password

		// Set timestamps
		user.Created_at = time.Now()
		user.Updated_at = time.Now()
		user.ID = primitive.NewObjectID()
		userID := user.ID.Hex()
		user.User_id = &userID

		// Generate JWT tokens
		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		// Insert user into DB
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctxTimeout, user)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User could not be created"})
			return
		}

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

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helpers.ChekcUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}
		var allusers []bson.M
		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allusers[0])
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("Step 1: Received Login Request")

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		// Bind JSON input
		if err := c.BindJSON(&user); err != nil {
			fmt.Println("Error parsing request body:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure email is not nil
		if user.Email == nil {
			fmt.Println("Step 2: Email is nil")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}
		fmt.Println("Step 2: Received email:", *user.Email)

		// Query the database for the user
		fmt.Println("Step 3: Searching for user in DB")
		err := userCollection.FindOne(ctx, bson.M{"email": *user.Email}).Decode(&foundUser)
		if err != nil {
			fmt.Println("Step 3 Error: User not found in DB", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or password is incorrect"})
			return
		}
		fmt.Println("Step 3: User found in DB:", foundUser.Email)

		// Ensure password is not nil
		if user.Password == nil {
			fmt.Println("Step 4: Password is nil")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
			return
		}
		fmt.Println("Step 4: Checking password")

		// Verify the password
		passwordIsValid, msg := helpers.VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			fmt.Println("Step 4 Error: Password incorrect", msg)
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}
		fmt.Println("Step 4: Password verified successfully")

		// Ensure user_type is not nil
		if foundUser.User_type == nil {
			fmt.Println("Step 5 Error: User type is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User type not found"})
			return
		}

		// Generate tokens
		fmt.Println("Step 5: Generating tokens")
		token, refreshToken, err := helpers.GenerateAllTokens(
			*foundUser.Email,
			*foundUser.First_name,
			*foundUser.Last_name,
			*foundUser.User_type,
			*foundUser.User_id,
		)
		if err != nil {
			fmt.Println("Step 5 Error: Token generation failed", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
			return
		}
		fmt.Println("Step 5: Tokens generated successfully")

		// Update user tokens in DB
		fmt.Println("Step 6: Updating tokens in DB")
		helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		// Fetch updated user
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
		if err != nil {
			fmt.Println("Step 6 Error: Failed to retrieve updated user", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("Step 6: User updated successfully")

		// Hide password before sending response
		foundUser.Password = nil

		// Send success response
		fmt.Println("Step 7: Login successful")
		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"user":    foundUser,
		})
	}
}

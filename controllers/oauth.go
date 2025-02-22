package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/arunprasad2002/go-jwt/helpers"
	"github.com/arunprasad2002/go-jwt/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Load credentials from environment variables
var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

var oauthStateString = "random-string"

// GoogleLogin redirects users to Google’s authentication page
func GoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusFound, url)
}

// GoogleCallback handles the callback from Google after authentication
func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
		return
	}

	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Fetch user info from Google
	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	email, emailOk := userInfo["email"].(string)
	firstName, firstNameOk := userInfo["given_name"].(string)
	lastName, lastNameOk := userInfo["family_name"].(string)

	if !emailOk || !firstNameOk || !lastNameOk {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data from Google"})
		return
	}

	// Check if user exists in DB
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var foundUser models.User
	err = userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&foundUser)

	if err != nil {
		// Create new user if not found
		newUser := models.User{
			Email:      &email,
			First_name: &firstName,
			Last_name:  &lastName,
			User_type:  stringPointer("user"), // Default user type
			User_id:    stringPointer(primitive.NewObjectID().Hex()),
		}

		_, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		foundUser = newUser
	}

	// Generate JWT tokens
	tokenStr, refreshToken, err := helpers.GenerateAllTokens(
		*foundUser.Email,
		*foundUser.First_name,
		*foundUser.Last_name,
		*foundUser.User_type,
		*foundUser.User_id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	// Update tokens in DB
	helpers.UpdateAllTokens(tokenStr, refreshToken, foundUser.User_id)

	// Redirect user to frontend with tokens (or store in cookies)
	redirectURL := fmt.Sprintf("http://localhost:3000/auth/callback?token=%s&refreshToken=%s", tokenStr, refreshToken)
	c.Redirect(http.StatusFound, redirectURL)
}

// Helper function to create string pointers
func stringPointer(s string) *string {
	return &s
}

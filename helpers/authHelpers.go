package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ChekcUserType(ctx *gin.Context, role string) (err error) {
	userType := ctx.GetString("user_type")
	err = nil

	if userType != role {
		err = errors.New("Unauthorize to access this resource")
		return err
	}

	return err
}

func MatchUserToUid(ctx *gin.Context, userId string) (err error) {
	userType := ctx.GetString("user_type")
	uid := ctx.GetString("uid")
	if userType == "USER" && uid != userId {
		err = errors.New("Unauthorize to access this resource")
		return err
	}
	err = ChekcUserType(ctx, userType)
	return err
}

// VerifyPassword checks if the provided password matches the hashed password
func VerifyPassword(providedPassword, storedHashedPassword string) (bool, string) {
	// Compare the hashed password with the provided password
	err := bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(providedPassword))
	if err != nil {
		return false, "Email or password is incorrect"
	}
	return true, "Password verified successfully"
}

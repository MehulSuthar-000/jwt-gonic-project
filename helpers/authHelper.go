package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func MatchUserTypeToUid(ctx *gin.Context, userId string) error {
	userType := ctx.GetString("user_type")
	uid := ctx.GetString("uid")

	if userType == "USER" && uid != userId {
		return errors.New("Unauthorized to access this resource")
	}

	err := CheckUserType(ctx, userType)
	return err
}

func CheckUserType(ctx *gin.Context, role string) error {
	userType := ctx.GetString("user_type")
	if userType != role {
		return errors.New("Unauthorized to access this resource")
	}

	return nil
}

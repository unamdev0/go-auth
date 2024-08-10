package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func MatchUserTypeToUID(ctx *gin.Context, userID string) (err error) {
	userType := ctx.GetString("user_type")
	uid := ctx.GetString("uid")

	err = nil

	if userType == "USER" && uid != userID {
		err = errors.New("unauthroized to access User data")
		return err
	}

	return err
}

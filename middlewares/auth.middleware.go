package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	helper "github.com/unamdev0/go-auth/helpers"
)

func Authenticate() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		clientToken := ctx.Request.Header.Get("token")

		if clientToken == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No token found"})
			ctx.Abort()
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			ctx.Abort()
			return
		}
		ctx.Set("email", claims.Email)
		ctx.Set("first_name", claims.FirstName)
		ctx.Set("last_name", claims.LastName)
		ctx.Set("uid", claims.UID)
		ctx.Set("user_type", claims.UserType)
		ctx.Next()

	}

}

package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/unamdev0/go-auth/database"
	helper "github.com/unamdev0/go-auth/helpers"
	"github.com/unamdev0/go-auth/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var validate = validator.New()

func GetUser() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		userId := ctx.Param("id")
		if err := helper.MatchUserTypeToUID(ctx, userId); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		anotherContext, cancel := context.WithTimeout(context.TODO(), 100*time.Second)

		defer cancel()
		var user *models.User

		err := UserCollection.FindOne(anotherContext, bson.M{"_id": userId}).Decode(&user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, user)
		return
	}

}

func Signup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		anotherContext, cancel := context.WithTimeout(context.TODO(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationError := validate.Struct(user)
		if validationError != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
		}

		count, err := UserCollection.CountDocuments(anotherContext, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		}

		if count > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user with email id already exists!!"})
		}

		count, err = UserCollection.CountDocuments(anotherContext, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		}

		if count > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user with this phone  already exists!!"})
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		token, refreshToken, _ := helper.GenerateAllTokens(*user)
		user.Token = &token
		user.RefreshToken = &refreshToken

		resultInsertionNumber, err := UserCollection.InsertOne(anotherContext, user)

		if err != nil {
			msg := fmt.Sprintf("error while inserting record in DB")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		ctx.JSON(http.StatusOK, resultInsertionNumber)
		return
	}

}

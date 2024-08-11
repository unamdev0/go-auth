package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

	}

}

func Signup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		anotherContext, cancel := context.WithTimeout(context.TODO(), 100*time.Second)
		defer cancel()

		var user *models.User
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationError := validate.Struct(user)
		if validationError != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
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
		password, _ := helper.HashPassword(*user.Password)
		user.Password = &password
		if count > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user with this phone  already exists!!"})
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		token, refreshToken, _ := helper.GenerateAllTokens(user)
		user.Token = &token
		user.RefreshToken = &refreshToken

		resultInsertionNumber, err := UserCollection.InsertOne(anotherContext, user)

		if err != nil {
			msg, _ := fmt.Printf("error while inserting record in DB")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		ctx.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		anotherContext, cancel := context.WithTimeout(ctx, 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := UserCollection.FindOne(anotherContext, bson.M{"email": user.Email}).Decode(&foundUser)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		isValidPassword, err := helper.VerifyPassword(*user.Password, *foundUser.Password)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !isValidPassword {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't compare password "})
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(&foundUser)

		helper.UpdateAllTokens(foundUser, token, refreshToken)
		err = UserCollection.FindOne(anotherContext, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if userType := ctx.GetString("user_type"); userType != "ADMIN" {
			ctx.JSON(http.StatusNonAuthoritativeInfo, gin.H{"error": "Unauthorized access"})
		}

		anotherContext, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		limit, err := strconv.Atoi(ctx.Query("limit"))

		if err != nil || limit < 1 {
			limit = 10
		}

		page, err := strconv.Atoi(ctx.Query("page"))

		if err != nil || limit < 1 {
			page = 1
		}

		startIndex := (page - 1) * limit

		//Not putting any matching condition
		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0}, {Key: "total_count", Value: 1}, {Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, limit}}}}}}}

		result, err := UserCollection.Aggregate(anotherContext, mongo.Pipeline{matchStage, groupStage, projectStage})

		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var allUsers []bson.M
		if err = result.All(anotherContext, &allUsers); err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, allUsers[0])
	}

}

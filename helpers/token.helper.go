package helper

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt"
	"github.com/unamdev0/go-auth/database"
	"github.com/unamdev0/go-auth/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type SignedToken struct {
	Email     string
	FirstName string
	LastName  string
	UID       string
	UserType  string
	jwt.StandardClaims
}

var UserCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY = os.Getenv("SECRET_KEY")

func GenerateAllTokens(user *models.User) (token string, refreshToken string, err error) {
	claims := &SignedToken{
		Email:     *user.Email,
		FirstName: *user.FirstName,
		LastName:  *user.LastName,
		UID:       user.ID.Hex(),
		UserType:  *user.UserType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix()},
	}

	refreshClaims := &SignedToken{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix()},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return "", "", err
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return "", "", err
	}

	return token, refreshToken, err
}

func VerifyPassword(passwordEntered string, passwordInDB string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(passwordInDB), []byte(passwordEntered)); err != nil {
		return false, errors.New("password is incorrect")
	}
	return true, nil

}

func HashPassword(password string) (string, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", errors.New("error generating hashed password")
	}
	return string(hashedPassword), err
}

func UpdateAllTokens(user models.User, token string, refreshToken string) {

	filter := bson.D{primitive.E{Key: "_id", Value: user.ID}}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "refresh_token", Value: refreshToken}, primitive.E{Key: "token", Value: token}}}}
	_, err := UserCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Panic(err)
	}

}

func ValidateToken(tokenInHeader string) (claims *SignedToken, msg string) {

	token, err := jwt.ParseWithClaims(tokenInHeader, &SignedToken{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedToken)
	if !ok {
		msg = ("the token is invalid")

		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = ("token is expired")

		return
	}
	return claims, msg
}

// func ValidateToken(signedToken string) (claims *SignedToken, msg string){
// 	token, err := jwt.ParseWithClaims(
// 		signedToken,
// 		&SignedToken{},
// 		func(token *jwt.Token)(interface{}, error){
// 			return []byte(SECRET_KEY), nil
// 		},
// 	)

// 	if err != nil {
// 		msg=err.Error()
// 		return
// 	}

// 	claims, ok:= token.Claims.(*SignedDetails)
// 	if !ok{
// 		msg = fmt.Sprintf("the token is invalid")
// 		msg = err.Error()
// 		return
// 	}

// 	if claims.ExpiresAt < time.Now().Local().Unix(){
// 		msg = fmt.Sprintf("token is expired")
// 		msg = err.Error()
// 		return
// 	}
// 	return claims, msg
// }

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

package helper

import (
	"os"

	jwt "github.com/golang-jwt/jwt"
	"github.com/unamdev0/go-auth/database"
	"github.com/unamdev0/go-auth/models"
	"go.mongodb.org/mongo-driver/mongo"
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

func GenerateAllTokens(user *models.User)

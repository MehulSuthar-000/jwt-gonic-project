package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitub.com/mehulsuthar-000/golang-jwt-project/database"
	"gitub.com/mehulsuthar-000/golang-jwt-project/helpers"
	"gitub.com/mehulsuthar-000/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HasPassword()

func VerifyPassword()

func Signup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": validationErr.Error(),
				},
			)
		}

		count, err := userCollection.CountDocuments(ctxWithTimeout, bson.M{
			"email": user.Email,
		})
		if err != nil {
			log.Panic(err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "error occured while checking for email",
				},
			)
		}

		count, err = userCollection.CountDocuments(ctxWithTimeout, bson.M{
			"phone": user.Phone,
		})
		if err != nil {
			log.Panic(err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "error occured while checking for phone number",
				},
			)
		}

		if count > 0 {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "this email or phone number is already exists",
				},
			)
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *&user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctxWithTimeout, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": msg,
				},
			)
		}
		ctx.JSON(
			http.StatusOK,
			resultInsertionNumber,
		)
	}
}

func Login()

func GetUsers()

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")

		if err := helpers.MatchUserTypeToUid(ctx, userId); err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)
		}

		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		err := userCollection.FindOne(ctxWithTimeout, bson.M{
			"user_id": userId,
		}).Decode(&user)
		if err != nil {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": err.Error(),
				},
			)
			return
		}

		ctx.JSON(
			http.StatusOK,
			user,
		)

	}
}

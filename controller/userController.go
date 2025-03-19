package controller

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitub.com/mehulsuthar-000/golang-jwt-project/database"
	"gitub.com/mehulsuthar-000/golang-jwt-project/helpers"
	"gitub.com/mehulsuthar-000/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedpassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "email or password is incorrect"
		check = false
	}

	return check, msg
}

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
		if count > 0 {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "this email or phone number is already exists",
				},
			)
		}

		password := HashPassword(*user.Password)
		user.Password = &password

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

		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctxWithTimeout, user)
		if insertErr != nil {
			msg := "User item was not created"
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

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)
		}

		err := userCollection.FindOne(ctxWithTimeout, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "email or password is incorrect",
				},
			)
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": msg,
				},
			)
			return
		}

		if foundUser.Email == nil {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "user not found",
				},
			)
		}

		token, refreshToken, err := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		if err != nil {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": err.Error(),
				},
			)
		}

		helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		err = userCollection.FindOne(ctxWithTimeout, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
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
			foundUser,
		)

	}
}

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := helpers.CheckUserType(ctx, "ADMIN"); err != nil {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": err.Error(),
				},
			)
			return
		}
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex, err := strconv.Atoi(ctx.Query("startIndex"))
		if err != nil {
			startIndex = (page - 1) * recordPerPage
		}

		matchStage := bson.D{
			{Key: "$match", Value: bson.D{{}}},
		}

		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil}, // Grouping all documents together
				{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "all_documents", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}, // Correct key for $push
			}},
		}

		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: bson.A{"$all_documents", startIndex, recordPerPage}},
				}},
			}},
		}

		result, err := userCollection.Aggregate(ctxWithTimeout, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})
		if err != nil {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "error occured while listing user itmes",
				},
			)
		}

		var allUsers []bson.M
		if err := result.All(ctxWithTimeout, &allUsers); err != nil {
			log.Fatal(err)
		}

		ctx.JSON(
			http.StatusOK,
			allUsers[0],
		)
	}
}

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

		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
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

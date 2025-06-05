package main

import (
	"context"
	"fmt"
	"github.com/Moji00f/recipe-api/docs"
	"github.com/Moji00f/recipe-api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
)

var collection *mongo.Collection

// @Summary Search recipes
// @Description Search recipes based on tags
// @Tags recipes
// @Accept json
// @Produce json
// @Param tag query string true "Recipe tag"
// @Success 200 {array} Recipe
// @Router /recipes/search [get]
func SearchRecipesHandler(c *gin.Context) {}

var recipesHandler *handlers.RecipesHandler
var client *mongo.Client
var ctx context.Context

func init() {
	ctx = context.Background()
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connect to Mongodb")

	//var listOfRecipes []interface{}
	//for _, recipe := range recipes {
	//	listOfRecipes = append(listOfRecipes, recipe)
	//}

	//collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	//insertManyResult, err := collection.InsertMany(context.Background(), listOfRecipes)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))

	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	recipesHandler = handlers.NewRecipesHandler(ctx, collection)
}

func main() {
	router := gin.Default()
	registerSwagger(router)
	router.POST("/recipes", recipesHandler.NewRecipeHandler)
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	//router.GET("/recipes/search", SearchRecipesHandler)
	router.Run(":6060")

	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}(client, ctx)
}

func registerSwagger(r *gin.Engine) {
	docs.SwaggerInfo.Title = "Golang Recipe Api"
	docs.SwaggerInfo.Description = "Golang Recipe Api"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Schemes = []string{"http"}
	docs.SwaggerInfo.Host = "localhost:6060"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Moji00f/recipe-api/docs"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// @Parameters recipes newRecipe
type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

var collection *mongo.Collection

// @Summary Create a new recipe
// @Description Adds a new recipe to the database
// @Tags recipes
// @Accept json
// @Produce json
// @Param recipe body Recipe true "Recipe data"
// @Success 200 {object} Recipe
// @Failure 400
// @Router /recipes [post]
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err = collection.InsertOne(ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new recipe"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

// @Summary List recipes
// @Description Returns a list of all available recipes
// @Tags recipes
// @Accept json
// @Produce json
// @Success 200 {array} Recipe
// @Router /recipes [get]
func ListRecipesHandler(c *gin.Context) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {

		}
	}(cur, ctx)

	recipes := make([]Recipe, 0)
	for cur.Next(ctx) {
		var recipe Recipe
		err := cur.Decode(&recipe)
		if err != nil {
			return
		}
		recipes = append(recipes, recipe)
	}

	c.JSON(http.StatusOK, recipes)
}

// @Summary Update an existing recipe
// @Description Modify recipe details based on the provided ID
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Param recipe body Recipe true "Updated recipe data"
// @Success 200 {object} Recipe
// @Failure 400
// @Failure 404
// @Router /recipes/{id} [put]
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err = collection.UpdateOne(ctx,
		bson.M{"_id": objectId}, bson.D{{"$set", bson.D{
			{"name", recipe.Name},
			{"instructions", recipe.Instructions},
			{"ingredients", recipe.Ingredients},
			{"tags", recipe.Tags},
		}}})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

// @Summary Delete an existing recipe
// @Description Remove a recipe from the database using its ID
// @Tags recipes
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 200
// @Failure 404
// @Router /recipes/{id} [delete]
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1
	for i := 0; i < len(recipes); i++ {
		//if recipes[i].ID == id {
		//	index = i
		//}
		fmt.Println(id)
		break
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found"})
		return
	}

	recipes = append(recipes[:index], recipes[index+1:]...)

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted"})
}

// @Summary Search recipes
// @Description Search recipes based on tags
// @Tags recipes
// @Accept json
// @Produce json
// @Param tag query string true "Recipe tag"
// @Success 200 {array} Recipe
// @Router /recipes/search [get]
func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)

	for i := 0; i < len(recipes); i++ {
		found := false
		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				found = true
			}
		}
		if found {
			listOfRecipes = append(listOfRecipes, recipes[i])
		}
	}

	c.JSON(http.StatusOK, listOfRecipes)
}

var recipes []Recipe
var ctx context.Context
var err error
var client *mongo.Client

func init() {
	recipes = make([]Recipe, 0)
	file, _ := os.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)

	if err = godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err = mongo.Connect(ctx, clientOptions)
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
}

func main() {
	router := gin.Default()
	registerSwagger(router)
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
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

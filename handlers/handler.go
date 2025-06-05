package handlers

import (
	"context"
	"fmt"
	"github.com/Moji00f/recipe-api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

type RecipesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection) *RecipesHandler {
	return &RecipesHandler{
		collection: collection,
		ctx:        ctx,
	}
}

// @Summary List recipes
// @Description Returns a list of all available recipes
// @Tags recipes
// @Accept json
// @Produce json
// @Success 200 {array} Recipe
// @Router /recipes [get]
func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {

		}
	}(cur, handler.ctx)

	recipes := make([]models.Recipe, 0)
	for cur.Next(handler.ctx) {
		var recipe models.Recipe
		err := cur.Decode(&recipe)
		if err != nil {
			return
		}
		recipes = append(recipes, recipe)
	}

	c.JSON(http.StatusOK, recipes)
}

// @Summary Create a new recipe
// @Description Adds a new recipe to the database
// @Tags recipes
// @Accept json
// @Produce json
// @Param recipe body Recipe true "Recipe data"
// @Success 200 {object} Recipe
// @Failure 400
// @Router /recipes [post]
func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new recipe"})
		return
	}
	c.JSON(http.StatusOK, recipe)
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
func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx,
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
func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
}

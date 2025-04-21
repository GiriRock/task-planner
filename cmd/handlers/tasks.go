package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/girirock/task-planner/cmd/models"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
)

// func GetTasks(ctx context.Context, db *mongo.Database) ([]*Task, error) {
func GetTasks(ctx echo.Context) error {
	c := ctx.Request().Context()
	clientOpts := options.Client().ApplyURI(
		fmt.Sprintf("%v", os.Getenv("DB_CONN")))
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Panic(err)
	}

	var tasks []*models.Task
	cur, err := client.Database("task-planner").Collection("tasks").Find(c, bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(c)
	defer client.Disconnect(c)
	for cur.Next(c) {
		var task models.Task
		err := cur.Decode(&task)
		if err != nil {
			return err
		}
		tasks = append(tasks, &task)
	}
	taksEncoded, err := json.Marshal(tasks)
	return ctx.JSON(200, string(taksEncoded))
}

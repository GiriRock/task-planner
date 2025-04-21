package handlers

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/girirock/task-planner/cmd/models"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

	var tasks []models.Task
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
		tasks = append(tasks, task)
	}
	return ctx.Render(200, "tasks", tasks)
}

func DeleteTask(ctx echo.Context) error {
	c := ctx.Request().Context()
	id := ctx.QueryParam("id")
	//url decode the id
	decodedID, err := url.QueryUnescape(id)
	if err != nil {
		return err
	}
	// remove string ObjectID from the id
	println(decodedID)
	_id, err := bson.ObjectIDFromHex(decodedID)
	if err != nil {
		log.Panic(err)
	}
	clientOpts := options.Client().ApplyURI(
		fmt.Sprintf("%v", os.Getenv("DB_CONN")))
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Panic(err)
	}
	defer client.Disconnect(c)
	var decodedTask models.Task
	task := client.Database("task-planner").Collection("tasks").FindOne(c, bson.M{"_id": _id})
	task.Decode(&decodedTask)
	fmt.Println(decodedTask)
	//_, err = client.Database("task-planner").Collection("tasks").DeleteOne(c, bson.M{"_id": decodedID})
	return GetTasks(ctx)
}

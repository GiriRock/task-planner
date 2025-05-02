package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

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
		log.Fatal(err)
	}
	accessToken, _ := ctx.Cookie("access_token")
	User, err := DecodeAccessToken(accessToken.Value)
	if err != nil {
		return ctx.Redirect(http.StatusFound, "/google-auth")
	}

	var tasks []models.Task
	cur, err := client.Database("task-planner").Collection("tasks").Find(c, bson.M{"uid": User.UID})
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
	re := regexp.MustCompile(`ObjectID\("([a-f0-9]+)"\)`)
	matches := re.FindStringSubmatch(decodedID)
	println(matches[1])
	_id, err := bson.ObjectIDFromHex(matches[1])
	if err != nil {
		log.Fatal(err)
	}
	clientOpts := options.Client().ApplyURI(
		fmt.Sprintf("%v", os.Getenv("DB_CONN")))
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(c)
	_, err = client.Database("task-planner").Collection("tasks").DeleteOne(c, bson.M{"_id": _id})
	return GetTasks(ctx)
}

func AddTask(ctx echo.Context) error {
	//TODO: implement add task
	c := ctx.Request().Context()
	clientOpts := options.Client().ApplyURI(
		fmt.Sprintf("%v", os.Getenv("DB_CONN")))
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	accessToken, _ := ctx.Cookie("access_token")
	User, err := DecodeAccessToken(accessToken.Value)
	if err != nil {
		return ctx.Redirect(http.StatusFound, "/google-auth")
	}
	var task models.Task
	if err := ctx.Bind(&task); err != nil {
		return err
	}
	task.UID = User.UID
	task.Completed = false
	defer client.Disconnect(c)
	_, err = client.Database("task-planner").Collection("tasks").InsertOne(c, task)
	return GetTasks(ctx)
}

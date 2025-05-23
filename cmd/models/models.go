package models

import "go.mongodb.org/mongo-driver/v2/bson"

// TODO: Change the ID to primitive.ObjectID
type Task struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Description string        `bson:"description"`
	DueDate     string        `bson:"dueDate"`
	Completed   bool          `bson:"completed"`
	UID         string        `bson:"uid"`
}

type User struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	Name    string        `bson:"name"`
	Picture string        `bson:"picture"`
	UID     string        `bson:"uid"`
}

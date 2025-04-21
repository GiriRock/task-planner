package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Task struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Description string        `bson:"description"`
	DueDate     string        `bson:"dueDate"`
	Completed   bool          `bson:"completed"`
}

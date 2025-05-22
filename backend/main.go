package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"

	"github.com/azezpz1/TeacherTools/handlers"
)

var firestoreClient *firestore.Client

func main() {

	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable not set")
	}
	client, err := firestore.NewClientWithDatabase(ctx, projectID, "teacher-tools-test")
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	firestoreClient = client
	defer firestoreClient.Close()

	r := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/student/teachers", handlers.GetTeachersForStudentHandler(client))

	r.Run(":" + port)
}

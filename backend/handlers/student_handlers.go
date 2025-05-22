package handlers

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"

	"github.com/azezpz1/TeacherTools/models"
)

func GetTeachersForStudentHandler(client *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		firstName := c.Query("firstName")
		lastName := c.Query("lastName")

		if firstName == "" || lastName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "firstName and lastName query parameters required"})
			return
		}

		// Find student
		iter := client.Collection("students").
			Where("firstName", "==", firstName).
			Where("lastName", "==", lastName).
			Limit(1).
			Documents(ctx)
		defer iter.Stop()

		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the teacherIDs field (array of references)
		rawRefs, err := doc.DataAt("teacherIDs")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing teacherIDs field"})
			return
		}

		refList, ok := rawRefs.([]interface{})
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "teacherIDs field is not an array of references"})
			return
		}

		var teachers []models.Teacher

		for _, r := range refList {
			ref, ok := r.(*firestore.DocumentRef)
			if !ok {
				continue // skip invalid refs
			}

			tDoc, err := ref.Get(ctx)
			if err != nil {
				continue // optionally handle errors
			}

			var t models.Teacher
			if err := tDoc.DataTo(&t); err != nil {
				continue
			}
			teachers = append(teachers, t)
		}

		c.JSON(http.StatusOK, teachers)
	}
}

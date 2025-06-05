package handlers

import (
	"context"
	"errors"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"

	"github.com/azezpz1/TeacherTools/models"
)

var (
	ErrStudentNotFound   = errors.New("student not found")
	ErrMissingTeacherIDs = errors.New("missing teacherIDs field")
	ErrInvalidTeacherIDs = errors.New("teacherIDs field is not an array of references")
)

var fetchTeachersForStudent = func(ctx context.Context, client *firestore.Client, firstName, lastName string) ([]models.Teacher, error) {
	iter := client.Collection("students").
		Where("firstName", "==", firstName).
		Where("lastName", "==", lastName).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	rawRefs, err := doc.DataAt("teacherIDs")
	if err != nil {
		return nil, ErrMissingTeacherIDs
	}

	refList, ok := rawRefs.([]interface{})
	if !ok {
		return nil, ErrInvalidTeacherIDs
	}

	var teachers []models.Teacher

	for _, r := range refList {
		ref, ok := r.(*firestore.DocumentRef)
		if !ok {
			continue
		}

		tDoc, err := ref.Get(ctx)
		if err != nil {
			continue
		}

		var t models.Teacher
		if err := tDoc.DataTo(&t); err != nil {
			continue
		}
		teachers = append(teachers, t)
	}

	return teachers, nil
}

func GetTeachersForStudentHandler(client *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		firstName := c.Query("firstName")
		lastName := c.Query("lastName")

		if firstName == "" || lastName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "firstName and lastName query parameters required"})
			return
		}

		teachers, err := fetchTeachersForStudent(ctx, client, firstName, lastName)
		if err != nil {
			switch err {
			case ErrStudentNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			case ErrMissingTeacherIDs:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "missing teacherIDs field"})
			case ErrInvalidTeacherIDs:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "teacherIDs field is not an array of references"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, teachers)
	}
}

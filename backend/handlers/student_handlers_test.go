package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/azezpz1/TeacherTools/models"
	"github.com/gin-gonic/gin"
)

func setupRouter(client *firestore.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/student/teachers", GetTeachersForStudentHandler(client))
	return r
}

func TestGetTeachersForStudentHandler_MissingParams(t *testing.T) {
	r := setupRouter(nil)
	req, _ := http.NewRequest("GET", "/student/teachers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetTeachersForStudentHandler_Errors(t *testing.T) {
	testCases := []struct {
		name           string
		mockError      error
		expectedStatus int
	}{
		{"Student Not Found", ErrStudentNotFound, http.StatusNotFound},
		{"Internal Server Error", errors.New("boom"), http.StatusInternalServerError},
		{"Missing Teacher IDs", ErrMissingTeacherIDs, http.StatusInternalServerError},
		{"Invalid Teacher IDs", ErrInvalidTeacherIDs, http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orig := fetchTeachersForStudent
			fetchTeachersForStudent = func(ctx context.Context, client *firestore.Client, firstName, lastName string) ([]models.Teacher, error) {
				return nil, tc.mockError
			}
			defer func() { fetchTeachersForStudent = orig }()

			r := setupRouter(nil)
			req, _ := http.NewRequest("GET", "/student/teachers?firstName=John&lastName=Doe", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetTeachersForStudentHandler_Success(t *testing.T) {
	expected := []models.Teacher{{FirstName: "Jane", LastName: "Smith"}}
	orig := fetchTeachersForStudent
	fetchTeachersForStudent = func(ctx context.Context, client *firestore.Client, firstName, lastName string) ([]models.Teacher, error) {
		return expected, nil
	}
	defer func() { fetchTeachersForStudent = orig }()

	r := setupRouter(nil)
	req, _ := http.NewRequest("GET", "/student/teachers?firstName=John&lastName=Doe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, w.Code)
	}

	var got []models.Teacher
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("unexpected body: %v", got)
	}
	if got[0].FirstName != expected[0].FirstName || got[0].LastName != expected[0].LastName {
		t.Fatalf("unexpected body: %v", got)
	}
}

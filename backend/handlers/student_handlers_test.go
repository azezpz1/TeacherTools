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

func TestGetTeachersForStudentHandler_StudentNotFound(t *testing.T) {
	orig := fetchTeachersForStudent
	fetchTeachersForStudent = func(ctx context.Context, client *firestore.Client, firstName, lastName string) ([]models.Teacher, error) {
		return nil, ErrStudentNotFound
	}
	defer func() { fetchTeachersForStudent = orig }()

	r := setupRouter(nil)
	req, _ := http.NewRequest("GET", "/student/teachers?firstName=John&lastName=Doe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetTeachersForStudentHandler_InternalError(t *testing.T) {
	orig := fetchTeachersForStudent
	fetchTeachersForStudent = func(ctx context.Context, client *firestore.Client, firstName, lastName string) ([]models.Teacher, error) {
		return nil, errors.New("boom")
	}
	defer func() { fetchTeachersForStudent = orig }()

	r := setupRouter(nil)
	req, _ := http.NewRequest("GET", "/student/teachers?firstName=John&lastName=Doe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetTeachersForStudentHandler_MissingTeacherIDs(t *testing.T) {
	orig := fetchTeachersForStudent
	fetchTeachersForStudent = func(ctx context.Context, client *firestore.Client, firstName, lastName string) ([]models.Teacher, error) {
		return nil, ErrMissingTeacherIDs
	}
	defer func() { fetchTeachersForStudent = orig }()

	r := setupRouter(nil)
	req, _ := http.NewRequest("GET", "/student/teachers?firstName=John&lastName=Doe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetTeachersForStudentHandler_InvalidTeacherIDs(t *testing.T) {
	orig := fetchTeachersForStudent
	fetchTeachersForStudent = func(ctx context.Context, client *firestore.Client, firstName, lastName string) ([]models.Teacher, error) {
		return nil, ErrInvalidTeacherIDs
	}
	defer func() { fetchTeachersForStudent = orig }()

	r := setupRouter(nil)
	req, _ := http.NewRequest("GET", "/student/teachers?firstName=John&lastName=Doe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d got %d", http.StatusInternalServerError, w.Code)
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

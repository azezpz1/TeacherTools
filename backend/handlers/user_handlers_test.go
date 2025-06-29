package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/azezpz1/TeacherTools/models"
)

func setupUserRouter(client *firestore.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/signup", CreateAccountHandler(client))
	r.POST("/login", LoginHandler(client))
	r.POST("/logout", LogoutHandler())
	return r
}

func TestCreateAccountHandler_Success(t *testing.T) {
	orig := createUserInDB
	createUserInDB = func(ctx context.Context, client *firestore.Client, user models.User) error {
		return nil
	}
	defer func() { createUserInDB = orig }()

	r := setupUserRouter(nil)
	body := `{"email":"a@test.com","password":"pass"}`
	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d", http.StatusCreated, w.Code)
	}
}

func TestCreateAccountHandler_Conflict(t *testing.T) {
	orig := createUserInDB
	createUserInDB = func(ctx context.Context, client *firestore.Client, user models.User) error {
		return ErrUserExists
	}
	defer func() { createUserInDB = orig }()

	r := setupUserRouter(nil)
	body := `{"email":"a@test.com","password":"pass"}`
	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected %d got %d", http.StatusConflict, w.Code)
	}
}

func TestCreateAccountHandler_InvalidJSON(t *testing.T) {
	r := setupUserRouter(nil)
	body := `{"email":"a@test.com","password":"pass"`
	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateAccountHandler_MissingValues(t *testing.T) {
	r := setupUserRouter(nil)
	body := `{"email":"","password":"pass"}`
	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	origFetch := fetchUserByEmail
	fetchUserByEmail = func(ctx context.Context, client *firestore.Client, email string) (models.User, error) {
		return models.User{Email: email, PasswordHash: string(hashed)}, nil
	}
	origSign := signJWT
	signJWT = func(email string) (string, error) { return "tok", nil }
	defer func() { fetchUserByEmail = origFetch; signJWT = origSign }()

	r := setupUserRouter(nil)
	body := `{"email":"a@test.com","password":"pass"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestLoginHandler_Invalid(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	orig := fetchUserByEmail
	fetchUserByEmail = func(ctx context.Context, client *firestore.Client, email string) (models.User, error) {
		return models.User{Email: email, PasswordHash: string(hashed)}, nil
	}
	defer func() { fetchUserByEmail = orig }()

	r := setupUserRouter(nil)
	body := `{"email":"a@test.com","password":"wrong"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	orig := fetchUserByEmail
	fetchUserByEmail = func(ctx context.Context, client *firestore.Client, email string) (models.User, error) {
		return models.User{}, ErrUserNotFound
	}
	defer func() { fetchUserByEmail = orig }()

	r := setupUserRouter(nil)
	body := `{"email":"a@test.com","password":"pass"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLoginHandler_MissingValues(t *testing.T) {
	r := setupUserRouter(nil)
	body := `{"email":"","password":"pass"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	r := setupUserRouter(nil)
	body := `{"email":"a@test.com","password":"pass"`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogoutHandler(t *testing.T) {
	r := setupUserRouter(nil)
	req, _ := http.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, w.Code)
	}
}

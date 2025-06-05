package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/azezpz1/TeacherTools/models"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

// functions for database operations for easier testing
var createUserInDB = func(ctx context.Context, client *firestore.Client, user models.User) error {
	_, err := client.Collection("users").Doc(user.Email).Create(ctx, user)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return ErrUserExists
		}
		return err
	}
	return nil
}

var fetchUserByEmail = func(ctx context.Context, client *firestore.Client, email string) (models.User, error) {
	doc, err := client.Collection("users").Doc(email).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	var u models.User
	if err := doc.DataTo(&u); err != nil {
		return models.User{}, err
	}
	return u, nil
}

var signJWT = func(email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev_secret"
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(email))
	mac.Write(b)
	sum := mac.Sum(nil)
	token := append(b, sum...)
	return base64.StdEncoding.EncodeToString(token), nil
}

func CreateAccountHandler(client *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing email or password"})
			return
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := createUserInDB(context.Background(), client, models.User{Email: req.Email, PasswordHash: string(hashed)}); err != nil {
			if err == ErrUserExists {
				c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusCreated)
	}
}

func LoginHandler(client *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing email or password"})
			return
		}
		u, err := fetchUserByEmail(context.Background(), client, req.Email)
		if err != nil {
			if err == ErrUserNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		token, err := signJWT(req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For JWT based auth there is nothing to do server side
		c.Status(http.StatusOK)
	}
}

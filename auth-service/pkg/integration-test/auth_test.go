package integration_test

import (
	"context"
	"fmt"
	"github.com/dreamteam/auth-service/pkg/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"testing"

	"github.com/dreamteam/auth-service/pkg/db"
	"github.com/dreamteam/auth-service/pkg/models"
	"github.com/dreamteam/auth-service/pkg/pb"
	"github.com/dreamteam/auth-service/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRegisterIntegration(t *testing.T) {
	testDB, err := createTestDatabase()
	assert.NoError(t, err)

	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}

	req := &pb.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := server.Register(context.Background(), req)
	assert.NoError(t, err)

	assert.Equal(t, int64(http.StatusConflict), resp.Status)

	var user models.User
	err = testDB.Where(&models.User{Email: req.Email}).First(&user).Error
	assert.NoError(t, err)

	assert.Equal(t, req.Email, user.Email)
}

func TestLoginIntegration(t *testing.T) {
	testDB, err := createTestDatabase()
	assert.NoError(t, err)
	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}
	email := "test@example.com"
	password := "password123"
	hashedPassword := utils.HashPassword(password)

	testDB.Create(&models.User{
		Email:    email,
		Password: hashedPassword,
	})

	req := &pb.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := server.Login(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, int64(http.StatusOK), resp.Status)
	assert.NotEmpty(t, resp.Token)
}

func TestValidateIntegration(t *testing.T) {
	testDB, err := createTestDatabase()
	assert.NoError(t, err)
	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}
	email := "test@example.com"
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	testDB.Create(&models.User{
		Email:    email,
		Password: hashedPassword,
	})
	token, _ := server.Jwt.GenerateToken(models.User{Email: email})
	req := &pb.ValidateRequest{
		Token: token,
	}
	resp, err := server.Validate(context.Background(), req)
	assert.NoError(t, err)
	fmt.Println(resp.Status)
}

func createTestDatabase() (*gorm.DB, error) {
	dsn := "host=horton.db.elephantsql.com user=moyjywkq   password=llyE2HAAR0lkdJPzqKJq7Hk5PeIf_p3t dbname=moyjywkq   port=5432 sslmode=disable TimeZone=Asia/Kolkata"
	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	err = testDB.AutoMigrate(&models.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %v", err)
	}

	return testDB, nil
}

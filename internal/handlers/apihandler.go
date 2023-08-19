package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Safwanseban/2fa-go/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type Apihandler struct {
	DB *gorm.DB
}

func NewApiHandler(server *gin.Engine, db *gorm.DB) *Apihandler {
	api := server.Group("/api")
	handler := &Apihandler{
		DB: db,
	}
	{
		api.GET("/healthchecker", func(ctx *gin.Context) {
			message := "Welcome to Two-Factor Authentication with Golang"
			ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
		})
		api.POST("/auth/register", handler.SignUpUser)
		api.POST("/auth/login", handler.LoginUser)
		api.POST("/auth/generateotp", handler.GenerateOTP)
		api.POST("/auth/verify", handler.VerifyOTP)
	}

	return handler
}
func (ac *Apihandler) SignUpUser(ctx *gin.Context) {
	var payload *models.RegisterUserInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	fmt.Println("ivede")
	newUser := models.User{
		Name:     payload.Name,
		Email:    strings.ToLower(payload.Email),
		Password: payload.Password,
	}

	result := ac.DB.Create(&newUser)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Email already exist, please use another email address"})
		return
	} else if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Registered successfully, please login"})
}

func (ac *Apihandler) LoginUser(ctx *gin.Context) {
	var payload *models.LoginUserInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var user models.User
	result := ac.DB.First(&user, "email = ?", strings.ToLower(payload.Email))
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email or Password"})
		return
	}

	userResponse := gin.H{
		"id":          user.ID.String(),
		"name":        user.Name,
		"email":       user.Email,
		"otp_enabled": user.Otp_enabled,
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "user": userResponse})
}

func (ac *Apihandler) GenerateOTP(ctx *gin.Context) {
	var payload *models.OTPInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "codevoweb.com",
		AccountName: "admin@admin.com",
		SecretSize:  15,
	})

	if err != nil {
		panic(err)
	}

	var user models.User
	result := ac.DB.First(&user, "id = ?", payload.UserId)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email or Password"})
		return
	}

	dataToUpdate := models.User{
		Otp_secret:   key.Secret(),
		Otp_auth_url: key.URL(),
	}

	ac.DB.Model(&user).Updates(dataToUpdate)

	otpResponse := gin.H{
		"base32":      key.Secret(),
		"otpauth_url": key.URL(),
	}
	ctx.JSON(http.StatusOK, otpResponse)
}

func (ac *Apihandler) VerifyOTP(ctx *gin.Context) {
	var payload *models.OTPInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	message := "Token is invalid or user doesn't exist"

	var user models.User
	result := ac.DB.First(&user, "id = ?", payload.UserId)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": message})
		return
	}

	valid := totp.Validate(payload.Token, user.Otp_secret)
	if !valid {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": message})
		return
	}

	dataToUpdate := models.User{
		Otp_enabled:  true,
		Otp_verified: true,
	}

	ac.DB.Model(&user).Updates(dataToUpdate)

	userResponse := gin.H{
		"id":          user.ID.String(),
		"name":        user.Name,
		"email":       user.Email,
		"otp_enabled": user.Otp_enabled,
	}
	ctx.JSON(http.StatusOK, gin.H{"otp_verified": true, "user": userResponse})
}

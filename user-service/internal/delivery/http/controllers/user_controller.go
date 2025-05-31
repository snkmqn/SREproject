package controllers

import (
	"user-service/internal/core/models"
	"user-service/internal/infrastructure/utils/jwt"
	"user-service/internal/usecases/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserController struct {
	service    *services.UserService
	jwtService jwt.JWTService
}

func NewUserController(service *services.UserService, jwtService jwt.JWTService) *UserController {
	return &UserController{
		service:    service,
		jwtService: jwtService,
	}
}

func (uc *UserController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	createdUser, err := uc.service.RegisterUser(c, user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

func (uc *UserController) Login(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := uc.service.AuthenticateUser(c, credentials.Email, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := uc.jwtService.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating JWT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

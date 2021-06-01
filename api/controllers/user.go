package controllers

import (
	"fmt"
	"net/http"
	"travel/api/responses"
	"travel/api/services"
	"travel/constants"
	"travel/infrastructure"
	"travel/models"

	"github.com/gin-gonic/gin"
)

// UserController -> data type
type UserController struct {
	logger          infrastructure.Logger
	userService     services.UserService
	firebaseService services.FirebaseService
}

// NewUserController -> creates new user controller
func NewUserController(logger infrastructure.Logger, userService services.UserService, firebaseService services.FirebaseService) UserController {
	return UserController{
		logger:          logger,
		userService:     userService,
		firebaseService: firebaseService,
	}
}

// CreateUser -> creates the user
func (u UserController) CreateUser(c *gin.Context) {
	requestUser := struct {
		models.User
		Password string `json:"password"`
	}{}
	if err := c.ShouldBindJSON(&requestUser); err != nil {
		u.logger.Zap.Error("Error (ShouldBindJSON) ::", err.Error())
		responses.ErrorJSON(c, http.StatusBadRequest, "Error parsing json request")
		return
	}
	// Checks for firebase existing user
	firebaseID := u.firebaseService.GetUserByEmail(requestUser.Email)
	if firebaseID != "" {
		u.logger.Zap.Error("Error [create firebase user] ::")
		responses.ErrorJSON(c, http.StatusBadRequest, "The provided email is already in use")
		return
	}
	firebaseID, firebaseErr := u.firebaseService.CreateUser(requestUser.Email, requestUser.Password, requestUser.Name, constants.CustomerUserType)
	if firebaseErr != nil {
		u.logger.Zap.Error("Error [create firebase user] ::", firebaseErr)
		responses.ErrorJSON(c, http.StatusBadRequest, "Error creating user in firebase")
		return
	}
	var user models.User
	user.ID = firebaseID
	user.Address = requestUser.Address
	user.Phone = requestUser.Phone
	user.Email = requestUser.Email
	user.Name = requestUser.Name
	user.UserType = constants.CustomerUserType
	_, err := u.userService.CreateUser(user)
	if err != nil {
		u.logger.Zap.Error("Error [db CreateUser]: ", err.Error())
		responses.ErrorJSON(c, http.StatusInternalServerError, "failed to Save User in Database")
		return
	}
	msg, err := u.firebaseService.GenerateEmailVerificationLink(requestUser.Email)
	if err != nil {
		fmt.Println("Error", err)
	}
	u.logger.Zap.Info(msg)
	responses.SuccessJSON(c, http.StatusOK, "User created successfully")

}

// GetUserProfile -> gets a user by ID
func (u UserController) GetUserProfile(c *gin.Context) {
	uid := c.MustGet(constants.UID).(string)
	user, err := u.userService.GetUserByID(uid)
	if err != nil {
		u.logger.Zap.Error("Error getting user ::", err)
		responses.ErrorJSON(c, http.StatusBadRequest, "failed to get user")
		return
	}
	responses.JSON(c, http.StatusOK, user)
}

//UpdateUser
func (u UserController) UpdateUser(c *gin.Context) {

	requestUser := struct {
		models.User
		Password string `json:"password"`
	}{}
	if err := c.ShouldBindJSON(&requestUser); err != nil {
		u.logger.Zap.Error("Error [ShouldBindJSON] ::", err)
		responses.ErrorJSON(c, http.StatusBadRequest, "failed to parse json data")
		return
	}

}

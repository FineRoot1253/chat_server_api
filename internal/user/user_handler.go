package user

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/cmd/chat_server_api/presenter"
	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/JunGeunHong1129/chat_server_api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler interface {
	GetUserListHandler(c *fiber.Ctx) error
	CreateUserHandler(c *fiber.Ctx) error
	UserLoginHandler(c *fiber.Ctx) error
	CheckUserEmailDupHandler(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHander(service Service) Handler {
	return handler{service: service}
}

func (handler handler) GetUserListHandler(c *fiber.Ctx) error {
	key := c.Query("nickname")

	userList, err := handler.service.GetUserList(key)
	if err != nil {
		c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	c.Context().Response.Header.Add("Content-Type", "application/json")

	return c.Status(200).JSON(presenter.Success(userList, "ok"))
}

func (handler handler) CreateUserHandler(c *fiber.Ctx) error {

	tx := c.Locals("TX").(*gorm.DB)

	var user models.User
	var userState models.UserState

	if err := json.Unmarshal(c.Body(), &user); err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	if err := json.Unmarshal(c.Body(), &userState); err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), bcrypt.DefaultCost)

	if err != nil {
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	log.Printf("Hashed PWD GEN complete : %s", hashedPassword)

	user.Pwd = string(hashedPassword)

	user.CreateAt = time.Now()

	if err := handler.service.WithTx(tx).CreateUser(&user); err != nil {
		tx.Rollback()
		return c.Status(200).JSON(presenter.Failure(err.Error()))
	}

	userState.User = user
	userState.User_State = 1
	userState.CreateAt = user.CreateAt

	if err := handler.service.WithTx(tx).CreateUserState(&userState); err != nil {
		tx.Rollback()
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB작업 도중 에러가 발생했습니다."})
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		errLog := &utils.CommonError{Func:"Commit",Data: "" ,Err:err}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	}
	c.Response().Header.Set("Content-Type", "application/json")

	return c.Status(200).JSON(presenter.Success(user, "ok"))
}

func (handler handler) UserLoginHandler(c *fiber.Ctx) error {
	// c.Request().Header.

	tx := c.Locals("TX").(*gorm.DB)

	auth := strings.SplitN(string(c.Request().Header.Peek("Authorization")), " ", 2)

	userFcmToken := c.Query("user_fcm_token")

	if len(auth) != 2 || auth[0] != "Basic" {
		log.Println("Error parsing basic auth")

		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		log.Println("Error parsing basic auth")

		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	var user models.User

	errorr := handler.service.WithTx(tx).GetUserWithError(pair[0], &user)
	// errorr := handler.service.WithTx(tx).GetUser(pair[0],&user)

	log.Println(user)

	if errors.Is(errorr, gorm.ErrRecordNotFound) {
		errLog := &utils.CommonError{Func:"GetUser:UserNotFound",Data: "" ,Err:errorr}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	} else if errorr != nil {
		errLog := &utils.CommonError{Func:"GetUser",Data: "" ,Err:errorr}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(pair[1]))
	if err != nil {
		errLog := &utils.CommonError{Func:"CompareHashAndPassword",Data: "" ,Err:err}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	}

	userState := models.UserState{User: user, User_State: 1, User_FCM_TOKEN: userFcmToken, CreateAt: time.Now()}

	if err := handler.service.WithTx(tx).CreateUserState(&userState); err != nil {
		log.Print(err)
		errLog := &utils.CommonError{Func:"CreateUserState",Data: "" ,Err:err}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
		// return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB작업 도중 에러가 발생했습니다."})
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		errLog := &utils.CommonError{Func:"Commit",Data: "" ,Err:err}
		return c.Status(200).JSON(presenter.Failure(errLog.Error()))
	}

	c.Response().Header.Set("Content-Type", "application/json")

	return c.Status(200).JSON(presenter.Success(user, "ok"))

}

func (handler handler) CheckUserEmailDupHandler(c *fiber.Ctx) error {
	key := c.Query("email_addr")
	log.Printf("EMAIL : %v\n", key)
	var user models.User

	err := handler.service.GetUserWithError(key, &user)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Pass")
		return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok"})
	} else if err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업 도중 에러가 발생했습니다."})
	}

	log.Println("User Dup Occurred")
	return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "이미 존재하는 이메일입니다."})

}
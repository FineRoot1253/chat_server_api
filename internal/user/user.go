package user

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JunGeunHong1129/chat_server_api/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

func GetUserList(c *fiber.Ctx) error {
	key := c.Query("nickname")
	var userList []models.User
	keyword := fmt.Sprint("%", key, "%")


	if err := db.Connector.Raw("select * from (select * from (select * from chat_server_dev.user_state where user_state_id  in (select max(user_state_id) from chat_server_dev.user_state group by user_id) and user_state > 0) as us , chat_server_dev.\"user\" as u where us.user_id=u.user_id) as mainu  where mainu.nickname like ? and mainu.user_fcm_token != '0' AND mainu.user_fcm_token != '';", keyword).Find(&userList).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 조회중 에러가 발생했습니다."})
	}

	c.Context().Response.Header.Add("Content-Type", "application/json")

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: userList})
}

func CreateUser(c *fiber.Ctx) error {

	var user models.User
	var userState models.UserState

	if err := json.Unmarshal(c.Body(), &user); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	if err := json.Unmarshal(c.Body(), &userState); err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "데이터 파싱중 에러가 발생했습니다."})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), bcrypt.DefaultCost)

	if err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "패스워드 암호화 도중 에러가 발생했습니다."})
	}

	log.Printf("Hashed PWD GEN complete : %s", hashedPassword)

	user.Pwd = string(hashedPassword)

	user.CreateAt = time.Now()

	if err := db.Connector.Save(&user).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB작업 도중 에러가 발생했습니다."})
	}

	userState.User=user
	userState.User_State=1
	userState.CreateAt = user.CreateAt

	if err := db.Connector.Create(&userState).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB작업 도중 에러가 발생했습니다."})
	}

	c.Response().Header.Set("Content-Type", "application/json")

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok", Result: user})
}

func UserLogin(c *fiber.Ctx) error {
	// c.Request().Header.
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

	errorr := db.Connector.Raw("select * from chat_server_dev.user u where email_addr = ?;", pair[0]).First(&user).Error

	log.Println(user)

	if errors.Is(errorr, gorm.ErrRecordNotFound) {
		log.Printf("Username provided is correct: %s\n", pair[0])
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "유저를 찾지 못했습니다."})
	} else if errorr != nil {
		log.Print(errorr)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB 작업 도중 에러가 발생했습니다."})
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(pair[1]))
	if err != nil {
		log.Printf("Password provided is correct: %s\n", err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "패스워드가 일치 하지 않습니다."})
	}


	userState := models.UserState{User: user,User_State: 1,User_FCM_TOKEN: userFcmToken,CreateAt: time.Now()};

	if err := db.Connector.Create(&userState).Error; err != nil {
		log.Print(err)
		return c.Status(200).JSON(models.ResultModel{Code: -1, Msg: "DB작업 도중 에러가 발생했습니다."})
	}

	return c.Status(200).JSON(models.ResultModel{Code: 1, Msg: "ok",Result:user})

}

func CheckUserEmailDup(c *fiber.Ctx) error {
	key := c.Query("email_addr")
	log.Printf("EMAIL : %v\n", key)
	var user models.User

	err := db.Connector.Where("email_addr = ?", key).First(&user).Error

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

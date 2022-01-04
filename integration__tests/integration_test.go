package integrestion__tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"

	// "net/http"
	"net/http/httptest"
	"testing"

	"github.com/JunGeunHong1129/chat_server_api/models"
	"github.com/JunGeunHong1129/chat_server_api/routes"
	"github.com/stretchr/testify/assert"
)

/// 공통)
/// app init

var createRoomResult struct {
	Data       models.Room     `json:"room"`
	MemberList []models.Member `json:"member_list"`
}

type DummyModel struct {
	Data     models.Room `json:"room"`
	MemberId int64       `json:"member_id"`
	UserId   int64       `json:"user_id"`
}

var dummyList []DummyModel

var deletedChatLogId string

/// 방 생성 확인
func TestCreateRoom(t *testing.T) {
	initFunc()
	app := routes.InitaliseHandlers()

	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		data         string // post data
		expectedCode int    // expected result model code
	}{
		{
			description:  "success case",
			route:        "/chat/v1/room/create",
			data:         `{"users":[20,21],"room_state":0,"room_name":"test-3"}`,
			expectedCode: 1,
		},
		{
			description:  "[failure case] Body를 비운채 채팅방 생성 요청을 합니다.",
			route:        "/chat/v1/room/create",
			data:         "",
			expectedCode: -1,
		},
	}
	isDummyInstanced := false
	for _, v := range tests {
		var result models.ResultModel
		req := httptest.NewRequest("POST", v.route, bytes.NewBuffer([]byte(v.data)))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req, -1)

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			panic(err)
		}
		resp.Body.Close()
		log.Print("결과 : " + strconv.Itoa(v.expectedCode) + " ::: " + strconv.Itoa(result.Code))

		/// Result가 빈터페이스라 지 알아서 map[string]interface{} 됨
		/// 다시 내 커스텀 구조체에 맞춰주기 위해 이렇게 돌림

		if isDummyInstanced == false {
			resultByte, err := json.Marshal(result.Result)
			if err != nil {
				panic(err)

			}
			if err := json.Unmarshal(resultByte, &createRoomResult); err != nil {
				panic(err)
			}

			for _, v := range createRoomResult.MemberList {
				dummyList = append(dummyList, DummyModel{
					createRoomResult.Data, v.Member_Id, v.User_Id})
			}
			isDummyInstanced = true
		}

		assert.Equal(t, v.expectedCode, result.Code, v.description)

	}
	log.Print("TestCreateRoom DONE")
}

/// redis 확인
func TestPublishMsgCheckOneTime(t *testing.T) {
	// initFunc()
	app := routes.InitaliseHandlers()
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		data         string // post data
		expectedCode int    // expected result model code
	}{
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[0], 0),
			expectedCode: 1,
		},
		{
			description:  "[failure case] Body를 비운채 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         "",
			expectedCode: -1,
		},
	}

	for _, v := range tests {
		var result models.ResultModel
		req := httptest.NewRequest("POST", v.route, bytes.NewBuffer([]byte(v.data)))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req, -1)

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			panic(err)
		}
		resp.Body.Close()
		log.Print("결과 : " + strconv.Itoa(v.expectedCode) + " ::: " + strconv.Itoa(result.Code))
		if result.Code == 1 {
			deletedChatLogId = result.Result.(string)
			log.Print("deletedChatLogId 결과 : ", deletedChatLogId)

		}
		assert.Equal(t, v.expectedCode, result.Code, v.description)
	}
	log.Print("TestPublishMsgCheckOneTime DONE")
}

/// postgresql 저장 확인
func TestPublishMsgCheck5Times(t *testing.T) {
	// initFunc()
	app := routes.InitaliseHandlers()
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		data         string // post data
		expectedCode int    // expected result model code
	}{
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[0], 0),
			expectedCode: 1,
		},
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[0], 0),
			expectedCode: 1,
		},
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[0], 0),
			expectedCode: 1,
		},
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[0], 0),
			expectedCode: 1,
		},
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[0], 0),
			expectedCode: 1,
		},
		{
			description:  "[failure case] Body를 비운채 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         "",
			expectedCode: -1,
		},
	}

	for _, v := range tests {
		var result models.ResultModel
		req := httptest.NewRequest("POST", v.route, bytes.NewBuffer([]byte(v.data)))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req, -1)

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			panic(err)
		}
		resp.Body.Close()
		log.Print("결과 : " + strconv.Itoa(v.expectedCode) + " ::: " + strconv.Itoa(result.Code))
		assert.Equal(t, v.expectedCode, result.Code, v.description)
	}
	log.Print("TestPublishMsgCheck5Times DONE")

}

/// 삭제 후 redis 확인 [id_state, chat_id, chat_id 리스트까지 총 3번]
func TestPublishRemoveMsgOneTimeUser0(t *testing.T) {
	// initFunc()
	app := routes.InitaliseHandlers()
	log.Print("삭제요청 시작!")
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		data         string // post data
		expectedCode int    // expected result model code
	}{
		{
			description:  "[success case] 메시지 삭제 처리 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubDataForDelete(dummyList[0], 2),
			expectedCode: 1,
		},
		{
			description:  "[failure case] Body를 비운채 메시지 삭제 처리 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         "",
			expectedCode: -1,
		},
	}

	for _, v := range tests {
		var result models.ResultModel

		req := httptest.NewRequest("POST", v.route, bytes.NewBuffer([]byte(v.data)))
		req.Header.Set("Content-Type", "application/json")
		// Perform the request plain with the app,
		// the second argument is a request latency
		// (set to -1 for no latency)
		resp, _ := app.Test(req, -1)

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			panic(err)
		}
		resp.Body.Close()
		log.Print("결과 : " + strconv.Itoa(v.expectedCode) + " ::: " + strconv.Itoa(result.Code))

		// Verify, if the status code is as expected
		assert.Equalf(t, v.expectedCode, result.Code, v.description)
	}
	log.Print("TestPublishRemoveMsgOneTimeUser0 DONE")

}

func TestUpdateMemberReadMsgIndex(t *testing.T) {
	// initFunc()
	app := routes.InitaliseHandlers()
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		data         string // post data
		expectedCode int    // expected result model code
	}{
		{
			description:  "[success case] 채팅방 나가기 로깅 요청을 합니다.",
			route:        "/chat/v1/room/updateLastReadMsgIdx",
			data:         `{"member_id":` + strconv.Itoa(int(dummyList[0].MemberId)) + `,"member_state":1}`,
			expectedCode: 1,
		},
		{
			description:  "[failure case] Body를 비운채 채팅방 나가기 로깅 요청을 합니다.",
			route:        "/chat/v1/room/updateLastReadMsgIdx",
			data:         "",
			expectedCode: -1,
		},
	}

	for _, v := range tests {
		var result models.ResultModel

		req := httptest.NewRequest("POST", v.route, bytes.NewBuffer([]byte(v.data)))
		req.Header.Set("Content-Type", "application/json")
		// Perform the request plain with the app,
		// the second argument is a request latency
		// (set to -1 for no latency)
		resp, _ := app.Test(req, -1)

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		log.Print("결과asdf : ", respBody)
		if err := json.Unmarshal(respBody, &result); err != nil {
			panic(err)
		}
		resp.Body.Close()
		log.Print("결과 : " + strconv.Itoa(v.expectedCode) + " ::: " + strconv.Itoa(result.Code))

		// Verify, if the status code is as expected
		assert.Equalf(t, v.expectedCode, result.Code, v.description)
	}
	log.Print("TestUpdateMemberReadMsgIndex DONE")

}

/// postgresql 저장 확인
func TestPublishMsgCheck5TimesUser1(t *testing.T) {
	// initFunc()
	app := routes.InitaliseHandlers()
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		data         string // post data
		expectedCode int    // expected result model code
	}{
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[1], 0),
			expectedCode: 1,
		},
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[1], 0),
			expectedCode: 1,
		},
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[1], 0),
			expectedCode: 1,
		},
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[1], 0),
			expectedCode: 1,
		},
		{
			description:  "[success case] 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         pubData(dummyList[1], 0),
			expectedCode: 1,
		},
		{
			description:  "[failure case] Body를 비운채 메시지 생성 요청을 합니다.",
			route:        "/chat/v1/log/chatSomeThing",
			data:         "",
			expectedCode: -1,
		},
	}

	for _, v := range tests {
		var result models.ResultModel
		req := httptest.NewRequest("POST", v.route, bytes.NewBuffer([]byte(v.data)))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req, -1)

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		log.Print(respBody)
		if err := json.Unmarshal(respBody, &result); err != nil {
			panic(err)
		}
		resp.Body.Close()
		log.Print("결과 : " + strconv.Itoa(v.expectedCode) + " ::: " + strconv.Itoa(result.Code))
		assert.Equal(t, v.expectedCode, result.Code, v.description)
	}
}
func TestGetRestOfMsgList(t *testing.T) {
	// initFunc()
	app := routes.InitaliseHandlers()
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		data         string // post data
		expectedCode int    // expected result model code
	}{
		{
			description:  "[success case] 채팅방 채팅이력 조회 요청을 합니다.",
			route:        "/chat/v1/log/restOfMsg",
			data:         `{"member_id":` + strconv.Itoa(int(dummyList[0].MemberId)) + `,"room_id":` + strconv.Itoa(int(dummyList[0].Data.Room_Id)) + `}`,
			expectedCode: 1,
		},
		{
			description:  "[failure case] Body를 비운채 채팅방 채팅이력 조회 요청을 합니다.",
			route:        "/chat/v1/log/restOfMsg",
			data:         "",
			expectedCode: -1,
		},
	}

	for _, v := range tests {
		var result models.ResultModel

		req := httptest.NewRequest("POST", v.route, bytes.NewBuffer([]byte(v.data)))
		req.Header.Set("Content-Type", "application/json")

		// Perform the request plain with the app,
		// the second argument is a request latency
		// (set to -1 for no latency)
		resp, _ := app.Test(req, -1)

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		log.Print(respBody)
		if err := json.Unmarshal(respBody, &result); err != nil {
			panic(err)
		}
		resp.Body.Close()
		log.Print("결과 : " + strconv.Itoa(v.expectedCode) + " ::: " + strconv.Itoa(result.Code))

		// Verify, if the status code is as expected
		assert.Equalf(t, v.expectedCode, result.Code, v.description)
	}
}

func pubData(dummy DummyModel, chatState int) string {
	return `{"chat_content":"6","room_id":` + strconv.Itoa(int(dummy.Data.Room_Id)) + `,"user_id":` + strconv.Itoa(int(dummy.UserId)) + `,"member_id":` + strconv.Itoa(int(dummy.MemberId)) + `,"chat_state":` + strconv.Itoa(int(dummy.MemberId)) + `}`
}
func pubDataForDelete(dummy DummyModel, chatState int) string {
	log.Print("삭제 요청 전 데이터 확인 : ", `{"chat_content":"6","room_id":`+strconv.Itoa(int(dummy.Data.Room_Id))+`,"chat_id":`+deletedChatLogId+`,"user_id":`+strconv.Itoa(int(dummy.UserId))+`,"member_id":`+strconv.Itoa(int(dummy.MemberId))+`,"chat_state":`+strconv.Itoa(int(dummy.MemberId))+`}`)
	return `{"chat_content":"6","room_id":` + strconv.Itoa(int(dummy.Data.Room_Id)) + `,"chat_id":"` + deletedChatLogId + `","user_id":` + strconv.Itoa(int(dummy.UserId)) + `,"member_id":` + strconv.Itoa(int(dummy.MemberId)) + `,"chat_state":` + strconv.Itoa(int(dummy.MemberId)) + `}`

}

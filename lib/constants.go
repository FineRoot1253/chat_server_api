package lib

type ChatStateType int64
type ErrorStateType int64

const (
	Normal_Msg ChatStateType = iota ///  일반 메시지
	Remove_Only_Me_Msg	/// 나에게서만 안보이게 하기 [미구현]
	Remove_To_All_Msg /// 남들한테 않보이게 하기
	User_Room_Exit_Msg /// 방 나가기
	User_Room_Add_Msg /// 새 유저 추가됨
	Image_Msg	/// 이미지 [미구현] (누가 S3 사줘잉)
	Imoticon_Msg /// 이모티콘 [미구현]
)

const (
	None ErrorStateType = iota
	Unexpected_Error
	Unmarshaling_Error
	Marshal_Error
	Redis_Error
	Rdb_Error
)

const (
	HOST          = "postgresql"
	POSTGRES_PWD  = "reindeer2021"
	POSTGRES_USER = "postgres"
	POSTGRES_DB   = "postgres"
	POSTGRES_PORT = 26000
	///TODO : 25000으로 수정 필요
	REDIS_PORT                 = 25000
	CHAT_SERVER_PORT           = 50000
	HAPOXY_PORT                = 1936
	RABBITMQ_DEFAULT_PASS      = "reindeer2017!"
	RABBITMQ_DEFAULT_USER      = "g9bon"
	RABBITMQ_DEFAULT_VHOST     = "chat"
	PER_SAVE_AMOUNT            = 5
	FIREBASE_ACCOUNT_FILE_PATH = "../chat-88196-firebase-adminsdk-5es5m-ccf8c2cbfd.json"
)

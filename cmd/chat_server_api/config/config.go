package config

const (
	/// for prod
	// HOST          = "postgresql"
	HOST          = "localhost"
	POSTGRES_PWD  = "reindeer2021"
	POSTGRES_USER = "postgres"
	POSTGRES_DB   = "postgres"
	POSTGRES_PORT = 26000
	///TODO : 25000으로 수정 필요
	REDIS_PORT                 = 25000
	CHAT_SERVER_PORT           = 50000
	HAPOXY_PORT                = 1936
	PER_SAVE_AMOUNT            = 5
	FIREBASE_ACCOUNT_FILE_PATH = "../chat-88196-firebase-adminsdk-5es5m-ccf8c2cbfd.json"
)

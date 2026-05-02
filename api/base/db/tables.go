package db

const (
	TABLE_USERS = `
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		auth_type INTEGER NOT NULL,
		password_hash TEXT NOT NULL,
		email_verified BOOLEAN DEFAULT FALSE,
		active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	`

	TABLE_SESSIONS = `
		id VARCHAR(255) PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		csrf_token VARCHAR(255) NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		started_at TIMESTAMP WITH TIME ZONE NOT NULL
	`

	TABLE_SESSIONS_GUEST = `
		id VARCHAR(255) PRIMARY KEY,
		client_ip VARCHAR(255) NOT NULL,
		csrf_token VARCHAR(255) NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL
	`

	TABLE_VERIFICATION_TOKEN = `
		id VARCHAR(255) PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token VARCHAR(255) NOT NULL,
		usage_id INTEGER NOT NULL,
		resend_count INTEGER DEFAULT 0,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,

		UNIQUE (user_id, usage_id)
	`

	TABLE_USER_LOGONS = `
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		auth_type INTEGER NOT NULL,
		device_type INTEGER NOT NULL,
		device_info VARCHAR(255),
		time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	`
)

var TABLES = []map[string]string{
	{"users": TABLE_USERS},
	{"sessions": TABLE_SESSIONS},
	{"sessions_guest": TABLE_SESSIONS_GUEST},
	{"verification_token": TABLE_VERIFICATION_TOKEN},
	{"user_logons": TABLE_USER_LOGONS},
}

var INDEXES = map[string]string{
	"sessions_guest - client_ip": "idx_sessions_guest_client_ip ON sessions_guest (client_ip);",
}

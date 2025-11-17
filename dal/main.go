package dal

type ChatRoom struct {
	Id          int64  `json:"id"`
	UserId      string `json:"user_id"`
	UniqueId    string `json:"unique_id"`
	DisplayName string `json:"display_name"`
	ChatRoomId  string `json:"chat_room_id"`
	AvatarUrl   string `json:"avatar_url"`
}

func (c ChatRoom) TableName() string {
	return "chat_rooms"
}

func createTable() error {
	var err error
	err = db.Exec(`CREATE TABLE IF NOT EXISTS chat_rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    unique_id TEXT NOT NULL,
    display_name TEXT NOT NULL,
    chat_room_id TEXT NOT NULL,
    avatar_url TEXT NOT NULL
)`).Error
	if err != nil {
		return err
	}
	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_chat_rooms_display_name ON chat_rooms(display_name);`).Error
	if err != nil {
		return err
	}
	err = db.Exec(`CREATE TABLE IF NOT EXISTS live_streams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    display_name TEXT NOT NULL,
    title TEXT NOT NULL,
    webrtc_url TEXT NOT NULL,
    rtmp_url TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    end_time INTEGER NOT NULL
)`).Error
	if err != nil {
		return err
	}
	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_live_streams_display_name ON live_streams(display_name);`).Error
	if err != nil {
		return err
	}
	err = db.Exec(`CREATE TABLE IF NOT EXISTS chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    display_name TEXT NOT NULL,
    chat_room_id TEXT NOT NULL,
    chat_message_id TEXT NOT NULL,
    msg_type INTEGER NOT NULL,
    content TEXT DEFAULT '',
    image_url TEXT DEFAULT '',
    video_url TEXT DEFAULT '',
    video_path TEXT DEFAULT '',
    image_path TEXT DEFAULT '',
    send_time INTEGER NOT NULL,
    Time_str TEXT NOT NULL DEFAULT ''
)`).Error
	if err != nil {
		return err
	}
	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id ON chat_messages(user_id);`).Error
	if err != nil {
		return err
	}
	err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_message_id ON chat_messages(chat_message_id);`).Error
	if err != nil {
		return err
	}
	return nil
}

type ChatMessage struct {
	Id            int64  `json:"id"`
	UserId        string `json:"user_id"`
	DisplayName   string `json:"display_name"`
	ChatRoomId    string `json:"chat_room_id"`
	ChatMessageId string `json:"chat_message_id"`
	MsgType       int32  `json:"msg_type"`
	Content       string `json:"content"`
	ImageUrl      string `json:"image_url"`
	VideoUrl      string `json:"video_url"`
	VideoPath     string `json:"video_path"`
	ImagePath     string `json:"image_path"`
	TimeStr       string `json:"time_str"`
	SendTime      int64  `json:"send_time"`
}

func (c ChatMessage) TableName() string {
	return "chat_messages"
}

type LiveStream struct {
	Id          int64  `json:"id"`
	UserId      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Title       string `json:"title"`
	WebrtcUrl   string `json:"webrtc_url"`
	RtmpUrl     string `json:"rtmp_url"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
}

func (c LiveStream) TableName() string {
	return "live_streams"
}

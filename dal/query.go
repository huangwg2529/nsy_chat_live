package dal

func GetChatRooms() ([]*ChatRoom, error) {
	innerChatRooms := make([]*ChatRoom, 0)
	err := db.Table(ChatRoom{}.TableName()).
		Find(&innerChatRooms).Error
	return innerChatRooms, err
}

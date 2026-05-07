package dal

func GetChatRooms() ([]*ChatRoom, error) {
	innerChatRooms := make([]*ChatRoom, 0)
	err := db.Table(ChatRoom{}.TableName()).
		Order("display_name asc, id asc").
		Find(&innerChatRooms).Error
	return innerChatRooms, err
}

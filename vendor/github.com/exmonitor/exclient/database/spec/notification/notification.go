package notification

type UserNotificationSettings struct {
	ID             int
	Target         string
	Type           string
	ResentAfterMin int
}

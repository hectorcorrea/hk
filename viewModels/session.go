package viewModels

// We make everything public here because it's a view model
// (unlike web.session in which everything is private)
type Session struct {
	Id        string
	LoginName string
	IsAuth    bool
	IsAdmin   bool
	IsGuest   bool
}

func NewSession(id, loginName string, isAdmin bool, isGuest bool) Session {
	return Session{
		Id:        id,
		LoginName: loginName,
		IsAdmin:   isAdmin,
		IsGuest:   isGuest,
		IsAuth:    isGuest || isAdmin,
	}
}

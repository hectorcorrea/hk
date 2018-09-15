package viewModels

type Login struct {
	Message   string
	TargetUrl string
	Session
}

func NewLogin(message string, url string, session Session) Login {
	login := Login{Message: message, Session: session}
	if url == "" {
		login.TargetUrl = "/"
	} else {
		login.TargetUrl = url
	}
	return login
}

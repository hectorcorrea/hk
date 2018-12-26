package web

import (
	"errors"
	"log"
	"net/http"
	"time"

	"hk/models"
	"hk/viewModels"
)

type session struct {
	resp      http.ResponseWriter
	req       *http.Request
	cookie    *http.Cookie
	loginName string
	sessionId string
	userType  string
}

func newSession(resp http.ResponseWriter, req *http.Request) session {
	cookie, err := req.Cookie("sessionId")
	if err == nil {
		// Known user with an existing session.
		sessionId := cookie.Value
		userSession, err := models.GetUserSession(sessionId)
		if err != nil {
			log.Printf("Session was not valid (%s), %s", cookie.Value, err)
			return session{resp: resp, req: req}
		}

		err = models.TouchUserSession(sessionId)
		if err != nil {
			log.Printf("Could not update session %s/%s, %s", sessionId, userSession.Login, err)
		}
		return session{
			resp:      resp,
			req:       req,
			cookie:    cookie,
			loginName: userSession.Login,
			sessionId: cookie.Value,
			userType:  userSession.UserType,
		}
	}

	ticketID := req.URL.Query().Get("ticketId")
	if ticketID != "" {
		// Unknown user, but link has a ticket ID
		s := session{
			resp:     resp,
			req:      req,
			userType: "guest",
		}
		err = s.loginTicket(ticketID)
		if err == nil {
			return s
		}
		log.Printf("Ticket was not valid (%s) %s", ticketID, err)
	}

	return session{resp: resp, req: req}
}

func (s *session) logout() {
	models.DeleteUserSession(s.sessionId)
	s.loginName = ""
	s.sessionId = ""
	if s.cookie != nil {
		s.cookie.Value = ""
		s.cookie.Expires = time.Unix(0, 0)
		s.cookie.Path = "/"
		s.cookie.HttpOnly = true
		http.SetCookie(s.resp, s.cookie)
	}
}

func (s *session) login(loginName, password string) error {
	if s.cookie == nil {
		s.cookie = &http.Cookie{Name: "sessionId"}
	}

	logged, err := models.LoginUser(loginName, password)
	if err != nil {
		return err
	}

	if logged {
		userSession, err := models.NewUserSession(loginName)
		if err != nil {
			log.Printf("ERROR creating new session: %s", err)
			return err
		}

		s.loginName = userSession.Login
		s.sessionId = userSession.SessionId
		s.cookie.Value = s.sessionId
		s.cookie.Expires = userSession.ExpiresOn
		s.cookie.Path = "/"
		s.cookie.HttpOnly = true
		http.SetCookie(s.resp, s.cookie)
		return nil
	}

	log.Printf("ERROR invalid user/password received: %s/***", loginName)
	return errors.New("Invalid user/password received")
}

func (s *session) loginTicket(ticketID string) error {
	if s.cookie == nil {
		s.cookie = &http.Cookie{Name: "sessionId"}
	}

	// Tickets are password-less users with a very short lifespan.
	logged, err := models.LoginTicket(ticketID)
	if err != nil {
		return err
	}

	if logged {
		userSession, err := models.NewUserSession(ticketID)
		if err != nil {
			log.Printf("ERROR creating new ticket session: %s", err)
			return err
		}

		s.loginName = userSession.Login
		s.sessionId = userSession.SessionId
		s.cookie.Value = s.sessionId
		s.cookie.Expires = userSession.ExpiresOn
		s.cookie.Path = "/"
		s.cookie.HttpOnly = true
		http.SetCookie(s.resp, s.cookie)
		return nil
	}

	log.Printf("ERROR invalid ticket ID received: %s", ticketID)
	return errors.New("Invalid ticket ID received")
}

func (s session) isAuth() bool {
	return s.loginName != ""
}

func (s session) isAdmin() bool {
	return s.userType == "admin"
}

func (s session) isGuest() bool {
	return s.userType == "guest"
}

// Provide toViewModel() here since this type does not have
// a model per-se.
func (s session) toViewModel() viewModels.Session {
	return viewModels.NewSession(
		s.sessionId, s.loginName,
		s.isAdmin(), s.isGuest())
}

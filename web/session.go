package web

import (
	"errors"
	"log"
	"net/http"
	"time"

	"hectorcorrea.com/hk/models"
	"hectorcorrea.com/hk/viewModels"
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
		sessionID := cookie.Value
		userSession, err := models.GetUserSession(sessionID)
		if err == nil {
			// Known user with a valid sessionID.
			return session{
				resp:      resp,
				req:       req,
				cookie:    cookie,
				loginName: userSession.Login,
				sessionId: cookie.Value,
				userType:  userSession.UserType,
			}
		}
		log.Printf("SessionId was not valid (%s) %s", sessionID, err)
	}

	cookie, err = req.Cookie("ticketId")
	if err == nil {
		ticketID := cookie.Value
		userSession, err := models.GetUserSession(ticketID)
		if err == nil {
			// Known user with a valid ticketID.
			return session{
				resp:      resp,
				req:       req,
				cookie:    cookie,
				loginName: userSession.Login,
				sessionId: cookie.Value,
				userType:  userSession.UserType,
			}
		}
		log.Printf("TicketId was not valid (%s) %s", ticketID, err)
	}

	ticket := req.URL.Query().Get("ticket")
	if ticket != "" {
		s := session{resp: resp, req: req, userType: "guest"}
		err := s.loginTicket(ticket)
		if err == nil {
			// User has a valid ticket in the URL.
			return s
		}
		log.Printf("Ticket was not valid (%s) %s", ticket, err)
	}

	// Anonymous user
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
		s.cookie = &http.Cookie{Name: "sessionId"}
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
	// Tickets are password-less users with a very short lifespan.
	logged, err := models.LoginTicket(ticketID)
	if err != nil {
		return err
	}

	if logged {
		userSession, err := models.NewTicketSession(ticketID)
		if err != nil {
			log.Printf("ERROR creating new ticket session: %s", err)
			return err
		}

		s.loginName = userSession.Login
		s.sessionId = userSession.SessionId
		s.cookie = &http.Cookie{Name: "ticketId"}
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

package services

import (
	"errors"
	"net/http"
	"time"

	"github.com/sunkit02/filete/logging"
	"github.com/sunkit02/filete/utils"
	"github.com/sunkit02/filete/web/types"
)

type AuthServiceConfig struct {
	SessionKey    string
	SessionLength time.Duration
}

type UserSession struct {
	Id      string
	Expires time.Time
}

const DEFAULT_SESSION_LENGTH = 1 * time.Hour

var sessionKey string
var sessionLength = DEFAULT_SESSION_LENGTH

var sessions map[string]UserSession

func InitAuthService(c AuthServiceConfig) {
	sessionKey = c.SessionKey
	if c.SessionLength != 0 {
		sessionLength = c.SessionLength
	}

	sessions = make(map[string]UserSession)
}

// Checks whether the given session key is valid and return an sessionId
// it is and returns an error if not.
func AuthenticateWithSessionKey(key string) (*UserSession, error) {
	if key != sessionKey {
		return nil, errors.New("Invalid session key")
	}

	session := createSession(sessionLength)

	return session, nil
}

// Returns true if the cookie is valid and false if not
func ValidateSessionCookie(cookie http.Cookie) bool {
	if cookie.Name != types.SessionIdCookieName {
		logging.Trace.Println("Invalid cookie name")
		return false
	}

	session, ok := sessions[cookie.Value]
	if !ok {
		return false
	}

	if session.Expires.UnixMilli() < time.Now().UnixMilli() {
		delete(sessions, cookie.Value)
		return false
	}
	return true
}

func createSession(livesFor time.Duration) *UserSession {
	sessionId := generateSessionId()
	// Regenerate cookie if there is a clash
	for {
		if _, ok := sessions[sessionId]; ok {
			sessionId = generateSessionId()
		} else {
			break
		}
	}

	session := UserSession{
		Id:      sessionId,
		Expires: time.Now().Add(livesFor),
	}

	sessions[sessionId] = session
	return &session
}

// Invalidates session. If sessionId doesn't exist then it is a no-op
func InvalidateSession(sessionId string) {
	delete(sessions, sessionId)
}

func generateSessionId() string {
	return utils.GenerateRandomString(40)
}

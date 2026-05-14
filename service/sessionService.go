package forum

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	model "forum/model"
	"net/http"
	"time"
)

type SessionService struct {
	DB *sql.DB
}

func NewSessionService(db *sql.DB) *SessionService {
	return &SessionService{
		DB: db,
	}
}

func (s *SessionService) CreateSession(user model.User) (string, error) {

	token, err := GenerateToken(32)
	if err != nil {
		return "", err
	}

	query := `INSERT INTO Sessions(user_id, session_token, expires_at, nb_connections) VALUES (?, ?, ?, ?)`

	_, err = s.DB.Exec(query, user.Id, token, time.Now().Add(60*time.Minute), 1)

	return token, err
}

func (s *SessionService) GetUserFromSession(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return 0, err
	}

	var userID int

	//fmt.Println("session-userid : ", userID)

	err = s.DB.QueryRow(`
        SELECT user_id
        FROM Sessions
        WHERE session_token = ?
        AND expires_at > datetime('now')
    `, cookie.Value).Scan(&userID)

	//fmt.Println("session-err : ", err)
	return userID, err
}

func (s *SessionService) Logout(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session")
	if err == nil {
		s.DB.Exec("DELETE FROM Sessions WHERE session_token = ?", cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (s *SessionService) CleanExpiredSessions() {
	_, err := s.DB.Exec(`
        DELETE FROM Sessions 
        WHERE expires_at < datetime('now')
    `)
	if err != nil {
		fmt.Println("Erreur nettoyage sessions :", err)
	}
}

func (s *SessionService) HasActiveSession(userID int) (bool, error) {
	var exists int

	err := s.DB.QueryRow(`
        SELECT 1 
        FROM Sessions 
        WHERE user_id = ? 
        AND expires_at > datetime('now')
        LIMIT 1
    `, userID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SessionService) StartCleanup(interval time.Duration) {
	go func() {
		for {
			s.CleanExpiredSessions()
			time.Sleep(interval)
		}
	}()
}

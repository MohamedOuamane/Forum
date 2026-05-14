package forum

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	model "forum/model"
	service "forum/service"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleHandler struct {
	UserService    *service.UserService
	SessionService *service.SessionService
}

func getGoogleConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (h *GoogleHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	config := getGoogleConfig()

	state := generateState()

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	url := config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *GoogleHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	config := getGoogleConfig()
	// 1. Vérifier state
	cookie, err := r.Cookie("oauthstate")
	if err != nil || r.FormValue("state") != cookie.Value {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// supprimer state
	http.SetCookie(w, &http.Cookie{
		Name:   "oauthstate",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// 2. Code OAuth
	code := r.FormValue("code")

	// 3. Exchange token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token error", 500)
		return
	}

	// 4. Client Google
	client := config.Client(context.Background(), token)

	// 5. API Google
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Google API error", 500)
		return
	}
	defer resp.Body.Close()

	// 6. JSON decode
	var gUser model.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil {
		http.Error(w, "JSON error", 500)
		return
	}

	// fallback avatar
	gUser.Picture = "avatars/Generic-Profile.jpg"

	// 7. DB logique
	user, err := h.UserService.HandleGoogleOAuthUser(
		gUser.ID,
		gUser.Email,
		gUser.Name,
		gUser.Picture,
	)
	if err != nil {
		http.Error(w, "DB error"+err.Error(), 500)
		return
	}

	tokenCookie, err := h.SessionService.CreateSession(*user)
	if err != nil {
		fmt.Println("CreateSession:", err)
		http.Redirect(w, r, r.URL.Path+"?show=login&error=1", http.StatusSeeOther)
	}
	// 8. session
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    tokenCookie,
		HttpOnly: true,
		Expires:  time.Now().Add(60 * time.Minute),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	// 9. redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

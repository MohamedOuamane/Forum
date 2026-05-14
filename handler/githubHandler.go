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
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GithubHandler struct {
	UserService    *service.UserService
	SessionService *service.SessionService
}

func getGithubConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/github/callback",
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
	}
}

func generateGithubState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (h *GithubHandler) HandleGithubLogin(w http.ResponseWriter, r *http.Request) {

	config := getGithubConfig()

	state := generateGithubState()

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthGithubState",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	url := config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *GithubHandler) HandleGithubCallback(w http.ResponseWriter, r *http.Request) {
	config := getGithubConfig()
	// Verifier state
	cookie, err := r.Cookie("oauthGithubState")
	if err != nil || r.FormValue("state") != cookie.Value {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// supprimer state
	http.SetCookie(w, &http.Cookie{
		Name:   "oauthGithubState",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	code := r.URL.Query().Get("code")

	token, err := config.Exchange(
		context.Background(),
		code,
	)

	if err != nil {
		http.Error(w, "Erreur OAuth", http.StatusInternalServerError)
		return
	}

	client := config.Client(context.Background(), token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Erreur API GitHub", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var gUser model.GithubUser

	err = json.NewDecoder(resp.Body).Decode(&gUser)
	if err != nil {
		http.Error(w, "Erreur decode JSON", http.StatusInternalServerError)
		return
	}
	name := gUser.Name
	if name == "" {
		name = gUser.Login
	}
	//Get email
	if gUser.Email == "" {
		client := config.Client(context.Background(), token)
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			http.Error(w, "Error get email", http.StatusInternalServerError)
			return
		}
		defer emailResp.Body.Close()

		var emails []model.GithubEmail

		err = json.NewDecoder(emailResp.Body).Decode(&emails)
		if err != nil {
			http.Error(w, "Erreur decode emails", http.StatusInternalServerError)
			return
		}

		for _, e := range emails {
			if e.Primary && e.Verified {
				gUser.Email = e.Email
				break
			}
		}
	}
	// Avatar
	gUser.Picture = "avatars/Generic-Profile.jpg"
	// DB logique
	user, err := h.UserService.HandleGithubOAuthUser(
		strconv.Itoa(gUser.ID),
		gUser.Email,
		name,
		gUser.Picture,
	)
	if err != nil {
		http.Error(w, "DB error "+err.Error(), 500)
		return
	}

	tokenCookie, err := h.SessionService.CreateSession(*user)
	if err != nil {
		fmt.Println("CreateSession:", err)
		http.Redirect(w, r, r.URL.Path+"?show=login&error=1", http.StatusSeeOther)
	}
	// session
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    tokenCookie,
		HttpOnly: true,
		Expires:  time.Now().Add(60 * time.Minute),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	// redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

package forum

import (
	"fmt"
	model "forum/model"
	service "forum/service"
	"html/template"
	"log"
	"net/http"
	"time"
)

type BaseHandler struct {
	UserService     *service.UserService
	SessionService  *service.SessionService
	TemplateCache   map[string]*template.Template
	PostService     *service.PostService
	CategoryService *service.CategoryService
	CommentService  *service.CommentService
	LikeService     *service.LikeService
	ImgService      *service.ImgService
	Template        *template.Template
}

func LoadTemplates() map[string]*template.Template {
	return map[string]*template.Template{
		"home": template.Must(template.ParseFiles(
			"templates/index.html",
			"STATIC/partials/header.html",
			"STATIC/partials/popupLogin.html",
			"STATIC/partials/popupRegister.html",
			"STATIC/partials/popupPost.html",
			"STATIC/partials/popupDelete.html",
			"STATIC/partials/post.html",
		)),
		"post": template.Must(template.ParseFiles(
			"templates/post.html",
			"STATIC/partials/header.html",
			"STATIC/partials/popupLogin.html",
			"STATIC/partials/popupRegister.html",
			"STATIC/partials/popupPost.html",
			"STATIC/partials/popupEditComment.html",
			"STATIC/partials/popupDelete.html",
			"STATIC/partials/comment.html",
			"STATIC/partials/post.html",
		)),
		"profil": template.Must(template.ParseFiles(
			"templates/profil.html",
			"STATIC/partials/header.html",
			"STATIC/partials/popupLogin.html",
			"STATIC/partials/popupRegister.html",
			"STATIC/partials/popupPost.html",
			"STATIC/partials/popupChangePassword.html",
			"STATIC/partials/popupChangeAvatar.html",
			"STATIC/partials/popupChangeName.html",
		)),
		"category": template.Must(template.ParseFiles(
			"templates/category.html",
			"STATIC/partials/header.html",
			"STATIC/partials/popupLogin.html",
			"STATIC/partials/popupRegister.html",
			"STATIC/partials/popupPost.html",
			"STATIC/partials/popupDelete.html",
			"STATIC/partials/post.html",
		)),
		"page404": template.Must(template.ParseFiles(
			"templates/page404.html",
			"STATIC/partials/header.html",
			"STATIC/partials/popupLogin.html",
			"STATIC/partials/popupRegister.html",
		)),
		"page400": template.Must(template.ParseFiles(
			"templates/page400.html",
			"STATIC/partials/header.html",
			"STATIC/partials/popupLogin.html",
			"STATIC/partials/popupRegister.html",
		)),
		"page500": template.Must(template.ParseFiles(
		"templates/page500.html",
		"STATIC/partials/header.html",
		"STATIC/partials/popupLogin.html",
		"STATIC/partials/popupRegister.html",
		)),
	}
}

// baseData builds the common data map shared by every handler.
// Auth state comes from the session cookie.
// Popup state comes from query params — no global variables needed.
func (b *BaseHandler) baseData(r *http.Request) map[string]interface{} {
	userID, err := b.SessionService.GetUserFromSession(r)
	if err != nil {
		userID = 0
	}

	loginError := ""
	if r.URL.Query().Get("error") == "1" {
		loginError = "Wrong Username or Password"
	}
	registerError := ""
	if r.URL.Query().Get("error") == "2" {
		registerError = "Username or email already exist"
	}
	sessionError := ""
	if r.URL.Query().Get("error") == "3" {
		sessionError = "Session already exist"
	}
	postError := ""
	if r.URL.Query().Get("error") == "4" {
		postError = "Le contenu du post ne peut pas être vide"
	}

	return map[string]interface{}{
		"LoggedIn":      userID != 0,
		"UserId":        userID,
		"ShowLogin":     r.URL.Query().Get("show") == "login",
		"ShowRegister":  r.URL.Query().Get("show") == "register",
		"RegisterError": registerError,
		"LoginError":    loginError,
		"SessionError":  sessionError,
		"PostError":     postError,
	}
}

// HandleActions handles login (action=1) and register (action=2) POST requests.
// These actions come from the header partials shared across all pages.
// If a redirect is issued, a non-nil error is returned — callers must return immediately.
func (b *BaseHandler) HandleActions(w http.ResponseWriter, r *http.Request) (*model.User, error) {
	if r.Method != http.MethodPost {
		return nil, nil
	}

	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	userID, err := b.SessionService.GetUserFromSession(r)
	loggedIn := err == nil && userID != 0
	action := r.FormValue("action")

	if !loggedIn {
		switch action {
		case "1": // login
			name := r.FormValue("name")
			password := r.FormValue("password")

			user, err := b.UserService.LoginUser(name, password)
			if err != nil {
				fmt.Println("loginUser:", err)
				http.Redirect(w, r, r.URL.Path+"?show=login&error=1", http.StatusSeeOther)
				return nil, fmt.Errorf("login failed")
			}

			exists, err := b.SessionService.HasActiveSession(user.Id)
			if err != nil {
				fmt.Println("HasActiveSession:", err)
				http.Redirect(w, r, r.URL.Path+"?show=login&error=1", http.StatusSeeOther)
				return nil, fmt.Errorf("session check failed")
			}

			if exists {
				http.Redirect(w, r, r.URL.Path+"?show=login&error=3", http.StatusSeeOther)
				return nil, fmt.Errorf("session already active")
			}

			token, err := b.SessionService.CreateSession(user)
			if err != nil {
				fmt.Println("CreateSession:", err)
				http.Redirect(w, r, r.URL.Path+"?show=login&error=1", http.StatusSeeOther)
				return nil, fmt.Errorf("session creation failed")
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    token,
				HttpOnly: true,
				Expires:  time.Now().Add(60 * time.Minute),
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			})

			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return &user, fmt.Errorf("redirected")

		case "2": // register
			name := r.FormValue("name")
			email := r.FormValue("email")
			password := r.FormValue("password")

			if err := b.UserService.CreateUser(name, email, password); err != nil {
				fmt.Println("CreateUser:", err)
				http.Redirect(w, r, r.URL.Path+"?show=register&error=2", http.StatusSeeOther)
				return nil, fmt.Errorf("register failed")
			}

			user, err := b.UserService.LoginUser(name, password)
			if err != nil {
				fmt.Println("loginUser:", err)
				http.Redirect(w, r, r.URL.Path+"?show=login&error=1", http.StatusSeeOther)
				return nil, fmt.Errorf("login failed")
			}
			token, err := b.SessionService.CreateSession(user)
			if err != nil {
				fmt.Println("CreateSession:", err)
				http.Redirect(w, r, r.URL.Path+"?show=login&error=1", http.StatusSeeOther)
				return nil, fmt.Errorf("session creation failed")
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    token,
				HttpOnly: true,
				Expires:  time.Now().Add(60 * time.Minute),
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			})

			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return &user, fmt.Errorf("redirected")
		}
	}

	return nil, nil
}

func (b *BaseHandler) GetCurrentUser(r *http.Request) (int, bool) {
	userID, err := b.SessionService.GetUserFromSession(r)
	if err != nil || userID == 0 {
		return 0, false
	}
	return userID, true
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)

		tmpl, ok := LoadTemplates()["page404"]

		if !ok {
			http.Error(w, "Page Not Found", http.StatusNotFound)
			return
		}

		err := tmpl.Execute(w, nil)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func HandleInternalServerError(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {

			if err := recover(); err != nil {

				log.Printf("INTERNAL ERROR: %v", err)

				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusInternalServerError)

				tmpl, ok := LoadTemplates()["page500"]

				if !ok {

					http.Error(
						w,
						"Internal Server Error",
						http.StatusInternalServerError,
					)

					return
				}

				execErr := tmpl.Execute(w, nil)

				if execErr != nil {

					http.Error(
						w,
						"Internal Server Error",
						http.StatusInternalServerError,
					)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session") // use YOUR cookie name

		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

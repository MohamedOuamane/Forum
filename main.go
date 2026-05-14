package main

import (
	"context"
	"database/sql"
	"fmt"
	database "forum/database/dbhelper"
	forum "forum/handler"
	service "forum/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./database/forum.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := database.InitDb(db); err != nil {
		log.Fatal(err)
	}

	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal("Error load .env")
	}

	userService := service.NewUserService(db)
	categoryService := service.NewCategoryService(db)
	likeservice := service.NewLikeService(db, userService)
	imgService := service.NewImgService(db)
	postService := service.NewPostService(db, categoryService, imgService, likeservice, userService)
	sessionService := service.NewSessionService(db)
	commentService := service.NewCommentService(db, likeservice, userService)

	// cleanup sessions
	sessionService.StartCleanup(5 * time.Minute)

	fs := http.FileServer(http.Dir("STATIC"))
	http.Handle("/STATIC/", http.StripPrefix("/STATIC/", fs))

	postGenericHandler := &forum.PostGenericHandler{
		UserService:     userService,
		ImgService:      imgService,
		SessionService:  sessionService,
		PostService:     postService,
		CategoryService: categoryService,
		CommentService:  commentService,
		LikeService:     likeservice,
		TemplateCache:   forum.LoadTemplates(),
	}

	homeHandler := &forum.HomeHandler{
		BaseHandler: &forum.BaseHandler{
			UserService:     userService,
			ImgService:      imgService,
			SessionService:  sessionService,
			PostService:     postService,
			CategoryService: categoryService,
			CommentService:  commentService,
			LikeService:     likeservice,
			TemplateCache:   forum.LoadTemplates(),
		},
	}

	postHandler := &forum.PostHandler{
		BaseHandler: &forum.BaseHandler{
			UserService:     userService,
			PostService:     postService,
			CategoryService: categoryService,
			SessionService:  sessionService,
			CommentService:  commentService,
			LikeService:     likeservice,
			ImgService:      imgService,
			TemplateCache:   forum.LoadTemplates(),
		},
	}

	profilHandler := &forum.ProfilHandler{
		BaseHandler: &forum.BaseHandler{
			UserService:     userService,
			CategoryService: categoryService,
			SessionService:  sessionService,
			CommentService:  commentService,
			ImgService:      imgService,
			PostService:     postService,
			TemplateCache:   forum.LoadTemplates(),
			LikeService:     likeservice,
		},
	}

	categoryHandler := &forum.CategoryHandler{
		BaseHandler: &forum.BaseHandler{
			UserService:     userService,
			SessionService:  sessionService,
			PostService:     postService,
			CategoryService: categoryService,
			CommentService:  commentService,
			LikeService:     likeservice,
			TemplateCache:   forum.LoadTemplates(),
			ImgService:      imgService,
		},
	}

	googleHandler := &forum.GoogleHandler{
		UserService:    userService,
		SessionService: sessionService,
	}

	githubHandler := &forum.GithubHandler{
		UserService:    userService,
		SessionService: sessionService,
	}

	http.HandleFunc("/post/action", postGenericHandler.HandleActions)
	http.HandleFunc("/home", homeHandler.HomeHandler)
	http.HandleFunc("/login", googleHandler.HandleGoogleLogin)
	http.HandleFunc("/callback", googleHandler.HandleGoogleCallback)
	http.HandleFunc("/loginGithub", githubHandler.HandleGithubLogin)
	http.HandleFunc("/auth/github/callback", githubHandler.HandleGithubCallback)
	http.HandleFunc("/", forum.ErrorHandler)
	http.HandleFunc("/post/", postHandler.PostHandler)
	http.HandleFunc("/logout", sessionService.Logout)
	http.HandleFunc("/category/", categoryHandler.CategoryHandler)
	http.HandleFunc("/profil/", profilHandler.ProfilHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: forum.HandleInternalServerError(http.DefaultServeMux),
	}

	// 🔥 Start server in goroutine
	go func() {
		fmt.Println("Server running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// 🧠 Catch Docker stop signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// graceful shutdown HTTP
	if err := server.Shutdown(ctx); err != nil {
		log.Println("Server shutdown error:", err)
	}

	// close DB proprement
	if err := db.Close(); err != nil {
		log.Println("DB close error:", err)
	}

	fmt.Println("Server stopped cleanly")
}

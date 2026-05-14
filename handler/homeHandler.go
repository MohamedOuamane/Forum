package forum

import (
	"fmt"
	model "forum/model"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HomeHandler struct {
	BaseHandler *BaseHandler
}

func (h *HomeHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	userID, err := h.BaseHandler.SessionService.GetUserFromSession(r)

	// ================= POST =================
	if r.Method == http.MethodPost {
		// ✅ Handle like/dislike/delete actions
		_, err := h.BaseHandler.HandleActions(w, r)
		if err != nil {
			return
		}

		if err != nil || userID == 0 {
			http.Redirect(w, r, "/home?show=login", http.StatusSeeOther)
			return
		}

		user, err := h.BaseHandler.UserService.GetUser(userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		contentType := r.Header.Get("Content-Type")

		// ===== MULTIPART (CREATE POST) =====
		if strings.Contains(contentType, "multipart/form-data") {
			r.Body = http.MaxBytesReader(w, r.Body, 20<<20)

			if err := r.ParseMultipartForm(20 << 20); err != nil {
				http.Error(w, "File too large", http.StatusBadRequest)
				return
			}

			action := r.FormValue("action")

			if action == "3" {
				title := r.FormValue("title")
				categories := r.Form["selected_categories"]
				content := r.FormValue("content")

				postID, err := h.BaseHandler.PostService.CreatePost(user, title, content, categories)
				if err != nil {
					fmt.Println("create post err:", err)
				} else {
					if files := r.MultipartForm.File["images"]; len(files) > 0 {
						for i, fh := range files {
							relPath, err := h.BaseHandler.ImgService.SavePostImage(postID, fh)
							if err != nil {
								fmt.Println("save post image err:", err)
								continue
							}
							if err := h.BaseHandler.ImgService.AddPostImage(postID, relPath, i); err != nil {
								fmt.Println("add post image err:", err)
							}
						}
					}
				}
			}

		} else {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// ================= GET =================
	categories, err := h.BaseHandler.CategoryService.GetCategories()
	if err != nil {
		log.Fatal(err)
	}

	posts, err := h.BaseHandler.PostService.GetPosts()
	if err != nil {
		log.Fatal(err)
	}

	contributor, err := h.BaseHandler.UserService.GetTopContributors()
	if err != nil {
		log.Fatal(err)
	}

	active, err := h.BaseHandler.UserService.CountActiveSessions()
	if err != nil {
		log.Fatal(err)
	}

	var postViews []model.PostView
	for _, p := range posts {
		// Get user's like status for this post
		userLike := 0
		if userID != 0 {
			userLike, _ = h.BaseHandler.LikeService.GetUserLikeStatus(p.Id, userID)
		}

		var categoryIds []string
		for _, c := range p.Category {
			id := strconv.Itoa(c.Id)
			categoryIds = append(categoryIds, id)
		}

		postViews = append(postViews, model.PostView{
			Post:             p,
			UserId:           userID,
			ImagesString:     strings.Join(p.Images, ","),
			CategoriesString: strings.Join(categoryIds, ","),
			UserLike:         userLike,
		})
	}

	data := h.BaseHandler.baseData(r)
	data["Title"] = "Forum"
	data["Categories"] = categories
	data["Posts"] = postViews
	data["contributor"] = contributor
	data["active"] = active
	

	LoadTemplates()["home"].Execute(w, data)
}
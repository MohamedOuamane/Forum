package forum

import (
	"fmt"
	model "forum/model"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type CategoryHandler struct {
	*BaseHandler
}

func (h *CategoryHandler) CategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	_, err := h.HandleActions(w, r)
	if err != nil {
		return
	}

	// Parse category ID
	idStr := strings.TrimPrefix(r.URL.Path, "/category/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		LoadTemplates()["page400"].Execute(w, nil)
		return
	}

	userID, err := h.SessionService.GetUserFromSession(r)

	// ================= POST =================
	if r.Method == http.MethodPost {

		if err != nil || userID == 0 {
			http.Redirect(w, r, fmt.Sprintf("/category/%d?show=login", id), http.StatusSeeOther)
			return
		}

		user, err := h.UserService.GetUser(userID)
		if err != nil {
			fmt.Println("get user err:", err)
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		contentType := r.Header.Get("Content-Type")

		// ===== MULTIPART (CREATE POST) =====
		if strings.Contains(contentType, "multipart/form-data") {

			r.Body = http.MaxBytesReader(w, r.Body, 20<<20)

			if err := r.ParseMultipartForm(20 << 20); err != nil {
				http.Error(w, "Request too large", http.StatusBadRequest)
				return
			}

			action := r.FormValue("action")

			if action == "3" {
				title := r.FormValue("title")
				categories := r.Form["selected_categories"]
				content := r.FormValue("content")

				postID, err := h.PostService.CreatePost(user, title, content, categories)
				if err != nil {
					fmt.Println("create post err:", err)
				} else {
					if files := r.MultipartForm.File["images"]; len(files) > 0 {
						for i, fh := range files {
							relPath, err := h.ImgService.SavePostImage(postID, fh)
							if err != nil {
								fmt.Println("save post image err:", err)
								continue
							}
							if err := h.ImgService.AddPostImage(postID, relPath, i); err != nil {
								fmt.Println("add post image err:", err)
							}
						}
					}
				}
			}

		} else {
			//REMEBMER TO REMOVE
			// ===== NORMAL FORM (future-safe) =====
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			action := r.Form.Get("action")

			// no other actions here yet
			_ = action
		}

		http.Redirect(w, r, fmt.Sprintf("/category/%d", id), http.StatusSeeOther)
		return
	}

	// ================= GET =================
	categories, err := h.CategoryService.GetCategories()
	if err != nil {
		log.Fatal(err)
	}

	postCategories, err := h.PostService.GetPostsForCategory(id)
	if err != nil {
		fmt.Println("get posts for category err:", err)
	}
	if len(postCategories) == 0 {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		LoadTemplates()["page400"].Execute(w, nil)
		return
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
	for _, p := range postCategories {
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
		})
	}

	data := h.baseData(r)
	data["Title"] = "Forum"
	data["Categories"] = categories
	data["Posts"] = postViews
	data["contributor"] = contributor
	data["active"] = active

	LoadTemplates()["category"].Execute(w, data)
}

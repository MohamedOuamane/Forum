package forum

import (
	"fmt"
	model "forum/model"
	"net/http"
	"strconv"
	"strings"
)

type ProfilHandler struct {
	*BaseHandler
}

func (h *ProfilHandler) ProfilHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	changeError := ""

	_, err := h.HandleActions(w, r)
	if err != nil {
		return
	}

	var user model.User
	idStr := strings.TrimPrefix(r.URL.Path, "/profil/")
	if idStr != "" {

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid profil ID", http.StatusBadRequest)
			return
		}
		user, err = h.UserService.GetUser(id)
		if err != nil {
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
			return
		}
		userID, err := h.SessionService.GetUserFromSession(r)
		if err != nil || userID == 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

	} else {
		userID, err := h.SessionService.GetUserFromSession(r)
		if err != nil || userID == 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		user, err = h.UserService.GetUser(userID)
		if err != nil {
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
			return
		}
	}

	// ================= POST =================
	if r.Method == http.MethodPost {

		contentType := r.Header.Get("Content-Type")

		// ===== MULTIPART (post creation + avatar) =====
		if strings.Contains(contentType, "multipart/form-data") {

			r.Body = http.MaxBytesReader(w, r.Body, 20<<20)

			if err := r.ParseMultipartForm(20 << 20); err != nil {
				http.Error(w, "Form error", http.StatusBadRequest)
				return
			}

			action := r.FormValue("action")

			switch action {

			case "3": // Create post
				title := r.FormValue("title")
				categories := r.Form["selected_categories"]
				content := r.FormValue("content")

				postID, err := h.PostService.CreatePost(user, title, content, categories)
				if err != nil {
					fmt.Println("create post err:", err)
					break
				}

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

			case "ChangeAvatar":
				files := r.MultipartForm.File["avatar"]
				if len(files) == 0 {
					http.Error(w, "No file uploaded", http.StatusBadRequest)
					return
				}

				relPath, err := h.ImgService.SaveAvatar(user.Id, files[0])
				if err != nil {
					fmt.Println("save avatar err:", err)
					http.Error(w, "Failed to save avatar", http.StatusInternalServerError)
					return
				}

				if err := h.ImgService.UpdateAvatar(user.Id, relPath); err != nil {
					fmt.Println("update avatar err:", err)
					http.Error(w, "Failed to update avatar", http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/profil/", http.StatusSeeOther)
			}

		} else {
			// ===== NORMAL FORM (future-safe) =====
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			action := r.Form.Get("action")
			switch action {
			case "changePassword":
				oldPassword := r.FormValue("currentPassword")
				newPassword := r.FormValue("newPassword")
				confirmPassword := r.FormValue("confirmPassword")
				if !h.UserService.CheckPassword(oldPassword, user.Username) {
					http.Redirect(w, r, r.URL.Path+"?show=changePasswordop&changeError=wrong", http.StatusSeeOther)
					break
				}
				if strings.Compare(newPassword, confirmPassword) != 0 {
					http.Redirect(w, r, r.URL.Path+"?show=changePassworddm&changeError=wrong", http.StatusSeeOther)
					break
				}
				err := h.UserService.ChangePassword(newPassword, user.Id)
				if err != nil {
					http.Error(w, "Failed to change password", http.StatusBadRequest)
				}
			case "changeName":
				newName := strings.TrimSpace(r.FormValue("newName"))

				// Validation
				if newName == "" {
					http.Redirect(w, r, r.URL.Path+"?show=changeName&error=empty", http.StatusSeeOther)
					return
				}

				if len(newName) < 3 {
					http.Redirect(w, r, r.URL.Path+"?show=changeName&error=short", http.StatusSeeOther)
					return
				}

				if len(newName) > 20 {
					http.Redirect(w, r, r.URL.Path+"?show=changeName&error=long", http.StatusSeeOther)
					return
				}

				// Update
				err := h.UserService.ChangeUsername(newName, user.Id)
				if err != nil {
					http.Redirect(w, r, r.URL.Path+"?show=changeName&error=fail", http.StatusSeeOther)
					return
				}
			}

		}

		http.Redirect(w, r, "/profil/", http.StatusSeeOther)
		return
	}

	// ================= GET =================
	likedComments, err := h.LikeService.GetLikedCommentsByUser(user.Id)
	if err != nil {
		http.Error(w, "Failed to get liked comments", http.StatusInternalServerError)
		return
	}
	posts, err := h.PostService.GetPostsByUser(user.Id)
	if err != nil {
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
	}

	comments, err := h.CommentService.GetCommentsUser(user.Id)
	if err != nil {
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
	}
	categories, err := h.CategoryService.GetCategories()
	if err != nil {
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
	}
	likedPosts, err := h.LikeService.GetLikedPostsByUser(user.Id)
	if err != nil {
		http.Error(w, "Failed to get liked posts", http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("show") == "changePassworddm" {
		changeError = "New passwords don't match"
	}
	if r.URL.Query().Get("show") == "changePasswordop" {
		changeError = "Old password not correct"
	}
	nameError := ""

	if r.URL.Query().Get("show") == "changeName" {
		switch r.URL.Query().Get("error") {
		case "empty":
			nameError = "Name cannot be empty"
		case "short":
			nameError = "Name must be at least 3 characters"
		case "long":
			nameError = "Name must be less than 20 characters"
		case "fail":
			nameError = "Failed to change name"
		}
	}

	data := h.baseData(r)
	data["Title"] = "Forum"
	data["Categories"] = categories
	data["Posts"] = posts
	data["Comments"] = comments
	data["Likes"] = likedPosts
	data["Username"] = user.Username
	data["Email"] = user.Email
	data["AvatarPath"] = h.ImgService.GetAvatar(user.Id)
	data["LikedComments"] = likedComments
	data["ChangeError"] = changeError
	data["NameError"] = nameError
	data["Id"] = user.Id
	LoadTemplates()["profil"].Execute(w, data)
}

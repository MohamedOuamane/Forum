package forum

import (
	"fmt"
	service "forum/service"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type PostGenericHandler struct {
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

func (p *PostGenericHandler) HandleActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "Cannot parse form", http.StatusBadRequest)
			return
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Cannot parse form", http.StatusBadRequest)
			return
		}
	}
	
	postId := 0
	commentId := 0
	action := r.FormValue("action")
	
	if action != "create_post" {
		// Try to get postId (for post actions)
		idStr := r.FormValue("postId")
		if idStr != "" {
			postIdTemp, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid post ID", http.StatusBadRequest)
				return
			}
			postId = postIdTemp
		}
		
		// Try to get commentId (for comment actions)
		commentIdStr := r.FormValue("commentId")
		if commentIdStr != "" {
			commentIdTemp, err := strconv.Atoi(commentIdStr)
			if err != nil {
				http.Error(w, "Invalid comment ID", http.StatusBadRequest)
				return
			}
			commentId = commentIdTemp
		}
	}

	userID, err := p.SessionService.GetUserFromSession(r)
	if err != nil || userID == 0 {
		switch {
		case strings.Contains(r.URL.Path, "/home"):
			http.Redirect(w, r, "/home?show=login", http.StatusSeeOther)
		case strings.Contains(r.URL.Path, "/post"):
			http.Redirect(w, r, fmt.Sprintf("/post/%d?show=login", postId), http.StatusSeeOther)
		}
		return
	}

	user, err := p.UserService.GetUser(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	switch action {
	case "delete_post":
		currentUrl := r.FormValue("currentUrl")
		comments, err := p.CommentService.GetCommentsPost(postId)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range comments {
			p.CommentService.DeleteComment(v.Id, user.Id)
		}
		err = p.PostService.DeletePost(postId, user.Id)
		if err != nil {
			fmt.Println("DELETE ERROR:", err)
			break
		} else {
			switch {
			case strings.Contains(currentUrl, "/home"):
				http.Redirect(w, r, "/home", http.StatusSeeOther)
			case strings.Contains(currentUrl, "/post"):
				http.Redirect(w, r, "/home", http.StatusSeeOther)
			case strings.Contains(currentUrl, "/category/"):
				http.Redirect(w, r, currentUrl, http.StatusSeeOther)
			}
			return
		}
		
	case "create_post":
		currentUrl := r.FormValue("currentUrl")
		title := r.FormValue("title")
		categories := r.Form["selected_categories"]
		content := r.FormValue("content")

		postID, err := p.PostService.CreatePost(user, title, content, categories)
		if err != nil {
			fmt.Println("create post err:", err)
		} else {
			if files := r.MultipartForm.File["images"]; len(files) > 0 {
				for i, fh := range files {
					relPath, err := p.ImgService.SavePostImage(postID, fh)
					if err != nil {
						fmt.Println("save post image err:", err)
						continue
					}
					if err := p.ImgService.AddPostImage(postID, relPath, i); err != nil {
						fmt.Println("add post image err:", err)
					}
				}
			}
		}
		switch {
		case strings.Contains(currentUrl, "/home"):
			http.Redirect(w, r, "/home", http.StatusSeeOther)
		case strings.Contains(currentUrl, "/post"):
			http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
		case strings.Contains(currentUrl, "/category/"):
			http.Redirect(w, r, currentUrl, http.StatusSeeOther)
		case strings.Contains(currentUrl, "/profil/"):
			http.Redirect(w, r, currentUrl, http.StatusSeeOther)
		}
		return
		
	case "edit_post":
		currentUrl := r.FormValue("currentUrl")
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["selected_categories"]
		deletedImages := r.FormValue("deletedImages")

		err := p.PostService.EditPost(postId, content, title, categories, deletedImages)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Cannot edit post", http.StatusBadRequest)
		}

		if files := r.MultipartForm.File["images"]; len(files) > 0 {
			for i, fh := range files {
				relPath, err := p.ImgService.SavePostImage(int64(postId), fh)
				if err != nil {
					fmt.Println("save image err:", err)
					continue
				}
				p.ImgService.AddPostImage(int64(postId), relPath, i)
			}
		}

		switch {
		case strings.Contains(currentUrl, "/home"):
			http.Redirect(w, r, "/home", http.StatusSeeOther)
		case strings.Contains(currentUrl, "/post"):
			http.Redirect(w, r, fmt.Sprintf("/post/%d", postId), http.StatusSeeOther)
		}
		return
		
	case "like":
		err = p.LikeService.LikeDislikePost(1, postId, user.Id)

	case "dislike":
		err = p.LikeService.LikeDislikePost(-1, postId, user.Id)

	case "remove_like":
		err = p.LikeService.DeleteLikeDislikePost(postId, user.Id)

	case "remove_dislike":
		err = p.LikeService.DeleteLikeDislikePost(postId, user.Id)

	// ✅ ADD COMMENT LIKE/DISLIKE CASES
	case "like_comment":
		err = p.LikeService.LikeDislikeComment(1, postId, user.Id, commentId)

	case "dislike_comment":
		err = p.LikeService.LikeDislikeComment(-1, postId, user.Id, commentId)

	case "remove_like_comment":
		err = p.LikeService.DeleteLikeDislikeComments(commentId, user.Id)

	case "remove_dislike_comment":
		err = p.LikeService.DeleteLikeDislikeComments(commentId, user.Id)
	}

	// ✅ Set JSON content type BEFORE writing response
	w.Header().Set("Content-Type", "application/json")

	// Return JSON response for like/dislike actions
	switch action {
	case "like", "dislike", "remove_like", "remove_dislike":
		likes, dislikes, err := p.LikeService.GetLikeNbrPost(postId)
		if err != nil {
			fmt.Println("Error getting likes:", err)
		}
		fmt.Fprintf(w, `{"likes": %d, "dislikes": %d}`, likes, dislikes)
		return

	// ✅ ADD COMMENT LIKE/DISLIKE RESPONSE CASES
	case "like_comment", "dislike_comment", "remove_like_comment", "remove_dislike_comment":
		likes, dislikes, err := p.LikeService.GetLikeNbrComment(commentId)
		if err != nil {
			fmt.Println("Error getting comment likes:", err)
		}
		fmt.Fprintf(w, `{"likes": %d, "dislikes": %d}`, likes, dislikes)
		return
	}
}
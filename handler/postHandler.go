package forum

import (
	"fmt"
	forum "forum/model"
	model "forum/model"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type PostHandler struct {
	*BaseHandler
}

func (h *PostHandler) PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	_, err := h.HandleActions(w, r)
	if err != nil {
		return
	}

	// Parse post ID
	idStr := strings.TrimPrefix(r.URL.Path, "/post/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.PostService.GetPost(id)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		LoadTemplates()["page400"].Execute(w, nil)
		return
	}

	categories, err := h.CategoryService.GetCategories()
	if err != nil {
		log.Fatal(err)
	}

	comments, err := h.CommentService.GetCommentsPost(post.Id)
	if err != nil {
		log.Fatal(err)
	}

	userID, err := h.SessionService.GetUserFromSession(r)

	// ✅ Fetch like counts and user like status for comments
	for i := range comments {
		comments[i].Likes, comments[i].Dislikes, _ = h.LikeService.GetLikeNbrComment(comments[i].Id)
		if userID != 0 {
			comments[i].UserLike, _ = h.LikeService.GetUserCommentLikeStatus(comments[i].Id, userID)
		}
	}

	// ================= POST =================
	if r.Method == http.MethodPost {

		if err != nil || userID == 0 {
			http.Redirect(w, r, fmt.Sprintf("/post/%d?show=login", id), http.StatusSeeOther)
			return
		}

		user, err := h.UserService.GetUser(userID)
		if err != nil {
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
					http.Redirect(w, r, fmt.Sprintf("/post/%d?error=4", id), http.StatusSeeOther)
					return
				} else {
					if files := r.MultipartForm.File["images"]; len(files) > 0 {
						for i, fh := range files {
							relPath, err := h.ImgService.SavePostImage(postID, fh)
							if err != nil {
								fmt.Println("save image err:", err)
								continue
							}
							h.ImgService.AddPostImage(postID, relPath, i)
						}
					}
				}
			}

		} else {
			// ===== NORMAL FORM =====
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			action := r.Form.Get("action")

			switch action {

			case "4":
				comment := strings.TrimSpace(r.Form.Get("Comment"))
				if comment == "" {
					http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
					return
				}
				if err := h.CommentService.CreateComment(user.Id, id, comment); err != nil {
					fmt.Println("create comment err:", err)
				}

			case "5":
				h.LikeService.LikeDislikePost(1, id, user.Id)

			case "6":
				h.LikeService.LikeDislikePost(-1, id, user.Id)
			case "like_comment":
				commentID, err := strconv.Atoi(r.Form.Get("comment_id"))
				if err == nil {
					h.LikeService.LikeDislikeComment(1, id, user.Id, commentID)
				}

			case "dislike_comment":
				commentID, err := strconv.Atoi(r.Form.Get("comment_id"))
				if err == nil {
					h.LikeService.LikeDislikeComment(-1, id, user.Id, commentID)
				}

			case "delete_post":
				fmt.Println("action:", action)
				for _, v := range comments {
					h.CommentService.DeleteComment(v.Id, user.Id)

				}
				err := h.PostService.DeletePost(id, user.Id)
				if err != nil {
					fmt.Println("DELETE ERROR:", err)
					break
				} else {
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}

			case "delete_comment":
				fmt.Println("action:", action)
				commentID, err := strconv.Atoi(r.FormValue("comment_id"))
				if err != nil {
					http.Error(w, "Invalid comment ID", http.StatusBadRequest)
					return
				}
				h.CommentService.DeleteComment(commentID, user.Id)

			case "edit_comment":
				comment := strings.TrimSpace(r.Form.Get("NewComment"))
				if comment == "" {
					http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
					return
				}

				commentID, err := strconv.Atoi(r.FormValue("comment_ID"))
				if err != nil {
					http.Error(w, "Invalid comment ID", http.StatusBadRequest)
					return
				}

				if err := h.CommentService.EditComment(comment, commentID, user.Id); err != nil {
					fmt.Println("edit comment err:", err)
				}
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/post/%d", id), http.StatusSeeOther)
		return
	}

	// ================= GET =================
	like, dislike, _ := h.LikeService.GetLikeNbrPost(post.Id)

	// ✅ Fetch user's like status for this post
	userPostLike := 0
	if userID != 0 {
		userPostLike, _ = h.LikeService.GetUserLikeStatus(post.Id, userID)
	}

	images, err := h.ImgService.GetPostImages(post.Id)
	if err != nil {
		fmt.Println("get post images err:", err)
	}

	var postCategories []forum.Post
	for _, v := range post.Category {
		posts, err := h.PostService.GetPostsCategory(v.Id)
		if err != nil {
			continue
		}
		for _, c := range posts {
			postc, err := h.PostService.GetPost(c)
			if err != nil {
				continue
			}
			postCategories = append(postCategories, postc)
		}
	}

	var commentsViews []model.CommentView
	for _, CV := range comments {
		commentsViews = append(commentsViews, model.CommentView{
			Comment: CV,
			UserId:  userID,
		})
	}

	var categoryIds []string
	for _, c := range post.Category {
		catId := strconv.Itoa(c.Id)
		categoryIds = append(categoryIds, catId)
	}

	active, err := h.BaseHandler.UserService.CountActiveSessions()
	if err != nil {
		log.Fatal(err)
	}

	postView := model.PostView{
		Post:             post,
		UserId:           userID,
		ImagesString:     strings.Join(post.Images, ","),
		CategoriesString: strings.Join(categoryIds, ","),
		UserLike:         userPostLike, // ✅ ADD THIS
	}

	data := h.baseData(r)
	data["Title"] = "Forum"
	data["Post"] = postView
	data["Categories"] = categories
	data["Posts"] = postCategories
	data["Comments"] = commentsViews
	data["Likes"] = like
	data["Dislike"] = dislike
	data["Images"] = images
	data["active"] = active

	LoadTemplates()["post"].Execute(w, data)
	if err != nil {
		log.Println("TEMPLATE ERROR:", err)
	}
}
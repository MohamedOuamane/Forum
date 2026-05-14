document.addEventListener("DOMContentLoaded", () => {
   
    // =========================
    // ✅ VALIDATE COMMENT CONTENT - CHECK FOR EMPTY/WHITESPACE ONLY
    // =========================
    const commentForm = document.querySelector("form");
    const submitBtn = document.getElementById("SendComment");
    
    if (commentForm && submitBtn) {
        submitBtn.addEventListener("click", function(e) {
            const commentText = document.getElementById("CommentArea")?.value.trim() || "";
            
            if (!commentText) {
                e.preventDefault();
                e.stopPropagation();
                alert("Comment cannot be empty or contain only spaces");
                return false;
            }
        });
    }

    // =========================
    // ✏️ EDIT COMMENT
    // =========================
    document.querySelectorAll(".edit-comment").forEach(button => {
        button.addEventListener("click", () => {
            console.log("CLICK EDIT");

            const commentId = button.dataset.commentId;
            const content = button.dataset.commentContent;

            document.getElementById("editCommentId").value = commentId;
            document.getElementById("NewComment").value = content;

            document.getElementById("EditComPopup").style.display = "block";
        });
    });


    // =========================
    // 🗑 DELETE COMMENT
    // =========================
    document.querySelectorAll(".delete-comment").forEach(button => {
        button.addEventListener("click", () => {

            const commentId = button.dataset.commentId;

            document.getElementById("deleteTargetId").value = commentId;

            document.getElementById("deleteModal").style.display = "block";
        });
    });


    // =========================
    // ❌ FERMER DELETE
    // =========================
    const noBtn = document.getElementById("noDeleteBtn");

    if (noBtn) {
        noBtn.addEventListener("click", () => {
            document.getElementById("deleteModal").style.display = "none";
        });
    }

    // =========================
    // 👍 LIKE/DISLIKE COMMENT (TOGGLE)
    // =========================
    document.querySelectorAll(".comment-footer .reaction-btn").forEach(button => {
        if (button.dataset.liked === "true") {
            button.classList.add("checked");
        }
    });

    document.querySelectorAll(".comment-footer .reaction-btn").forEach(button => {
        button.addEventListener("click", async (e) => {
            e.preventDefault();

            const commentId = button.dataset.commentId;
            const postId = button.dataset.postId;
            const userId = button.dataset.userId;
            const baseAction = button.dataset.action;
            const isCurrentlyLiked = button.classList.contains("checked");

            if (userId != 0) {
                try {
                    const finalAction = isCurrentlyLiked ? `remove_${baseAction}_comment` : `${baseAction}_comment`;

                    const res = await fetch("/post/action", {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/x-www-form-urlencoded"
                        },
                        body: `postId=${postId}&commentId=${commentId}&action=${finalAction}`
                    });

                    if (!res.ok) {
                        throw new Error("Erreur serveur");
                    }

                    const data = await res.json();

                    const parent = button.closest(".comment-footer");

                    parent.querySelector(".likeCommentCount").textContent = data.likes;
                    parent.querySelector(".dislikeCommentCount").textContent = data.dislikes;

                    button.classList.toggle("checked");

                    const oppositeAction = baseAction === "like" ? "dislike" : "like";
                    const oppositeButton = parent.querySelector(
                        `[data-action="${oppositeAction}"]`
                    );
                    if (oppositeButton) {
                        oppositeButton.classList.remove("checked");
                    }

                } catch (err) {
                    console.error(err);
                }
            } else {
                openPopup("loginPopup", "mustLogin");
            }
        });
    });

});
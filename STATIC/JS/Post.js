document.addEventListener("DOMContentLoaded", () => {
    // ✅ Initialize post buttons with checked state based on data-liked attribute
    document.querySelectorAll(".like-post .icon-btn").forEach(button => {
        if (button.dataset.liked === "true") {
            button.classList.add("checked");
        }
    });

    // Handle post like/dislike clicks
    document.querySelectorAll(".like-post .icon-btn").forEach(button => {
        button.addEventListener("click", async (e) => {
            e.preventDefault();

            const postId = button.dataset.postId;
            const userId = button.dataset.userId;
            const baseAction = button.dataset.action; // "like" or "dislike"
            
            if (userId != 0) {
                try {
                    // Check if button is already liked/disliked
                    const isLiked = button.classList.contains("checked");
                    
                    // Determine action to send
                    let action;
                    if (isLiked) {
                        // If already liked, send remove_like
                        action = `remove_${baseAction}`;
                    } else {
                        // If not liked, send like
                        action = baseAction;
                    }

                    const res = await fetch("/post/action", {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/x-www-form-urlencoded"
                        },
                        body: `postId=${postId}&action=${action}`
                    });

                    if (!res.ok) {
                        throw new Error("Erreur serveur");
                    }

                    const data = await res.json();

                    const parent = button.closest(".like-post");

                    // Update counts
                    parent.querySelector(".likeCount").textContent = data.likes;
                    parent.querySelector(".dislikeCount").textContent = data.dislikes;

                    // ✅ Toggle the button's checked state
                    button.classList.toggle("checked");

                    // ✅ Remove checked state from the opposite button
                    const otherButton = parent.querySelector(
                        baseAction === "like" ? ".dislike-btn" : ".like-btn"
                    );
                    if (otherButton) {
                        otherButton.classList.remove("checked");
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
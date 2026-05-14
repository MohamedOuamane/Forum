function openPopup(id, mode = "post", triggerBtn = null) {
    let deletedImages = [];
    const el = document.getElementById(id);
    if (el) el.style.display = "block";

    const textarea = document.getElementById("content");
    if (textarea) textarea.value = "";

    const input = document.getElementById("title");
    if (input) input.value = "";

    for (let i = 1; i <= 4; i++) {
        const cb = document.getElementById("cat" + i);

        if (cb) {
            cb.checked = false;
        }
    }
    if (id === "postPopup") {
        
        
        const btn = document.getElementById("Publish");
        const containerImg = document.getElementById("container-img");
        containerImg.innerHTML = "";
        if (btn) {
            if (mode === "edit" && triggerBtn) {
                

                const postId = triggerBtn.dataset.postId;
                const images = triggerBtn.dataset.images;
                const categories = triggerBtn.dataset.categories.trim();
                const title = triggerBtn.dataset.title;
                const content = triggerBtn.dataset.content;
                const hiddenId = document.getElementById("postId");
                hiddenId.value = postId;
                const currentUrl = document.getElementById("currentUrl");
                currentUrl.value = window.location.pathname;
                const cats = categories.split(",");

                cats.forEach(cat => {
                    const cb = document.getElementById("cat"+cat.trim());
                    cb.checked = true;
                });
                if(images){
                    
                    const imgs = images.split(",");
                   

                    imgs.forEach(img => {
                        const wrapper = document.createElement("div");
                        wrapper.style.position = "relative";
                        wrapper.style.display = "inline-block";

                        const imageEl = document.createElement("img");
                        imageEl.src = "/STATIC/uploads/" + img;
                        imageEl.dataset.filename = img;
                        imageEl.style.width = "100px";
                        imageEl.style.height = "100px"
                        imageEl.style.margin = "5px";

                        const deleteEl = document.createElement("button");
                        deleteEl.classList.add("delete-btn");
                        deleteEl.innerHTML = "×";
                        deleteEl.type = "button";

                       deleteEl.onclick = () => {
                            const filename = imageEl.dataset.filename;
                            deletedImages.push(filename);
                            wrapper.remove();
                        };

                        wrapper.appendChild(imageEl);
                        wrapper.appendChild(deleteEl);
                        containerImg.appendChild(wrapper);
                    });
                }
                textarea.value=content
                input.value= title
                btn.value = "edit_post";
                btn.innerText = "Save";
                btn.addEventListener("click", function() {
                    document.getElementById("deletedImages").value = deletedImages.join(",");
                });
            } else {
                const currentUrl = document.getElementById("currentUrl");
                currentUrl.value = window.location.pathname;
                btn.value = "create_post";
                btn.innerText = "Publish";
            }
        }
    } else if (id === "deleteModal") {
        const yesBtn = document.getElementById("yesDeleteBtn");
        const hiddenInput = document.getElementById("deleteTargetId");
        const postId = triggerBtn.dataset.postId
        const commentId = triggerBtn.dataset.commentId
        console.log(commentId)
        if (yesBtn) {
            if (mode === "delete_comment" && triggerBtn) {
                const form = document.getElementById("deleteForm");
                console.log(form)
                form.action = "/post/" + postId;
                const commentId = triggerBtn.dataset.commentId
                yesBtn.value = "delete_comment";
                if (hiddenInput) hiddenInput.value = commentId; // set it here, not on click
            } else {
                yesBtn.value = "delete_post";
                console.log("delete_post")
                const currentUrl = document.getElementById("deleteCurrentUrl")
                currentUrl.value = window.location.pathname;
                if (hiddenInput){
                    hiddenInput.value = postId;
                    hiddenInput.name = "postId";
                }
            }
        }
    } else if (id === "loginPopup") {

        const title = document.getElementById("mustLogin");

        if (title) {

            if (mode === "mustLogin") {
                title.innerText = "You must login first";
            } else {
                title.innerText = "Login";
            }

        }
    }
    // ✅ ADD CATEGORY VALIDATION FOR POST POPUP
    if (id === "postPopup") {
        const publishBtn = document.getElementById("Publish");
        if (publishBtn) {
            publishBtn.addEventListener("click", (e) => {
                const checkboxes = document.querySelectorAll('input[name="selected_categories"]');
                const isChecked = Array.from(checkboxes).some(checkbox => checkbox.checked);
                const errorMsg = document.getElementById("categoryError");
                
                if (!isChecked) {
                    e.preventDefault();
                    if (errorMsg) errorMsg.style.display = "block";
                } else {
                    if (errorMsg) errorMsg.style.display = "none";
                }
            });
        }
    }
}

// ✅ VALIDATE POST CONTENT - CHECK FOR EMPTY/WHITESPACE ONLY
function isValidPostContent() {
    const title = document.getElementById("title")?.value.trim() || "";
    const content = document.getElementById("content")?.value.trim() || "";
    
    // Check if title or content are empty or only whitespace
    if (!title || !content) {
        alert("Title and content cannot be empty or contain only spaces");
        return false;
    }
    
    return true;
}

// Attach validation to publish button
document.addEventListener("DOMContentLoaded", () => {
    const publishBtn = document.getElementById("Publish");
    if (publishBtn) {
        publishBtn.addEventListener("click", (e) => {
            if (!isValidPostContent()) {
                e.preventDefault();
                return false;
            }
        });
    }
});

function closePopup(id) {
    const el = document.getElementById(id);
    if (el) el.style.display = "none";
}

function switchPopup(closeId, openId) {
    closePopup(closeId);
    openPopup(openId);
}

function openEditPopup(button) {
  document.getElementById("EditComPopup").style.display = "block";

  // injecte l'id
  document.getElementById("editCommentId").value = button.dataset.commentId;

  // injecte le contenu
  document.getElementById("NewComment").value = button.dataset.commentContent;
}

// Only attach avatar preview listener if the element exists (profile page only)
const avatarInput = document.getElementById('avatarInput');
if (avatarInput) {
    avatarInput.addEventListener('change', function(e) {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = function(e) {
                document.getElementById('avatarPreview').src = e.target.result;
            };
            reader.readAsDataURL(file);
        }
    });
}

// Close any popup when clicking outside it — null-safe for pages where a popup doesn't exist
window.onclick = function(event) {
    ["loginPopup", "registerPopup", "postPopup", "avatarPopup", "changePasswordPopup", "EditComPopup", "deleteModal"].forEach(function(id) {
        const el = document.getElementById(id);
        if (el && event.target === el) {
            el.style.display = "none";
        }
    });
};

// Attach the No button for popupDelete
const noDeleteBtn = document.getElementById('noDeleteBtn');
if (noDeleteBtn) {
    noDeleteBtn.addEventListener('click', function() {
        closePopup('deleteModal');
    });
}
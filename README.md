### 🔨 Build & Start the Project

To build the Docker images and start the containers in detached mode:

```bash
docker compose up -d --build
```

To stop and remove the running containers:

```bash
docker compose down
```


# Task list 
# 2️⃣ SQLite Database

- [X] Install SQLite driver (`sqlite3`)
- [X] Create database connection
- [X] Initialize database on server start
- [ ] Create database schema

---

## Database Tables

### Users

- [X] Create `users` table
- [X] Fields:
  - [X] id
  - [X] role_id
  - [X] username
  - [X] email
  - [X] password
  - [X] created_at
  - [X] avatar

---

### Sessions

- [X] Create `sessions` table
- [ ] Fields:
  - [X] id
  - [X] user_id
  - [X] session_token
  - [X] expires_at
  - [X] nb_connections
---

### Posts

- [X] Create `posts` table
- [ ] Fields:
  - [X] id
  - [X] user_id
  - [X] title
  - [X] content
  - [X] created_at

---

### PostsCategories

- [X] Create `PostCategories` table
- [X] Fields:
  - [X] post_id
  - [X] category_id

---


### Comments

- [X] Create `comments` table
- [X] Fields:
  - [X] id
  - [X] user_id
  - [X] post_id
  - [X] content
  - [X] created_at

---

### Categories

- [X] Create `categories` table
- [X] Fields:
  - [X] id
  - [X] name
  - [X] logo

---

### Likes

- [X] Create `likes` table
- [X] Fields:
  - [X] id
  - [X] user_id
  - [X] post_id
  - [X] comment_id
  - [X] value (1 = like, -1 = dislike)
  - [X] created_at

---

# 3️⃣ Web Server Setup

- [X] Create Go HTTP server
- [X] Setup router
- [X] Implement static file serving
- [X] Implement HTML template rendering
- [X] Create base layout template

---

# 4️⃣ Authentication System

## User Registration

- [X] Create registration page
- [X] Add form validation
- [X] Ask for:
  - [X] email
  - [X] username
  - [X] password
- [X] Check if email already exists
- [ ] <span style="color:red">Return error if email already exists</span>
- [X] Hash password using **bcrypt** (bonus)
- [X] Store user in database

---

## User Login

- [X] Create login page
- [X] Verify email exists
- [X] Verify password matches stored password
- [X] Return error if credentials are invalid
- [ ] Create user session

---

## Sessions & Cookies

- [X] Generate session token
- [X] Store session in database
- [X] Create cookie with expiration date
- [ ] Allow only one active session per user
- [ ] Implement session validation middleware
- [X] Implement logout route

Bonus:

- [ ] Use UUID for session tokens

---

# 5️⃣ Posts System

## Create Posts

- [X] Create post creation form
- [X] Allow only logged-in users to create posts
- [X] Add title field
- [X] Add content field
- [X] Allow selecting one or more categories
- [X] Insert post into database
- [X] Insert category relationships

---

## Display Posts

- [X] Create homepage displaying posts
- [ ] Show:
  - [X] post title
  - [X] post content
  - [X] author
  - [ ] categories
  - [X] creation date
  - [X] likes/dislikes

---

## View Single Post

- [X] Create post detail page
- [X] Display full post
- [X] Display comments

---

# 6️⃣ Comments System

- [X] Create comment form
- [X] Allow only registered users to comment
- [X] Store comments in database
- [X] Display comments under posts
- [X] Show comment author
- [ ] Show comment creation date

---

# 7️⃣ Likes & Dislikes

## Posts

- [X] Add like button
- [X] Add dislike button
- [X] Allow only logged-in users to vote
- [X] Prevent duplicate votes from same user
- [X] Store vote in database
- [X] Display total likes and dislikes

---

## Comments

- [X] Add like button to comments
- [X] Add dislike button to comments
- [X] Store comment votes
- [X] Display total likes and dislikes

---

# 8️⃣ Post Filtering

## Filter by Categories

- [ ] Implement category filtering
- [X] Show posts belonging to selected category

---

## Filter by Created Posts

- [X] Show posts created by logged-in user

---

## Filter by Liked Posts

- [X] Show posts liked by logged-in user

---

# 9️⃣ Error Handling

- [ ] Handle HTTP errors

Required error pages:

- [ ] 400 Bad Request
- [ ] 401 Unauthorized
- [ ] 404 Not Found
- [ ] 500 Internal Server Error

- [ ] Handle database errors
- [ ] Handle invalid form input
- [ ] Handle missing routes

---

# 🔟 Frontend (No Frameworks)

- [X] Create HTML templates
- [X] Create navigation bar 
- [X] Create login form
- [X] Create register form
- [X] Create post creation form
- [X] Create comment form
- [X] Implement basic CSS styling

Restrictions:

- ❌ No frontend frameworks (React, Vue, Angular)

---

# 1️⃣1️⃣ Security

- [ ] Hash passwords using bcrypt
- [ ] Validate all user inputs
- [ ] Prevent SQL injection
- [ ] Sanitize HTML output
- [ ] Validate session cookies

Bonus:

- [ ] Implement CSRF protection

---

# 1️⃣2️⃣ Docker

## Dockerfile

- [ ] Create Dockerfile
- [ ] Build Go application
- [ ] Copy compiled binary
- [ ] Expose application port
- [ ] Run server inside container

---

## Docker Compose (optional)

- [ ] Create docker-compose.yml
- [ ] Configure container service
- [ ] Expose port (example: 8080)
- [ ] Mount database volume

---

# 1️⃣3️⃣ Testing

- [ ] Write unit tests for database functions
- [ ] Test authentication logic
- [ ] Test post creation
- [ ] Test comment creation
- [ ] Test like/dislike system

Run tests:

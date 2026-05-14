# Forum

A modern forum web application built with Go and SQLite, featuring user authentication, posts, comments, likes/dislikes, and category-based browsing.

## Overview

Forum is a full-stack web application designed to let users register, sign in, create posts, interact with content, and manage their own profile activity.  
The project uses server-side rendering with Go templates and a lightweight frontend built with HTML, CSS, and vanilla JavaScript.

## Features

- User registration and login
- Google OAuth authentication (note that you need to provide you own GOOGLE_CLIENT_ID= and GOOGLE_CLIENT_SECRET= for this to work)
- GitHub OAuth authentication (note that you need to provide you own GITHUB_CLIENT_ID= and GITHUB_CLIENT_SECRET= for this to work)
- Create and browse posts
- Comment on posts
- Like and dislike posts and comments
- Browse by categories
- View posts created by the logged-in user
- View posts liked by the logged-in user
- User profile page
- SQLite database integration
- Docker support

## Tech Stack

- **Backend:** Go
- **Database:** SQLite
- **Authentication:** Google OAuth, GitHub OAuth, email/password
- **Frontend:** HTML, CSS, JavaScript
- **Deployment:** Docker, Docker Compose

## Project Structure

- `main.go` — application entry point and server setup
- `handler/` — HTTP request handlers
- `service/` — business logic layer
- `database/` — database helpers and initialization
- `model/` — application data models
- `templates/` — server-rendered HTML templates
- `STATIC/` — styles, scripts, and static assets

## Getting Started

### Prerequisites

- Go 1.25+
- SQLite
- Docker and Docker Compose

### Run with Docker

```bash
docker compose up -d --build

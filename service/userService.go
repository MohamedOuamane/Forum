package forum

import (
	"database/sql"
	"errors"
	"fmt"
	model "forum/model"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	DB *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		DB: db,
	}
}

func (s *UserService) CreateUser(name string, email string, password string) error {
	user := model.User{}
	user.Username = name
	user.Creation = time.Now()
	user.Email = email
	var count int

	err := s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", name).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("Username already exist")
	}
	err = s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("Email already exist")
	}
	avatar := "avatars/Generic-Profile.jpg"
	user.Avatar = &avatar
	//Generate Hash password
	hash, err := HashPassword(password)
	if err != nil {
		panic(err)
	}
	user.Password = &hash
	//get normal id
	rows, err := s.DB.Query("SELECT id FROM Roles WHERE name='USER'")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&user.Role.Id)
		if err != nil {
			return err
		}
	}

	query := `INSERT INTO Users(username, email, password, created_at, role_id, avatar) VALUES (?, ?, ?, ?, ?, ?)`

	_, err = s.DB.Exec(query, user.Username, user.Email, user.Password, user.Creation, user.Role.Id, user.Avatar)
	return err
}

func HashPassword(password string) (string, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
func (s *UserService) CheckPassword(inputPassword string, user string) bool {
	var storedHash string
	err := s.DB.QueryRow(
		"SELECT password FROM users WHERE username = ? OR email = ?",
		user,
		user,
	).Scan(&storedHash)
	if err != nil {

		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
	if err != nil {
		return false
	}

	return true
}
func (s *UserService) LoginUser(name string, password string) (model.User, error) {
	user := model.User{}
	var avatar sql.NullString
	if s.CheckPassword(password, name) {
		err := s.DB.QueryRow(
			"SELECT id, role_id, username, email, password, created_at, avatar  FROM users WHERE username = ? OR email = ?",
			name,
			name,
		).Scan(&user.Id, &user.Role.Id, &user.Username, &user.Email, &user.Password, &user.Creation, &avatar)

		if err != nil {
			return user, err
		}

		if avatar.Valid {
			user.Avatar = &avatar.String
		} else {
			empty := ""
			user.Avatar = &empty
		}
		return user, nil
	}

	return model.User{}, fmt.Errorf("invalid credentials")
}
func (s *UserService) GetUser(id int) (model.User, error) {
	var user model.User
	var avatar sql.NullString
	err := s.DB.QueryRow("SELECT id,role_id, username, email, avatar FROM Users WHERE id= ?", id).Scan(&user.Id, &user.Role.Id, &user.Username, &user.Email, &avatar)
	if avatar.Valid {
		user.Avatar = &avatar.String
	} else {
		empty := ""
		user.Avatar = &empty
	}
	if err != nil {
		fmt.Println("it's not working")
		return user, err
	}

	return user, nil
}

func (s *UserService) FindUserByGoogleId(googleId string) (*model.User, error) {
	var user model.User
	err := s.DB.QueryRow(`
		SELECT id, role_id, username, email, password, google_id, avatar
		FROM Users
		WHERE google_id = ?
	`, googleId).Scan(&user.Id, &user.Role.Id, &user.Username, &user.Email, &user.Password, &user.GoogleId, &user.Avatar)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) FindUserByGithubId(githubId string) (*model.User, error) {
	var user model.User
	err := s.DB.QueryRow(`
		SELECT id, role_id, username, email, password, google_id, avatar
		FROM Users
		WHERE github_id = ?
	`, githubId).Scan(&user.Id, &user.Role.Id, &user.Username, &user.Email, &user.Password, &user.GoogleId, &user.Avatar)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) FindUserByEmail(email string) (*model.User, error) {
	var user model.User

	err := s.DB.QueryRow(`
		SELECT id, role_id, username, email, password, google_id, avatar
		FROM Users WHERE email = ?`, email).
		Scan(&user.Id, &user.Role.Id, &user.Username, &user.Email, &user.Password, &user.GoogleId, &user.Avatar)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) HandleGoogleOAuthUser(gId, email, name, avatar string) (*model.User, error) {

	// 1. Chercher par Google ID
	user, err := s.FindUserByGoogleId(gId)
	if err == nil {
		return user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// 2. Chercher par email
	user, err = s.FindUserByEmail(email)
	if err == nil {
		_, err := s.DB.Exec(`
			UPDATE Users SET google_id = ? WHERE id = ?`,
			gId, user.Id,
		)
		if err != nil {
			return nil, err
		}

		user.GoogleId = &gId
		return user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// 3. Créer nouvel utilisateur
	username := generateUsername(name)

	if avatar == "" {
		avatar = "avatars/Generic-Profile.jpg"
	}

	res, err := s.DB.Exec(`
		INSERT INTO Users (role_id, username, email, google_id, avatar)
		VALUES (2, ?, ?, ?, ?)`,
		username, email, gId, avatar,
	)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()

	return &model.User{
		Id: int(id),
		Role: model.Role{
			Id: 2,
		},
		Username: username,
		Email:    email,
		GoogleId: &gId,
		Avatar:   &avatar,
	}, nil
}

func (s *UserService) HandleGithubOAuthUser(gId, email, name, avatar string) (*model.User, error) {

	// 1. Chercher par Github ID
	user, err := s.FindUserByGithubId(gId)
	if err == nil {
		return user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// 2. Chercher par email
	user, err = s.FindUserByEmail(email)
	if err == nil {
		_, err := s.DB.Exec(`
			UPDATE Users SET github_id = ? WHERE id = ?`,
			gId, user.Id,
		)
		if err != nil {
			return nil, err
		}

		user.GithubId = &gId
		return user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// 3. Créer nouvel utilisateur
	username := generateUsername(name)

	if avatar == "" {
		avatar = "avatars/Generic-Profile.jpg"
	}

	res, err := s.DB.Exec(`
		INSERT INTO Users (role_id, username, email, github_id, avatar)
		VALUES (2, ?, ?, ?, ?)`,
		username, email, gId, avatar,
	)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()

	return &model.User{
		Id: int(id),
		Role: model.Role{
			Id: 2,
		},
		Username: username,
		Email:    email,
		GithubId: &gId,
		Avatar:   &avatar,
	}, nil
}

func generateUsername(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "_")
}

func (s *UserService) ChangePassword(password string, userID int) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err1 := s.DB.Exec(`
			UPDATE Users
			SET password = ?
			WHERE id = ? 
		`, hash, userID)
	if err1 != nil {
		return err1
	}
	return nil
}

func (s *UserService) ChangeUsername(username string, userID int) error {
	_, err := s.DB.Exec(`
		UPDATE Users
		SET username = ?
		WHERE id = ?
	`, username, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) GetTopContributors() ([]model.TopContributor, error) {

	rows, err := s.DB.Query(`
		SELECT u.id, u.username, u.avatar, COUNT(p.id) as post_count
		FROM Users u
		JOIN Posts p ON u.id = p.user_id
		GROUP BY u.id, u.username, u.avatar
		ORDER BY post_count DESC
		LIMIT 5
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributors []model.TopContributor

	for rows.Next() {
		var c model.TopContributor
		var avatar sql.NullString

		err := rows.Scan(&c.ID, &c.Username, &avatar, &c.PostCount)
		if err != nil {
			return nil, err
		}

		if avatar.Valid {
			c.Avatar = avatar.String
		} else {
			c.Avatar = ""
		}

		contributors = append(contributors, c)
	}

	return contributors, nil
}

func (s *UserService) CountActiveSessions() (int, error) {
	var count int

	err := s.DB.QueryRow(`
		SELECT COUNT(DISTINCT user_id)
		FROM Sessions
		WHERE expires_at > CURRENT_TIMESTAMP
	`).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

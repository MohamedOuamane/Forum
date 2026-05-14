package forum

import (
	"database/sql"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

const (
	UploadBase  = "./STATIC/uploads"
	MaxFileSize = 20 << 20
	AvatarSize  = 256
)

type ImgService struct {
	DB *sql.DB
}

func NewImgService(db *sql.DB) *ImgService {
	return &ImgService{
		DB: db,
	}
}

func (i *ImgService) validateAndOpen(fh *multipart.FileHeader) (image.Image, string, error) {
	if fh.Size > MaxFileSize {
		return nil, "", fmt.Errorf("file too large")
	}

	f, err := fh.Open()
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	buf := make([]byte, 512)

	_, err = f.Read(buf)
	if err != nil {
		return nil, "", err
	}

	ct := http.DetectContentType(buf)

	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}

	if !allowed[ct] {
		return nil, "", fmt.Errorf("unsupported type: %s", ct)
	}

	f.Seek(0, io.SeekStart)

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, "", err
	}

	return img, ct, nil
}

func (i *ImgService) SaveAvatar(userID int, fh *multipart.FileHeader) (string, error) {
	src, _, err := i.validateAndOpen(fh)
	if err != nil {
		fmt.Println("the error is here")
		return "", err
	}
	fmt.Println("image decoded successfully, bounds:", src.Bounds())

	thumb := imaging.Fill(src, AvatarSize, AvatarSize, imaging.Center, imaging.Lanczos)
	fmt.Println("image resized successfully")
	dir := filepath.Join(UploadBase, "avatars")
	os.MkdirAll(dir, 0755)

	relPath := fmt.Sprintf("avatars/%d.jpg", userID)
	absPath := filepath.Join(UploadBase, relPath)

	if err := imaging.Save(thumb, absPath); err != nil {
		fmt.Println("imaging.Save error:", err)
		return "", err
	}
	return relPath, nil
}

func (i *ImgService) SavePostImage(postID int64, fh *multipart.FileHeader) (string, error) {

	_, contentType, err := i.validateAndOpen(fh)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(UploadBase, "posts", fmt.Sprintf("%d", postID))
	os.MkdirAll(dir, 0755)

	id := uuid.NewString()

	// ================= GIF =================
	if contentType == "image/gif" {

		relPath := fmt.Sprintf("posts/%d/%s.gif", postID, id)
		absPath := filepath.Join(UploadBase, relPath)

		srcFile, err := fh.Open()
		if err != nil {
			return "", err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(absPath)
		if err != nil {
			return "", err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return "", err
		}

		return relPath, nil
	}

	// ================= AUTRES IMAGES =================

	src, _, err := i.validateAndOpen(fh)
	if err != nil {
		return "", err
	}

	resized := imaging.Resize(src, 1200, 0, imaging.Lanczos)

	relPath := fmt.Sprintf("posts/%d/%s.jpg", postID, id)
	absPath := filepath.Join(UploadBase, relPath)

	if err := imaging.Save(resized, absPath); err != nil {
		return "", err
	}

	return relPath, nil
}

func (s *ImgService) UpdateAvatar(userID int, path string) error {
	_, err := s.DB.Exec(`UPDATE Users SET avatar = ? WHERE id = ?`, path, userID)
	return err
}

func (s *ImgService) AddPostImage(postID int64, path string, position int) error {
	_, err := s.DB.Exec(
		`INSERT INTO Pictures (post_id, path, position) VALUES (?, ?, ?)`,
		postID, path, position,
	)
	return err
}
func (s *ImgService) GetPostImages(postID int) ([]string, error) {
	rows, err := s.DB.Query(
		`SELECT path FROM Pictures WHERE post_id = ? ORDER BY position ASC`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}
func (s *ImgService) GetAvatar(userID int) string {
	var path string
	err := s.DB.QueryRow(`SELECT avatar FROM Users WHERE id = ?`, userID).Scan(&path)
	if err != nil || path == "" {
		return "avatars/Generic-Profile.jpg"
	}
	return path
}

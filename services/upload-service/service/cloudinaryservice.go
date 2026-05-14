package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/runns/shared/models"
	"gorm.io/gorm"
)

// UploadResult holds what we store in the database
type UploadResult struct {
	PublicID  string `json:"public_id"`  // stored in DB — used to delete/transform
	SecureURL string `json:"secure_url"` // full HTTPS URL returned to client
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	Bytes     int    `json:"bytes"`
}

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
	db  *gorm.DB
}

func NewCloudinaryService(db *gorm.DB) (*CloudinaryService, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, errors.New("CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY and CLOUDINARY_API_SECRET must be set")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary init: %w", err)
	}

	return &CloudinaryService{cld: cld, db: db}, nil
}

// UploadAvatar uploads a user's profile photo to Cloudinary.
// It deletes the old photo if one exists, then uploads the new one.
// Only the public_id is stored in the DB; the URL is derived on read.
func (s *CloudinaryService) UploadAvatar(ctx context.Context, userID uint, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// Validate file type
	if err := validateImageType(header.Filename); err != nil {
		return nil, err
	}

	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		return nil, errors.New("file too large: maximum size is 5MB")
	}

	// Fetch user to get existing avatar public_id for deletion
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Delete old avatar from Cloudinary if it exists
	if user.AvatarPublicID != "" {
		_, _ = s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
			PublicID: user.AvatarPublicID,
		})
	}

	// Upload new avatar
	// folder: rydex/avatars/<userID>
	// public_id: user_<userID>  (deterministic — makes re-upload idempotent)
	publicID := fmt.Sprintf("rydex/avatars/user_%d", userID)

	uploadResult, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       publicID,
		Folder:         "rydex/avatars",
		Overwrite:      api.Bool(true),
		UniqueFilename: api.Bool(false),
		Transformation: "c_fill,g_face,w_400,h_400,q_auto,f_auto", // crop to face, 400x400
		Tags:           api.CldAPIArray{"rydex", "avatar"},
	})
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	// Persist public_id and secure_url to user record
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"avatar_public_id": uploadResult.PublicID,
		"avatar_url":       uploadResult.SecureURL,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to save avatar: %w", err)
	}

	return &UploadResult{
		PublicID:  uploadResult.PublicID,
		SecureURL: uploadResult.SecureURL,
		Width:     uploadResult.Width,
		Height:    uploadResult.Height,
		Format:    uploadResult.Format,
		Bytes:     uploadResult.Bytes,
	}, nil
}

// DeleteAvatar removes the user's avatar from Cloudinary and clears DB fields.
func (s *CloudinaryService) DeleteAvatar(ctx context.Context, userID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if user.AvatarPublicID == "" {
		return errors.New("no avatar to delete")
	}

	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: user.AvatarPublicID,
	})
	if err != nil {
		return fmt.Errorf("cloudinary delete failed: %w", err)
	}

	return s.db.Model(&user).Updates(map[string]interface{}{
		"avatar_public_id": "",
		"avatar_url":       "",
	}).Error
}

// GetAvatarURL returns a Cloudinary transformation URL for a given public_id.
// This lets you request different sizes without storing multiple URLs.
func (s *CloudinaryService) GetAvatarURL(publicID string, width, height int) string {
	if publicID == "" {
		return ""
	}
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	return fmt.Sprintf(
		"https://res.cloudinary.com/%s/image/upload/c_fill,g_face,w_%d,h_%d,q_auto,f_auto/%s",
		cloudName, width, height, publicID,
	)
}

// GetProfile returns the user with a transformed avatar URL
func (s *CloudinaryService) GetProfile(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("RiderProfile").Preload("ClientProfile").
		First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func validateImageType(filename string) error {
	lower := strings.ToLower(filename)
	allowed := []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}
	for _, ext := range allowed {
		if strings.HasSuffix(lower, ext) {
			return nil
		}
	}
	return fmt.Errorf("unsupported file type: allowed types are jpg, jpeg, png, webp, gif")
}
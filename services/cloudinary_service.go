package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

// CloudinaryService handles image uploads to Cloudinary
type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

var cloudinaryInstance *CloudinaryService

// UploadResult holds the result of an upload operation
type UploadResult struct {
	SecureURL string
	PublicID  string
	Width     int
	Height    int
	Format    string
	Bytes     int
}

// InitCloudinary initializes Cloudinary service
func InitCloudinary(cloudName, apiKey, apiSecret string) error {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	cloudinaryInstance = &CloudinaryService{
		cld: cld,
	}

	return nil
}

// GetCloudinaryService returns the Cloudinary service instance
func GetCloudinaryService() *CloudinaryService {
	return cloudinaryInstance
}

// UploadImage uploads an image to Cloudinary (package level function)
func UploadImage(file multipart.File, filename, folder string) (*UploadResult, error) {
	if cloudinaryInstance == nil {
		return nil, fmt.Errorf("cloudinary not initialized")
	}

	ctx := context.Background()

	// Generate unique filename
	uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), uuid.New().String()[:8])
	publicID := uniqueName

	// Upload to Cloudinary
	uploadResp, err := cloudinaryInstance.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:     publicID,
		Folder:       folder,
		ResourceType: "image",
	})

	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return &UploadResult{
		SecureURL: uploadResp.SecureURL,
		PublicID:  uploadResp.PublicID,
		Width:     uploadResp.Width,
		Height:    uploadResp.Height,
		Format:    uploadResp.Format,
		Bytes:     uploadResp.Bytes,
	}, nil
}

// UploadFromURL uploads an image from URL to Cloudinary (package level function)
func UploadFromURL(imageURL, folder string) (*UploadResult, error) {
	if cloudinaryInstance == nil {
		return nil, fmt.Errorf("cloudinary not initialized")
	}

	ctx := context.Background()

	// Generate unique public ID
	uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), uuid.New().String()[:8])
	publicID := uniqueName

	uploadResp, err := cloudinaryInstance.cld.Upload.Upload(ctx, imageURL, uploader.UploadParams{
		PublicID:     publicID,
		Folder:       folder,
		ResourceType: "image",
	})

	if err != nil {
		return nil, fmt.Errorf("cloudinary upload from URL failed: %w", err)
	}

	return &UploadResult{
		SecureURL: uploadResp.SecureURL,
		PublicID:  uploadResp.PublicID,
		Width:     uploadResp.Width,
		Height:    uploadResp.Height,
		Format:    uploadResp.Format,
		Bytes:     uploadResp.Bytes,
	}, nil
}

// DeleteImage deletes an image from Cloudinary by public ID
func DeleteImage(publicID string) error {
	if cloudinaryInstance == nil {
		return fmt.Errorf("cloudinary not initialized")
	}

	ctx := context.Background()

	_, err := cloudinaryInstance.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})

	if err != nil {
		return fmt.Errorf("cloudinary delete failed: %w", err)
	}

	return nil
}

// DeleteImageByURL deletes an image from Cloudinary by URL
func DeleteImageByURL(imageURL string) error {
	// Check if it's a Cloudinary URL
	if !strings.Contains(imageURL, "cloudinary.com") {
		// Not a cloudinary URL, skip
		return nil
	}

	// Extract public ID from URL
	publicID := extractPublicIDFromURL(imageURL)
	if publicID == "" {
		return fmt.Errorf("invalid cloudinary URL")
	}

	return DeleteImage(publicID)
}

// extractPublicIDFromURL extracts public ID from Cloudinary URL
func extractPublicIDFromURL(imageURL string) string {
	if !strings.Contains(imageURL, "cloudinary.com") {
		return ""
	}

	// Split by /upload/
	parts := strings.Split(imageURL, "/upload/")
	if len(parts) != 2 {
		return ""
	}

	// Get path after version (v123456/)
	path := parts[1]
	// Remove version prefix if exists
	if strings.HasPrefix(path, "v") {
		vParts := strings.SplitN(path, "/", 2)
		if len(vParts) == 2 {
			path = vParts[1]
		}
	}

	// Remove file extension
	publicID := strings.TrimSuffix(path, filepath.Ext(path))

	return publicID
}

// TestUpload tests Cloudinary upload with a sample reader
func TestUpload(reader io.Reader, filename string) (*UploadResult, error) {
	if cloudinaryInstance == nil {
		return nil, fmt.Errorf("cloudinary not initialized")
	}

	ctx := context.Background()

	uploadResp, err := cloudinaryInstance.cld.Upload.Upload(ctx, reader, uploader.UploadParams{
		PublicID:     "test_" + filename,
		Folder:       "qr-restaurant/test",
		ResourceType: "image",
	})

	if err != nil {
		return nil, fmt.Errorf("test upload failed: %w", err)
	}

	return &UploadResult{
		SecureURL: uploadResp.SecureURL,
		PublicID:  uploadResp.PublicID,
		Width:     uploadResp.Width,
		Height:    uploadResp.Height,
		Format:    uploadResp.Format,
		Bytes:     uploadResp.Bytes,
	}, nil
}

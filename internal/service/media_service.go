package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/afdhali/GolangBlogpostServer/internal/dto"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/pkg/image"
	"github.com/afdhali/GolangBlogpostServer/pkg/storage"
	"github.com/afdhali/GolangBlogpostServer/pkg/validator"
	"github.com/google/uuid"
)

type MediaService interface {
	Upload(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader, req *dto.UploadMediaRequest) (*dto.MediaResponse, error)
	// GetAll(ctx context.Context, params *dto.MediaQueryParams) ([]*dto.MediaListResponse, int64, error)
	GetAll(ctx context.Context, params *dto.MediaQueryParams, currentUser *entity.User) ([]*dto.MediaListResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.MediaResponse, error)
	GetByPostID(ctx context.Context, postID uuid.UUID) ([]*dto.MediaResponse, error)
	GetFeaturedByPostID(ctx context.Context, postID uuid.UUID) (*dto.MediaResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateMediaRequest, user *entity.User) (*dto.MediaResponse, error)
	Delete(ctx context.Context, id uuid.UUID, user *entity.User) error
}

type mediaService struct {
	mediaRepo      repository.MediaRepository
	postRepo       repository.PostRepository
	storage        storage.Storage
	imageValidator *image.Validator
	imageProcessor *image.Processor
	validator      *validator.CustomValidator
}

func NewMediaService(
	mediaRepo repository.MediaRepository,
	postRepo repository.PostRepository,
	storage storage.Storage,
	imageValidator *image.Validator,
	imageProcessor *image.Processor,
	validator *validator.CustomValidator,
) MediaService {
	return &mediaService{
		mediaRepo:      mediaRepo,
		postRepo:       postRepo,
		storage:        storage,
		imageValidator: imageValidator,
		imageProcessor: imageProcessor,
		validator:      validator,
	}
}

func (s *mediaService) Upload(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader, req *dto.UploadMediaRequest) (*dto.MediaResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Validate image
	if err := s.imageValidator.Validate(header); err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	// Check if post exists (if post_id provided)
	if req.PostID != nil {
		_, err := s.postRepo.FindByID(ctx, *req.PostID)
		if err != nil {
			return nil, errors.New("post not found")
		}
	}

	// Process image (compress & resize)
	processedFile, err := s.imageProcessor.Process(file, header)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Convert io.Reader to multipart.File
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, processedFile); err != nil {
		return nil, fmt.Errorf("failed to read processed image: %w", err)
	}

	// Create a new multipart.File from buffer
	processedMultipart := &bytesFileMedia{
		Reader: bytes.NewReader(buf.Bytes()),
		size:   int64(buf.Len()),
	}

	// Save to storage
	fileInfo, err := s.storage.Save(ctx, processedMultipart, header, "posts")
	if err != nil {
		return nil, fmt.Errorf("failed to save media: %w", err)
	}

	// Create media entity
	media := &entity.Media{
		Filename:     fileInfo.Filename,
		OriginalName: fileInfo.OriginalName,
		MimeType:     fileInfo.MimeType,
		Path:         fileInfo.Path,
		URL:          fileInfo.URL,
		Size:         fileInfo.Size,
		AltText:      req.AltText,
		Description:  req.Description,
		MediaType:    entity.MediaTypeImage,
		PostID:       req.PostID,
		UserID:       userID,
		IsFeatured:   req.IsFeatured,
	}

	// Save media to database
	if err := s.mediaRepo.Create(ctx, media); err != nil {
		// Rollback: delete uploaded file
		s.storage.Delete(ctx, fileInfo.Path)
		return nil, fmt.Errorf("failed to save media to database: %w", err)
	}

	// Reload with relations
	media, _ = s.mediaRepo.FindByID(ctx, media.ID)

	return dto.ToMediaResponse(media), nil
}

// func (s *mediaService) GetAll(ctx context.Context, params *dto.MediaQueryParams) ([]*dto.MediaListResponse, int64, error) {
// 	// Validate params
// 	if err := s.validator.Validate(params); err != nil {
// 		return nil, 0, fmt.Errorf("validation error: %w", err)
// 	}

// 	// Default pagination
// 	if params.Page < 1 {
// 		params.Page = 1
// 	}
// 	if params.Limit < 1 {
// 		params.Limit = 10
// 	}

// 	// Get medias
// 	medias, total, err := s.mediaRepo.FindAll(ctx, params.Page, params.Limit, params.MediaType, params.PostID, params.UserID, params.IsFeatured, params.SortBy, params.SortOrder)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("failed to get medias: %w", err)
// 	}

// 	// Convert to response
// 	responses := dto.ToMediaListResponses(medias)

// 	return responses, total, nil
// }

func (s *mediaService) GetAll(
    ctx context.Context,
    params *dto.MediaQueryParams,
    currentUser *entity.User, // bisa nil untuk public call
) ([]*dto.MediaListResponse, int64, error) {
    // Validate params
    if err := s.validator.Validate(params); err != nil {
        return nil, 0, fmt.Errorf("validation error: %w", err)
    }

    // Default pagination
    if params.Page < 1 {
        params.Page = 1
    }
    if params.Limit < 1 {
        params.Limit = 10
    }
    if params.Limit > 100 {
        params.Limit = 100
    }

    // ---------- LOGIKA FILTER USER_ID BERDASARKAN ROLE ----------
    var effectiveUserID *uuid.UUID

    if params.UserID != nil {
        // Ada request filter berdasarkan user_id
        if currentUser == nil {
            // Public call, tetap pakai filter user_id yang diminta
            effectiveUserID = params.UserID
        } else if currentUser.IsAdmin() {
            // Admin/SuperAdmin boleh melihat semua, jadi abaikan filter user_id yang dikirim
            // (atau tetap pakai jika memang ingin filter spesifik user tertentu)
            // Di sini kita pilih: admin tetap bisa filter per user, jadi pakai yang dikirim
            effectiveUserID = params.UserID
        } else {
            // User biasa → wajib hanya boleh lihat milik sendiri
            if *params.UserID != currentUser.ID {
                return nil, 0, errors.New("you can only view your own media")
            }
            effectiveUserID = params.UserID // = currentUser.ID
        }
    } else if currentUser != nil && !currentUser.IsAdmin() {
        // Tidak ada user_id di query, tapi user login adalah user biasa
        // → otomatis filter hanya media miliknya
        effectiveUserID = &currentUser.ID
    }
    // jika currentUser == nil (public) atau admin dan tidak ada user_id → effectiveUserID tetap nil (lihat semua)

    // ---------- PANGGIL REPOSITORY ----------
    medias, total, err := s.mediaRepo.FindAll(
        ctx,
        params.Page,
        params.Limit,
        params.MediaType,
        params.PostID,
        effectiveUserID,        // ← ini yang sudah di-adjust
        params.IsFeatured,
        params.SortBy,
        params.SortOrder,
    )
    if err != nil {
        return nil, 0, fmt.Errorf("failed to get medias: %w", err)
    }

    // Konversi ke response (User sudah di-preload di repository)
    responses := dto.ToMediaListResponses(medias)

    return responses, total, nil
}

func (s *mediaService) GetByID(ctx context.Context, id uuid.UUID) (*dto.MediaResponse, error) {
	media, err := s.mediaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("media not found")
	}

	return dto.ToMediaResponse(media), nil
}

func (s *mediaService) GetByPostID(ctx context.Context, postID uuid.UUID) ([]*dto.MediaResponse, error) {
	// Check if post exists
	_, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	medias, err := s.mediaRepo.FindByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get medias: %w", err)
	}

	return dto.ToMediaResponses(medias), nil
}

func (s *mediaService) GetFeaturedByPostID(ctx context.Context, postID uuid.UUID) (*dto.MediaResponse, error) {
	// Check if post exists
	_, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	media, err := s.mediaRepo.FindFeaturedByPostID(ctx, postID)
	if err != nil {
		return nil, errors.New("featured media not found")
	}

	return dto.ToMediaResponse(media), nil
}

func (s *mediaService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateMediaRequest, user *entity.User) (*dto.MediaResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get media
	media, err := s.mediaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("media not found")
	}

	// Permission check
	// ✅ Owner (user yang upload) dapat update media milik sendiri
	// ✅ Admin dapat update media apapun
	// ❌ User lain tidak dapat update
	if !s.canManageMedia(user, media) {
		return nil, errors.New("you don't have permission to update this media")
	}

	// Check post if changed
	if req.PostID != nil && (media.PostID == nil || *req.PostID != *media.PostID) {
		_, err := s.postRepo.FindByID(ctx, *req.PostID)
		if err != nil {
			return nil, errors.New("post not found")
		}
		media.PostID = req.PostID
	}

	// Update fields
	if req.AltText != "" {
		media.AltText = req.AltText
	}

	if req.Description != "" {
		media.Description = req.Description
	}

	if req.IsFeatured != nil {
		media.IsFeatured = *req.IsFeatured
	}

	if err := s.mediaRepo.Update(ctx, media); err != nil {
		return nil, fmt.Errorf("failed to update media: %w", err)
	}

	// Reload with relations
	media, _ = s.mediaRepo.FindByID(ctx, media.ID)

	return dto.ToMediaResponse(media), nil
}

func (s *mediaService) Delete(ctx context.Context, id uuid.UUID, user *entity.User) error {
	media, err := s.mediaRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("media not found")
	}

	// Permission check
	// ✅ Owner (user yang upload) dapat delete media milik sendiri
	// ✅ Admin dapat delete media apapun
	// ❌ User lain tidak dapat delete
	if !s.canManageMedia(user, media) {
		return errors.New("you don't have permission to delete this media")
	}

	// Delete file from storage
	if err := s.storage.Delete(ctx, media.Path); err != nil {
		return fmt.Errorf("failed to delete media file: %w", err)
	}

	// Delete from database
	return s.mediaRepo.Delete(ctx, id)
}

// 👇 HELPER METHOD - Permission check
// canManageMedia checks if user can manage (update/delete) the media
// - Owner (user yang upload) dapat manage media miliknya sendiri
// - Admin dapat manage media apapun
// - User lain tidak dapat manage
func (s *mediaService) canManageMedia(user *entity.User, media *entity.Media) bool {
	// Admin (Super Admin dan Admin) dapat manage media apapun
	if user.IsAdmin() {
		return true
	}

	// Owner dapat manage media milik sendiri
	return user.ID == media.UserID
}

// Helper type
type bytesFileMedia struct {
	*bytes.Reader
	size int64
}

func (b *bytesFileMedia) Close() error {
	return nil
}
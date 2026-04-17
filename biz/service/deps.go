package service

import (
	"context"
	"mime/multipart"
	"video-platform/biz/dal/model"
	"video-platform/biz/repository"
	"video-platform/pkg/auth"
	"video-platform/pkg/upload"
)

type userRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error)
	UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error
}

type relationRepository interface {
	GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error)
	FollowUser(ctx context.Context, fromUserID, toUserID uint) error
	UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error)
	ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error)
	ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error)
	ListFriends(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error)
}

type videoRepository interface {
	CreateVideo(ctx context.Context, video *model.Video) error
	GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error)
	ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error)
	GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error)
	ListVideosByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error)
	SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, offset, limit int) ([]model.Video, error)
	ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) (*repository.VideoCommentListResult, error)
	ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]repository.UserProfile, error)
	ListHotVideos(ctx context.Context, offset, limit int) ([]model.Video, error)
}

type authProvider interface {
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword, password string) error
	GenerateTokenPair(userID uint) (accessToken, refreshToken string, err error)
	RefreshTokens(refreshToken string) (newAccessToken, newRefreshToken string, err error)
}

type uploadProvider interface {
	PrepareAvatar(userID uint, originalFilename string) (savePath, avatarURL string, err error)
	SaveFile(file *multipart.FileHeader, savePath string) error
	RemoveAvatar(avatarURL string) error
	PrepareVideo(userID uint, originalFilename string) (savePath, videoURL string, err error)
	PrepareVideoCover(userID uint, originalFilename string) (savePath, coverURL string, err error)
	RemoveVideo(videoURL string) error
	RemoveVideoCover(coverURL string) error
}

type defaultUserRepository struct{}

func (defaultUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	return repository.CreateUser(ctx, user)
}

func (defaultUserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return repository.GetUserByUsername(ctx, username)
}

func (defaultUserRepository) GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error) {
	return repository.GetUserByID(ctx, userID)
}

func (defaultUserRepository) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	return repository.UpdateUserAvatar(ctx, userID, avatarURL)
}

type defaultRelationRepository struct{}

func (defaultRelationRepository) GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error) {
	return repository.GetUserByID(ctx, userID)
}

func (defaultRelationRepository) FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return repository.FollowUser(ctx, fromUserID, toUserID)
}

func (defaultRelationRepository) UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	return repository.UnfollowUser(ctx, fromUserID, toUserID)
}

func (defaultRelationRepository) ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error) {
	return repository.ListFollowings(ctx, userID, offset, limit)
}

func (defaultRelationRepository) ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error) {
	return repository.ListFollowers(ctx, userID, offset, limit)
}

func (defaultRelationRepository) ListFriends(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error) {
	return repository.ListFriends(ctx, userID, offset, limit)
}

type defaultVideoRepository struct{}

func (defaultVideoRepository) CreateVideo(ctx context.Context, video *model.Video) error {
	return repository.CreateVideo(ctx, video)
}

func (defaultVideoRepository) GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error) {
	return repository.GetUserByID(ctx, userID)
}

func (defaultVideoRepository) ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	return repository.ListUserIDsByUsername(ctx, username)
}

func (defaultVideoRepository) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return repository.GetVideoByID(ctx, videoID)
}

func (defaultVideoRepository) ListVideosByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error) {
	return repository.ListVideosByUserID(ctx, userID, offset, limit)
}

func (defaultVideoRepository) SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, offset, limit int) ([]model.Video, error) {
	return repository.SearchVideos(ctx, keywords, userIDs, fromDate, toDate, sortBy, offset, limit)
}

func (defaultVideoRepository) ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) (*repository.VideoCommentListResult, error) {
	return repository.ListVideoComments(ctx, videoID, cursor, limit)
}

func (defaultVideoRepository) ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]repository.UserProfile, error) {
	return repository.ListUserSnapshotsByIDs(ctx, userIDs)
}

func (defaultVideoRepository) ListHotVideos(ctx context.Context, offset, limit int) ([]model.Video, error) {
	return repository.ListHotVideos(ctx, offset, limit)
}

type defaultAuthProvider struct{}

func (defaultAuthProvider) HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

func (defaultAuthProvider) CheckPassword(hashedPassword, password string) error {
	return auth.CheckPassword(hashedPassword, password)
}

func (defaultAuthProvider) GenerateTokenPair(userID uint) (accessToken, refreshToken string, err error) {
	return auth.GenerateTokenPair(userID)
}

func (defaultAuthProvider) RefreshTokens(refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	return auth.RefreshTokens(refreshToken)
}

type defaultUploadProvider struct{}

func (defaultUploadProvider) PrepareAvatar(userID uint, originalFilename string) (savePath, avatarURL string, err error) {
	return upload.PrepareAvatar(userID, originalFilename)
}

func (defaultUploadProvider) SaveFile(file *multipart.FileHeader, savePath string) error {
	return upload.SaveFile(file, savePath)
}

func (defaultUploadProvider) RemoveAvatar(avatarURL string) error {
	return upload.RemoveAvatar(avatarURL)
}

func (defaultUploadProvider) PrepareVideo(userID uint, originalFilename string) (savePath, videoURL string, err error) {
	return upload.PrepareVideo(userID, originalFilename)
}

func (defaultUploadProvider) PrepareVideoCover(userID uint, originalFilename string) (savePath, coverURL string, err error) {
	return upload.PrepareVideoCover(userID, originalFilename)
}

func (defaultUploadProvider) RemoveVideo(videoURL string) error {
	return upload.RemoveVideo(videoURL)
}

func (defaultUploadProvider) RemoveVideoCover(coverURL string) error {
	return upload.RemoveVideoCover(coverURL)
}

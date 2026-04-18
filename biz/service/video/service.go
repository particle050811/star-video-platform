package video

import (
	"context"
	"errors"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
	"video-platform/biz/dal/model"
	v1 "video-platform/biz/model/video"
	commentrepo "video-platform/biz/repository/comment"
	userrepo "video-platform/biz/repository/user"
	videorepo "video-platform/biz/repository/video"
	usersvc "video-platform/biz/service/user"
	"video-platform/pkg/pagination"
	"video-platform/pkg/parser"
	"video-platform/pkg/upload"

	"gorm.io/gorm"
)

type videoRepository interface {
	CreateVideo(ctx context.Context, video *model.Video) error
	GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error)
	ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error)
	GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error)
	ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) (*videorepo.VideoListResult, error)
	SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, cursor uint, limit int) (*videorepo.VideoListResult, error)
	ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) (*commentrepo.VideoCommentListResult, error)
	ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error)
	ListHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) (*videorepo.VideoListResult, error)
}

type videoUploadProvider interface {
	SaveFile(file *multipart.FileHeader, savePath string) error
	PrepareVideo(userID uint, originalFilename string) (savePath, videoURL string, err error)
	PrepareVideoCover(userID uint, originalFilename string) (savePath, coverURL string, err error)
	RemoveVideo(videoURL string) error
	RemoveVideoCover(coverURL string) error
}

type defaultVideoRepository struct{}

func (defaultVideoRepository) CreateVideo(ctx context.Context, video *model.Video) error {
	return videorepo.CreateVideo(ctx, video)
}

func (defaultVideoRepository) GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
	return userrepo.GetUserByID(ctx, userID)
}

func (defaultVideoRepository) ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	return userrepo.ListUserIDsByUsername(ctx, username)
}

func (defaultVideoRepository) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return videorepo.GetVideoByID(ctx, videoID)
}

func (defaultVideoRepository) ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) (*videorepo.VideoListResult, error) {
	return videorepo.ListVideosByUserID(ctx, userID, cursor, limit)
}

func (defaultVideoRepository) SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, cursor uint, limit int) (*videorepo.VideoListResult, error) {
	return videorepo.SearchVideos(ctx, keywords, userIDs, fromDate, toDate, sortBy, cursor, limit)
}

func (defaultVideoRepository) ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) (*commentrepo.VideoCommentListResult, error) {
	return commentrepo.ListVideoComments(ctx, videoID, cursor, limit)
}

func (defaultVideoRepository) ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
	return userrepo.ListUserSnapshotsByIDs(ctx, userIDs)
}

func (defaultVideoRepository) ListHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) (*videorepo.VideoListResult, error) {
	return videorepo.ListHotVideos(ctx, cursor, limit)
}

type defaultVideoUploadProvider struct{}

func (defaultVideoUploadProvider) SaveFile(file *multipart.FileHeader, savePath string) error {
	return upload.SaveFile(file, savePath)
}

func (defaultVideoUploadProvider) PrepareVideo(userID uint, originalFilename string) (savePath, videoURL string, err error) {
	return upload.PrepareVideo(userID, originalFilename)
}

func (defaultVideoUploadProvider) PrepareVideoCover(userID uint, originalFilename string) (savePath, coverURL string, err error) {
	return upload.PrepareVideoCover(userID, originalFilename)
}

func (defaultVideoUploadProvider) RemoveVideo(videoURL string) error {
	return upload.RemoveVideo(videoURL)
}

func (defaultVideoUploadProvider) RemoveVideoCover(coverURL string) error {
	return upload.RemoveVideoCover(coverURL)
}

type videoService struct {
	repo   videoRepository
	upload videoUploadProvider
}

var Video = videoService{
	repo:   defaultVideoRepository{},
	upload: defaultVideoUploadProvider{},
}

func (s videoService) PublishVideo(ctx context.Context, userID uint, title, description string, videoFile, coverFile *multipart.FileHeader) (err error) {
	if title == "" {
		return ErrVideoTitleRequired
	}
	if videoFile == nil {
		return ErrVideoFileRequired
	}

	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usersvc.ErrUserNotFound
		}
		return err
	}

	var videoURL string
	var coverURL string
	defer func() {
		if err != nil {
			_ = s.upload.RemoveVideo(videoURL)
			_ = s.upload.RemoveVideoCover(coverURL)
		}
	}()

	var videoPath string
	videoPath, videoURL, err = s.upload.PrepareVideo(userID, videoFile.Filename)
	if err != nil {
		if errors.Is(err, upload.ErrUnsupportedVideoExt) {
			return ErrUnsupportedVideoExt
		}
		return err
	}

	if err := s.upload.SaveFile(videoFile, videoPath); err != nil {
		return err
	}

	if coverFile != nil {
		coverPath, savedCoverURL, coverErr := s.upload.PrepareVideoCover(userID, coverFile.Filename)
		if coverErr != nil {
			if errors.Is(coverErr, upload.ErrUnsupportedVideoCoverExt) {
				return ErrUnsupportedVideoCoverExt
			}
			return coverErr
		}
		coverURL = savedCoverURL
		if err := s.upload.SaveFile(coverFile, coverPath); err != nil {
			return err
		}
	}

	if err := s.repo.CreateVideo(ctx, &model.Video{
		UserID:      userID,
		VideoURL:    videoURL,
		CoverURL:    coverURL,
		Title:       title,
		Description: description,
	}); err != nil {
		return err
	}

	return nil
}

func (s videoService) ListPublishedVideos(ctx context.Context, userID uint, cursor uint, limit int32) (*v1.VideoList, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usersvc.ErrUserNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListVideosByUserID(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return buildVideoList(result), nil
}

func (s videoService) SearchVideos(ctx context.Context, req *v1.SearchVideosRequest, cursor uint) (*v1.VideoList, error) {
	var userIDs []uint
	if username := strings.TrimSpace(req.Username); username != "" {
		foundUserIDs, err := s.repo.ListUserIDsByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		userIDs = foundUserIDs
	}

	result, err := s.repo.SearchVideos(ctx, req.Keywords, userIDs, req.FromDate, req.ToDate, req.SortBy, cursor, pagination.NormalizeLimit(req.Limit))
	if err != nil {
		return nil, err
	}

	return buildVideoList(result), nil
}

func (s videoService) ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int32) (*v1.VideoCommentList, error) {
	if _, err := s.repo.GetVideoByID(ctx, videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListVideoComments(ctx, videoID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	if result == nil {
		return buildVideoCommentList(nil, nil), nil
	}

	commentUserIDs := make([]uint, 0, len(result.Items))
	for _, item := range result.Items {
		commentUserIDs = append(commentUserIDs, item.UserID)
	}

	users, err := s.repo.ListUserSnapshotsByIDs(ctx, commentUserIDs)
	if err != nil {
		return nil, err
	}

	return buildVideoCommentList(result, users), nil
}

func (s videoService) GetHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int32) (*v1.VideoList, error) {
	result, err := s.repo.ListHotVideos(ctx, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return buildHotVideoList(result), nil
}

func buildVideoCommentList(result *commentrepo.VideoCommentListResult, users []userrepo.UserProfile) *v1.VideoCommentList {
	items := make([]*v1.VideoComment, 0)
	if result == nil {
		return &v1.VideoCommentList{
			Items: items,
		}
	}

	userMap := make(map[uint]userrepo.UserProfile, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	items = make([]*v1.VideoComment, 0, len(result.Items))
	for _, item := range result.Items {
		user := userMap[item.UserID]
		items = append(items, &v1.VideoComment{
			Id:        strconv.FormatUint(uint64(item.ID), 10),
			UserId:    strconv.FormatUint(uint64(item.UserID), 10),
			Username:  user.Username,
			AvatarUrl: user.AvatarURL,
			Content:   item.Content,
			LikeCount: item.LikeCount,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}

	nextCursor := ""
	if result.HasMore {
		nextCursor = strconv.FormatUint(uint64(result.NextCursor), 10)
	}

	return &v1.VideoCommentList{
		Items:      items,
		Total:      result.Total,
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	}
}

func buildVideoList(result *videorepo.VideoListResult) *v1.VideoList {
	items := make([]*v1.Video, 0)
	if result == nil {
		return &v1.VideoList{
			Items: items,
		}
	}

	items = make([]*v1.Video, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, buildVideo(item))
	}

	nextCursor := ""
	if result.HasMore {
		nextCursor = strconv.FormatUint(uint64(result.NextCursor), 10)
	}

	return &v1.VideoList{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	}
}

func BuildVideoList(result *videorepo.VideoListResult) *v1.VideoList {
	return buildVideoList(result)
}

func buildHotVideoList(result *videorepo.VideoListResult) *v1.VideoList {
	data := buildVideoList(result)
	if result != nil && result.HasMore {
		data.NextCursor = result.NextCursorToken
	}
	return data
}

func buildVideo(video model.Video) *v1.Video {
	return &v1.Video{
		Id:           strconv.FormatUint(uint64(video.ID), 10),
		UserId:       strconv.FormatUint(uint64(video.UserID), 10),
		VideoUrl:     video.VideoURL,
		CoverUrl:     video.CoverURL,
		Title:        video.Title,
		Description:  video.Description,
		VisitCount:   video.VisitCount,
		LikeCount:    video.LikeCount,
		CommentCount: video.CommentCount,
		CreatedAt:    video.CreatedAt.Format(time.RFC3339),
	}
}

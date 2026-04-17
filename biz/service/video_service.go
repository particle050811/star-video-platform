package service

import (
	"context"
	"errors"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
	"video-platform/biz/dal/model"
	v1 "video-platform/biz/model/video"
	"video-platform/biz/repository"
	"video-platform/pkg/pagination"
	"video-platform/pkg/upload"

	"gorm.io/gorm"
)

type videoService struct {
	repo   videoRepository
	upload uploadProvider
}

var Video = videoService{
	repo:   defaultVideoRepository{},
	upload: defaultUploadProvider{},
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
			return ErrUserNotFound
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

func (s videoService) ListPublishedVideos(ctx context.Context, userID uint, pageNum, pageSize int32) (*v1.VideoList, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	offset, limit := pagination.Normalize(pageNum, pageSize)
	videos, err := s.repo.ListVideosByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	return buildVideoList(videos), nil
}

func (s videoService) SearchVideos(ctx context.Context, req *v1.SearchVideosRequest) (*v1.VideoList, error) {
	offset, limit := pagination.Normalize(req.PageNum, req.PageSize)

	var userIDs []uint
	if username := strings.TrimSpace(req.Username); username != "" {
		foundUserIDs, err := s.repo.ListUserIDsByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		userIDs = foundUserIDs
	}

	videos, err := s.repo.SearchVideos(ctx, req.Keywords, userIDs, req.FromDate, req.ToDate, req.SortBy, offset, limit)
	if err != nil {
		return nil, err
	}

	return buildVideoList(videos), nil
}

func (s videoService) ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int32) (*v1.VideoCommentList, error) {
	if _, err := s.repo.GetVideoByID(ctx, videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	if limit <= 0 {
		limit = pagination.DefaultPageSize
	}
	if limit > pagination.MaxPageSize {
		limit = pagination.MaxPageSize
	}

	result, err := s.repo.ListVideoComments(ctx, videoID, cursor, int(limit))
	if err != nil {
		return nil, err
	}

	commentUserIDs := make([]uint, 0, len(result.Items))
	for _, item := range result.Items {
		commentUserIDs = append(commentUserIDs, item.UserID)
	}

	users, err := s.repo.ListUserSnapshotsByIDs(ctx, commentUserIDs)
	if err != nil {
		return nil, err
	}

	userMap := make(map[uint]repository.UserProfile, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	items := make([]*v1.VideoComment, 0, len(result.Items))
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
	}, nil
}

func (s videoService) GetHotVideos(ctx context.Context, pageNum, pageSize int32) (*v1.VideoList, error) {
	offset, limit := pagination.Normalize(pageNum, pageSize)
	videos, err := s.repo.ListHotVideos(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	return buildVideoList(videos), nil
}

func buildVideoList(videos []model.Video) *v1.VideoList {
	items := make([]*v1.Video, 0, len(videos))
	for _, item := range videos {
		items = append(items, buildVideo(item))
	}
	return &v1.VideoList{Items: items}
}

func buildVideo(video model.Video) *v1.Video {
	result := &v1.Video{
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
		UpdatedAt:    video.UpdatedAt.Format(time.RFC3339),
	}

	if video.DeletedAt.Valid {
		result.DeletedAt = video.DeletedAt.Time.Format(time.RFC3339)
	}

	return result
}

package dto

import (
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
)

type CommentResponse struct {
    ID        uuid.UUID        `json:"id"`
    Content   string           `json:"content"`
    PostID    uuid.UUID        `json:"post_id"`
    UserID    uuid.UUID        `json:"user_id"`
    Author    *CommentAuthor   `json:"author"`
    ParentID  *uuid.UUID       `json:"parent_id,omitempty"`
    Replies   []*CommentReply  `json:"replies,omitempty"`
    CreatedAt time.Time        `json:"created_at"`
    UpdatedAt time.Time        `json:"updated_at"`
}

type CommentAuthor struct {
    ID       uuid.UUID `json:"id"`
    Username string    `json:"username"`
    FullName string    `json:"full_name"`
    Avatar   string    `json:"avatar,omitempty"`
}

type CommentReply struct {
    ID        uuid.UUID      `json:"id"`
    Content   string         `json:"content"`
    Author    *CommentAuthor `json:"author"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}

// Converter functions
func ToCommentResponse(comment *entity.Comment) *CommentResponse {
    response := &CommentResponse{
        ID:        comment.ID,
        Content:   comment.Content,
        PostID:    comment.PostID,
        UserID:    comment.UserID,
        ParentID:  comment.ParentID,
        CreatedAt: comment.CreatedAt,
        UpdatedAt: comment.UpdatedAt,
    }

    // Add author if exists
    if comment.User != nil {
        response.Author = &CommentAuthor{
            ID:       comment.User.ID,
            Username: comment.User.Username,
            FullName: comment.User.FullName,
            Avatar:   comment.User.Avatar,
        }
    }

    // Add replies if exists
    if len(comment.Replies) > 0 {
        response.Replies = make([]*CommentReply, len(comment.Replies))
        for i, reply := range comment.Replies {
            response.Replies[i] = ToCommentReply(&reply)
        }
    }

    return response
}

func ToCommentReply(comment *entity.Comment) *CommentReply {
    reply := &CommentReply{
        ID:        comment.ID,
        Content:   comment.Content,
        CreatedAt: comment.CreatedAt,
        UpdatedAt: comment.UpdatedAt,
    }

    if comment.User != nil {
        reply.Author = &CommentAuthor{
            ID:       comment.User.ID,
            Username: comment.User.Username,
            FullName: comment.User.FullName,
            Avatar:   comment.User.Avatar,
        }
    }

    return reply
}

func ToCommentResponses(comments []*entity.Comment) []*CommentResponse {
    responses := make([]*CommentResponse, len(comments))
    for i, comment := range comments {
        responses[i] = ToCommentResponse(comment)
    }
    return responses
}
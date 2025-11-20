package dto

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
    Page       int   `json:"page"`
    Limit      int   `json:"limit"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
}

// MessageResponse for simple message responses
type MessageResponse struct {
    Message string `json:"message"`
}

// ErrorDetail for validation errors
type ErrorDetail struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// HealthCheckResponse for health check endpoint
type HealthCheckResponse struct {
    Status    string `json:"status"`
    Timestamp string `json:"timestamp"`
    Version   string `json:"version"`
}

// Helper functions
func NewPaginationMeta(page, limit int, total int64) *PaginationMeta {
    totalPages := int(total) / limit
    if int(total)%limit > 0 {
        totalPages++
    }

    return &PaginationMeta{
        Page:       page,
        Limit:      limit,
        Total:      total,
        TotalPages: totalPages,
    }
}

func NewMessageResponse(message string) *MessageResponse {
    return &MessageResponse{
        Message: message,
    }
}
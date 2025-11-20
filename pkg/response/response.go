package response

import "github.com/gin-gonic/gin"

type Response struct {
    Code   int         `json:"code"`
    Status string      `json:"status"`
    Data   interface{} `json:"data,omitempty"`
}

type PaginationResponse struct {
    Page       int         `json:"page"`
    Limit      int         `json:"limit"`
    TotalData  int64       `json:"total_data"`
    TotalPages int         `json:"total_pages"`
    Data       interface{} `json:"data"`
}

func Success(c *gin.Context, code int, data interface{}) {
    status := getStatusText(code)
    c.JSON(code, Response{
        Code:   code,
        Status: status,
        Data:   data,
    })
}

func Error(c *gin.Context, code int, message string, errors interface{}) {
    status := getStatusText(code)
    data := gin.H{"message": message}
    
    if errors != nil {
        data["errors"] = errors
    }

    c.JSON(code, Response{
        Code:   code,
        Status: status,
        Data:   data,
    })
}

func SuccessWithPagination(c *gin.Context, code int, page, limit int, total int64, data interface{}) {
    status := getStatusText(code)
    totalPages := int(total) / limit
    if int(total)%limit > 0 {
        totalPages++
    }

    c.JSON(code, Response{
        Code:   code,
        Status: status,
        Data: PaginationResponse{
            Page:       page,
            Limit:      limit,
            TotalData:  total,
            TotalPages: totalPages,
            Data:       data,
        },
    })
}

func getStatusText(code int) string {
    statusMap := map[int]string{
        200: "OK",
        201: "CREATED",
        400: "BAD_REQUEST",
        401: "UNAUTHORIZED",
        403: "FORBIDDEN",
        404: "NOT_FOUND",
        409: "CONFLICT",
        422: "UNPROCESSABLE_ENTITY",
        500: "INTERNAL_SERVER_ERROR",
    }

    if status, ok := statusMap[code]; ok {
        return status
    }
    return "UNKNOWN"
}
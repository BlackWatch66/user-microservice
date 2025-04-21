package handler

import (
	"net/http"
	"strconv"

	"github.com/blackwatch66/user-microservice/api/http/middleware"
	"github.com/blackwatch66/user-microservice/internal/model"
	"github.com/blackwatch66/user-microservice/internal/service"
	"github.com/gin-gonic/gin"
)

// UserHandler 封装用户相关的 HTTP handlers
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建一个新的 UserHandler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// RegisterRoutes 注册用户相关的 HTTP 路由
func (h *UserHandler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	userGroup := router.Group("/api/users")
	{
		userGroup.POST("/signup", h.Register)
		userGroup.POST("/login", h.Login)

		authedGroup := userGroup.Group("/")
		authedGroup.Use(authMiddleware)
		{
			authedGroup.GET("/:id", h.GetProfile)           // GET /api/users/{id}
			authedGroup.PUT("/:id", h.UpdateProfile)        // PUT /api/users/{id}
			authedGroup.GET("/:id/addresses", h.ListAddresses) // GET /api/users/{id}/addresses
			authedGroup.POST("/:id/addresses", h.AddAddress) // POST /api/users/{id}/addresses
			authedGroup.PUT("/:id/addresses/:addrId", h.UpdateAddress) // PUT /api/users/{id}/addresses/{addrId}
			authedGroup.DELETE("/:id/addresses/:addrId", h.DeleteAddress) // DELETE /api/users/{id}/addresses/{addrId}
		}
	}
}

// Register 处理用户注册请求
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"` // 添加密码最小长度验证
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user", "details": err.Error()})
		}
		return
	}

	// 避免返回密码哈希
    userResponse := map[string]interface{}{
        "id": user.ID,
        "email": user.Email,
        "created_at": user.CreatedAt,
    }
	c.JSON(http.StatusCreated, userResponse)
}

// Login 处理用户登录请求
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "invalid email or password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, err := getUserIDFromParam(c)
	if err != nil {
		return // Error response handled in getUserIDFromParam
	}

	if !checkPermissions(c, userID) {
        return // Error response handled in checkPermissions
    }

	user, err := h.userService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, err := getUserIDFromParam(c)
	if err != nil {
		return
	}

    if !checkPermissions(c, userID) {
        return
    }

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	user, err := h.userService.UpdateUserProfile(c.Request.Context(), userID, req.FirstName, req.LastName)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListAddresses 获取用户地址列表
func (h *UserHandler) ListAddresses(c *gin.Context) {
	userID, err := getUserIDFromParam(c)
	if err != nil {
		return
	}

    if !checkPermissions(c, userID) {
        return
    }

	addresses, err := h.userService.GetUserAddresses(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list addresses", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, addresses)
}

// AddAddress 添加用户地址
func (h *UserHandler) AddAddress(c *gin.Context) {
	userID, err := getUserIDFromParam(c)
	if err != nil {
		return
	}

    if !checkPermissions(c, userID) {
        return
    }

	var req model.Address
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	address, err := h.userService.AddUserAddress(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, address)
}

// UpdateAddress 更新用户地址
func (h *UserHandler) UpdateAddress(c *gin.Context) {
	userID, err := getUserIDFromParam(c)
	if err != nil {
		return
	}
	addrID, err := getAddressIDFromParam(c)
	if err != nil {
        return
    }

    if !checkPermissions(c, userID) {
        return
    }

	var req model.Address
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	address, err := h.userService.UpdateUserAddress(c.Request.Context(), userID, addrID, req)
	if err != nil {
		if err.Error() == "address not found or does not belong to user" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, address)
}

// DeleteAddress 删除用户地址
func (h *UserHandler) DeleteAddress(c *gin.Context) {
	userID, err := getUserIDFromParam(c)
	if err != nil {
		return
	}
    addrID, err := getAddressIDFromParam(c)
    if err != nil {
        return
    }

    if !checkPermissions(c, userID) {
        return
    }

	err = h.userService.DeleteUserAddress(c.Request.Context(), userID, addrID)
	if err != nil {
        if err.Error() == "address not found or does not belong to user" {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address", "details": err.Error()})
        }
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper function to get user ID from URL param
func getUserIDFromParam(c *gin.Context) (uint, error) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return 0, err
	}
	return uint(userID), nil
}

// Helper function to get address ID from URL param
func getAddressIDFromParam(c *gin.Context) (uint, error) {
    addrIDStr := c.Param("addrId")
    addrID, err := strconv.ParseUint(addrIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID format"})
        return 0, err
    }
    return uint(addrID), nil
}

// Helper function to check if the authenticated user matches the requested user ID
func checkPermissions(c *gin.Context, requestedUserID uint) bool {
    claims, exists := middleware.GetUserClaims(c)
    if !exists || claims.UserID != requestedUserID {
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
        return false
    }
    return true
} 
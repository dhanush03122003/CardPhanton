package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"webauthn-server/internal/apierrors"
	"webauthn-server/internal/user"
)

// Handler exposes admin approval endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates an admin handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetStatus reports whether the authenticated user is an administrator.
func (h *Handler) GetStatus(c *gin.Context) {
	role, exists := c.Get("role")
	c.JSON(http.StatusOK, gin.H{"isAdmin": exists && role == "ADMIN"})
}

// GetAdmins returns administrator accounts without search or pagination.
func (h *Handler) GetAdmins(c *gin.Context) {
	users, err := h.service.GetAdminUsers(c.Request.Context())
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "ADMINS_FETCH_FAILED", "Admin Fetch Failed", string(ErrMsgFetchUsers))
		return
	}
	if users == nil {
		users = []user.User{}
	}
	c.JSON(http.StatusOK, users)
}

// GetUsers returns searchable, paginated normal users.
func (h *Handler) GetUsers(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	search := c.Query("search")
	users, total, err := h.service.GetUsers(c.Request.Context(), search, pageSize, (page-1)*pageSize)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "USERS_FETCH_FAILED", "User Fetch Failed", string(ErrMsgFetchUsers))
		return
	}
	if users == nil {
		users = []user.User{}
	}
	totalPages := (total + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"search":      search,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": totalPages,
	})
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return value
}

// GetUserDetails returns a user's account, authenticators, and both audit streams.
func (h *Handler) GetUserDetails(c *gin.Context) {
	details, err := h.service.GetUserDetails(c.Request.Context(), c.Param("id"))
	if errors.Is(err, ErrUserNotFound) {
		apierrors.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User Not Found", string(ErrMsgUserNotFound))
		return
	}
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "USER_DETAILS_FETCH_FAILED", "User Details Fetch Failed", string(ErrMsgFetchUserDetails))
		return
	}
	c.JSON(http.StatusOK, details)
}

func (h *Handler) adminID(c *gin.Context) (string, bool) {
	adminID, exists := c.Get("userId")
	if !exists {
		apierrors.Error(c, http.StatusUnauthorized, "ADMIN_IDENTITY_MISSING", "Admin Identity Missing", string(ErrMsgAdminIdentityMissing))
		return "", false
	}
	id, ok := adminID.(string)
	if !ok || id == "" {
		apierrors.Error(c, http.StatusUnauthorized, "INVALID_ADMIN_IDENTITY", "Invalid Admin Identity", string(ErrMsgInvalidAdminIdentity))
		return "", false
	}
	return id, true
}

// ApproveUser enables a pending user.
func (h *Handler) ApproveUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		apierrors.Error(c, http.StatusBadRequest, "USER_ID_REQUIRED", "User ID Required", string(ErrMsgUserIDRequired))
		return
	}
	adminID, ok := h.adminID(c)
	if !ok {
		return
	}
	if err := h.service.ApproveUser(c.Request.Context(), adminID, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			apierrors.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User Not Found", string(ErrMsgUserNotFound))
			return
		}
		apierrors.Error(c, http.StatusInternalServerError, "USER_APPROVAL_FAILED", "User Approval Failed", string(ErrMsgApproveUser))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User approved successfully"})
}

func (h *Handler) SuspendUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		apierrors.Error(c, http.StatusBadRequest, "USER_ID_REQUIRED", "User ID Required", string(ErrMsgUserIDRequired))
		return
	}
	adminID, ok := h.adminID(c)
	if !ok {
		return
	}
	if err := h.service.SuspendUser(c.Request.Context(), adminID, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			apierrors.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User Not Found", string(ErrMsgUserNotFound))
			return
		}
		apierrors.Error(c, http.StatusInternalServerError, "USER_SUSPENSION_FAILED", "User Suspension Failed", string(ErrMsgSuspendUser))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User suspended successfully"})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	adminID, ok := h.adminID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteUser(c.Request.Context(), adminID, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			apierrors.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User Not Found", string(ErrMsgUserNotFound))
			return
		}
		apierrors.Error(c, http.StatusInternalServerError, "USER_DELETION_FAILED", "User Deletion Failed", string(ErrMsgDeleteUser))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted successfully"})
}

func (h *Handler) DeleteAuthenticator(c *gin.Context) {
	adminID, ok := h.adminID(c)
	if !ok {
		return
	}
	err := h.service.DeleteAuthenticator(c.Request.Context(), adminID, c.Param("id"), c.Param("authId"))
	if errors.Is(err, ErrLastAuthenticator) {
		apierrors.Error(c, http.StatusBadRequest, "LAST_AUTHENTICATOR", "Authenticator Deletion Not Allowed", string(ErrMsgLastAuthenticator))
		return
	}
	if errors.Is(err, ErrAuthenticatorNotFound) || errors.Is(err, ErrUserNotFound) {
		apierrors.Error(c, http.StatusNotFound, "AUTHENTICATOR_NOT_FOUND", "Authenticator Not Found", string(ErrMsgAuthenticatorNotFound))
		return
	}
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "AUTHENTICATOR_DELETION_FAILED", "Authenticator Deletion Failed", string(ErrMsgDeleteAuthenticator))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

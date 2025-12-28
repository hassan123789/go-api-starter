// Package handler provides HTTP handlers for authentication.
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RefreshTokenRequest represents a token refresh request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the response for token refresh.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until expiration
}

// TokenHandler handles token-related operations.
type TokenHandler struct {
	tokenManager TokenManager
	userService  UserFinder
}

// TokenManager interface for token operations.
type TokenManager interface {
	ValidateRefreshToken(tokenString string) (userID int64, email, role string, err error)
	GenerateTokenPair(userID int64, email, role string) (accessToken, refreshToken string, err error)
}

// UserFinder interface for finding users.
type UserFinder interface {
	GetByID(id int64) (interface{}, error)
}

// NewTokenHandler creates a new token handler.
func NewTokenHandler(tm TokenManager, us UserFinder) *TokenHandler {
	return &TokenHandler{
		tokenManager: tm,
		userService:  us,
	}
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} RefreshTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *TokenHandler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required")
	}

	// Validate refresh token
	userID, email, role, err := h.tokenManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	// Verify user still exists and is active
	if h.userService != nil {
		_, err := h.userService.GetByID(userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
		}
	}

	// Generate new token pair
	accessToken, refreshToken, err := h.tokenManager.GenerateTokenPair(userID, email, role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate tokens")
	}

	return c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes default
	})
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

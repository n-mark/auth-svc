package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"minimal-service/internal/auth"
	"minimal-service/internal/model"
	"minimal-service/internal/repository"
	"minimal-service/pkg/response"
)

// tokenTTL is how long a freshly issued JWT remains valid.
const tokenTTL = 24 * time.Hour

// UserHandler implements registration, login and profile endpoints.
type UserHandler struct {
	repo           *repository.UserRepository
	passwordHasher auth.PasswordHasher
	jwtManager     *auth.JWTManager
}

// NewUserHandler creates a new user handler.
func NewUserHandler(
	repo *repository.UserRepository,
	hasher auth.PasswordHasher,
	jwtManager *auth.JWTManager,
) *UserHandler {
	return &UserHandler{
		repo:           repo,
		passwordHasher: hasher,
		jwtManager:     jwtManager,
	}
}

// Register handles POST /register.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var dto model.RegisterDTO
	if err := decodeJSON(r.Body, &dto); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if dto.Username == "" || dto.Password == "" {
		response.Error(w, http.StatusBadRequest, "username and password are required")
		return
	}

	hash, err := h.passwordHasher.Hash(dto.Password)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	u := &model.User{
		Username:     dto.Username,
		PasswordHash: hash,
		Email:        dto.Email,
		Phone:        dto.Phone,
	}
	if err := h.repo.Create(u); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			response.Error(w, http.StatusConflict, "user already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	response.JSON(w, http.StatusCreated, toProfileDTO(u))
}

// Login handles POST /login. Returns a JWT for the given username/password.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var dto model.LoginDTO
	if err := decodeJSON(r.Body, &dto); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.repo.GetByUsername(dto.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	if err := h.passwordHasher.Compare(u.PasswordHash, dto.Password); err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.jwtManager.GenerateToken(strconv.Itoa(u.ID), tokenTTL)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"token_type": "Bearer",
		"expires_in": int(tokenTTL.Seconds()),
	})
}

// Validate is the endpoint used by Traefik ForwardAuth middleware.
// It checks the Authorization header and returns 200 with X-User-Id set,
// or 401 if the token is missing/invalid.
func (h *UserHandler) Validate(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response.Error(w, http.StatusUnauthorized, "missing authorization header")
		return
	}

	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		response.Error(w, http.StatusUnauthorized, "invalid authorization header")
		return
	}
	token := authHeader[len(prefix):]

	userID, err := h.jwtManager.VerifyToken(token)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid token")
		return
	}

	w.Header().Set("X-User-Id", userID)
	w.WriteHeader(http.StatusOK)
}

func userIDFromRequest(r *http.Request) (int, bool) {
	sub, ok := auth.UserIDFromContext(r.Context())
	if !ok || sub == "" {
		return 0, false
	}
	id, err := strconv.Atoi(sub)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func toProfileDTO(u *model.User) model.ProfileDTO {
	return model.ProfileDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
	}
}

package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"minimal-service/internal/auth"
	"minimal-service/internal/messaging"
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
	broker         messaging.Broker
	confirmBaseURL string
}

// NewUserHandler creates a new user handler.
func NewUserHandler(
	repo *repository.UserRepository,
	hasher auth.PasswordHasher,
	jwtManager *auth.JWTManager,
	broker messaging.Broker,
	confirmBaseURL string,
) *UserHandler {
	return &UserHandler{
		repo:           repo,
		passwordHasher: hasher,
		jwtManager:     jwtManager,
		broker:         broker,
		confirmBaseURL: confirmBaseURL,
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

	plainToken, hashedToken, err := generateConfirmationToken()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate confirmation token")
		return
	}
	u := &model.User{
		Username:          dto.Username,
		PasswordHash:      hash,
		Email:             dto.Email,
		Phone:             dto.Phone,
		Status:            model.UserStatusConfirmPending,
		ConfirmationToken: &hashedToken,
	}
	if err := h.repo.Create(u); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			response.Error(w, http.StatusConflict, "user already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	confirmationLink := h.generateConfirmationLink(plainToken)
	event := model.UserCreatedEvent{
		EventId:          uuid.New(),
		EventType:        "user.created",
		NotificationType: "email",
		Message:          confirmationLink,
		Version:          "1.0",
		Email:            u.Email,
		Payload:          "",
	}
	if err := h.broker.ReportUserCreated(event); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to publish user created event")
		return
	}

	response.JSON(w, http.StatusCreated, toProfileDTO(u))
}

func (h *UserHandler) generateConfirmationLink(token string) string {
	return fmt.Sprintf("%s/confirm?token=%s", h.confirmBaseURL, token)
}

// generateConfirmationToken generates a cryptographically secure random token
// and returns both the plain token (to be sent to the user) and its SHA-256 hash (to be stored).
func generateConfirmationToken() (plain string, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate random token: %w", err)
	}
	plain = hex.EncodeToString(b)
	hashed = hashConfirmationToken(plain)
	return plain, hashed, nil
}

func hashConfirmationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Confirm handles GET /confirm?token=... and activates the user account.
func (h *UserHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	plainToken := r.URL.Query().Get("token")
	if plainToken == "" {
		response.Error(w, http.StatusBadRequest, "missing token")
		return
	}

	hashedToken := hashConfirmationToken(plainToken)
	u, err := h.repo.GetByConfirmationToken(hashedToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "invalid or expired token")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	if u.Status != model.UserStatusConfirmPending {
		response.Error(w, http.StatusConflict, "user already confirmed")
		return
	}

	if err := h.repo.ConfirmUser(u.ID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to confirm user")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"status": "confirmed"})
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

	if u.Status != model.UserStatusActive {
		response.Error(w, http.StatusForbidden, "account not activated")
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

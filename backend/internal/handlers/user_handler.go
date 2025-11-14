package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nabsk911/chronify/internal/auth"
	"github.com/nabsk911/chronify/internal/db"
	"github.com/nabsk911/chronify/internal/utils"
)

type authRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserHandler struct {
	userStore *db.Queries
	logger    *log.Logger
}

func NewUserHandler(userStore *db.Queries, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore: userStore,
		logger:    logger,
	}
}

// Register
func (uh *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		uh.logger.Printf("Failed to decode register request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid request payload!"})
		return
	}

	// Validation
	if req.Email == "" || req.Password == "" || req.Username == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Email, username and password are required"})
		return
	}

	if !utils.IsValidEmail(req.Email) {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid email format"})
		return
	}

	if len(req.Username) < 3 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Username must be at least 3 characters"})
		return
	}

	if len(req.Password) < 8 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Password must be at least 8 characters"})
		return
	}

	password_hash, err := auth.SetPasswordHash(req.Password)

	if err != nil {
		uh.logger.Printf("Failed to hash password: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Internal server error!"})
		return
	}

	user := db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: password_hash,
	}

	_, err = uh.userStore.CreateUser(r.Context(), user)
	if err != nil {
		uh.logger.Printf("Failed to create user in store: %v", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "users_email_key":
					utils.WriteJSON(w, http.StatusConflict, utils.Envelope{"message": "Email already exists"})
					return
				case "users_username_key":
					utils.WriteJSON(w, http.StatusConflict, utils.Envelope{"message": "Username already exists"})
					return
				}
			}
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to create user!"})
		return

	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"message": "User created successfully!"})
}

// Login
func (uh *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		uh.logger.Printf("Failed to decode login request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid request payload!"})
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Email and password are required"})
		return
	}

	if !utils.IsValidEmail(req.Email) {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid email format"})
		return
	}

	user, err := uh.userStore.GetUserByEmail(r.Context(), req.Email)

	if err != nil {
		uh.logger.Printf("Failed to retrieve user %s: %v", req.Email, err)
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"message": "Invalid credentials!"})
		return
	}

	passwordMatches, err := auth.CheckPasswordHash(req.Password, user.PasswordHash)

	if err != nil {
		uh.logger.Printf("Error checking password hash: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Internal server error!"})
		return
	}

	if !passwordMatches {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"message": "Invalid credentials!"})
		return
	}

	token, err := auth.GenerateToken(user.ID.String())
	if err != nil {
		uh.logger.Printf("Failed to generate token: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Internal server error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"token": token,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

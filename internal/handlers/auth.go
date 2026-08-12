package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/codersgyan/olx-api/internal/httpx"
	"github.com/codersgyan/olx-api/internal/middlerware"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"create_at"`
}

type AuthHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewAuthHandler(db *sql.DB, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger,
	}
}

func (ah AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	requestId := middlerware.RequestIDFromContext(ctx)
	log := ah.logger.With("request_id", requestId)

	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode", "request_id", requestId, "err", err)
		httpx.Error(w, http.StatusBadGateway, "invalid body", httpx.CodeMalformedJSON)
		return
	}

	if err := req.Validate(); err != nil {
		var verr *ValidationError
		errors.As(err, &verr)
		httpx.ValidationError(w, http.StatusUnprocessableEntity, err.Error(), httpx.CodeValidationError, verr.Field)
		return
	}

	// cost = how slow you want your hashing function
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		log.Error("Hashing failed", "err", err, "request_id", requestId)
		httpx.Error(w, http.StatusInternalServerError, "something went wrong", httpx.CodeInternalError)
		return
	}

	row := ah.db.QueryRowContext(
		ctx, `
		INSERT INTO users (name, email, password) VALUES ($1,$2,$3) RETURNING id, create_at`, req.Name, req.Email, hash)

	var u user
	err = row.Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httpx.Error(w, http.StatusInternalServerError, "Email already taken", httpx.CodeConflict)
			return
		}

		log.Error("scanning failed", "err", err, "request_id", requestId)
		httpx.Error(w, http.StatusInternalServerError, "something went wrong", httpx.CodeInternalError)

		return
	}

	out := SignupResponse{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
	}

	log.Info("new user registered", "user_id", out.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(out)
}

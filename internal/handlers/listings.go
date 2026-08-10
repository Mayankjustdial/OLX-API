package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/codersgyan/olx-api/internal/httpx"
	"github.com/codersgyan/olx-api/internal/middlerware"
)

type listing struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	City        string    `json:"city"`
	CreatedAt   time.Time `json:"create_at"`
}

type ListingHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewListingHandler(db *sql.DB, logger *slog.Logger) *ListingHandler {
	return &ListingHandler{
		db:     db,
		logger: logger,
	}
}

func (lh ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	// request scoped context
	ctx := r.Context()

	rows, err := lh.db.QueryContext(ctx,
		`SELECT id, title, description, price, city, create_at
		FROM listings
		ORDER BY create_at DESC
		LIMIT 100
		`)

	if err != nil {
		lh.logger.Error("listing query error", "err", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	listings := []listing{}

	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			lh.logger.Error("rows scan error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		lh.logger.Info("listing fetched", "total", len(listings))
		listings = append(listings, l)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows.err: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(listings)

}

func (lh ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Delete
	ctx := r.Context()
	requestId := middlerware.RequestIDFromContext(ctx)
	// requestId := ctx.Value("requestCtxId").(string)

	id := r.PathValue("id")

	// lh.logger.Debug("debug log", "listing_id", id)
	// lh.logger.Info("starting query", "listing_id", id)
	// lh.logger.Warn("warn log", "listing_id", id)

	_, err := lh.db.ExecContext(ctx,
		`DELETE FROM listings WHERE id = $1`, id)
	if err != nil {
		// log.Printf("delete: %v", err)
		lh.logger.Error("delete failed", "Listing_id", id, "request_id", requestId, "err", err)
		// http.Error(w, "internal error", http.StatusInternalServerError)

		httpx.Error(w, http.StatusInternalServerError, "Something went wrong", httpx.CodeInternalError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("okk"))

}

func (lh ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestId := middlerware.RequestIDFromContext(ctx)

	var req listing
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lh.logger.Error("failed to decode", "request_id", requestId, "err", err)
		httpx.Error(w, http.StatusBadGateway, "invalid body", httpx.CodeMalformedJSON)
	}

	row := lh.db.QueryRowContext(
		ctx, `
		INSERT INTO listings (title, description, price, city) VALUES ($1,$2,$3,$4) RETURNING id`, req.Title, req.Description, req.Price, req.City)

	var id string

	if err := row.Scan(&id); err != nil {
		lh.logger.Error("failed to decode", "request_id", requestId, "err", err)
		httpx.Error(w, http.StatusBadGateway, "something went wrong", httpx.CodeMalformedJSON)
		return
	}

	lh.logger.Info("listing created", "request_id", requestId, "listing_id", id)

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
}

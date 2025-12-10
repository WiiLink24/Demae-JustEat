package skip

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
)

type Client struct {
	Context    context.Context
	Address    string
	PostalCode string
	WiiID      string
	Db         *pgxpool.Pool
}

func NewClient(ctx context.Context, db *pgxpool.Pool, req *http.Request, hollywoodID string) (Client, error) {
	client := Client{
		Context:    ctx,
		Address:    req.Header.Get("X-Address"),
		PostalCode: req.Header.Get("X-PostalCode"),
		WiiID:      hollywoodID,
		Db:         db,
	}

	return client, nil
}

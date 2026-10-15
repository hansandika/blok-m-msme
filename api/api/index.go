// Package handler is the Vercel Go serverless entrypoint.
// vercel.json rewrites every path to /api, so the chi router sees the original URL.
package handler

import (
	"net/http"

	"github.com/hansandika/blok-m-msme/api/pkg/vercelapp"
)

// Handler is the Vercel entrypoint.
func Handler(w http.ResponseWriter, r *http.Request) { vercelapp.Handler(w, r) }

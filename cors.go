package main

import "net/http"

// Orígenes del frontend autorizados a llamar a la API
var origenesPermitidos = map[string]bool{
	"http://localhost:5173": true,
	// "https://tu-frontend.vercel.app": true,
}

// ConCORS responde el preflight (OPTIONS) y agrega las cabeceras CORS
func ConCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origen := r.Header.Get("Origin")
		if origenesPermitidos[origen] {
			w.Header().Set("Access-Control-Allow-Origin", origen)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

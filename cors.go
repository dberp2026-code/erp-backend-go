package main

import "net/http"

// Orígenes del frontend autorizados a llamar a la API (más flexibles para Docker/Local)
var origenesPermitidos = map[string]bool{
    "http://localhost:5173":     true,
    "http://127.0.0.1:5173":    true,
    "http://localhost":          true,
}

// ConCORS responde el preflight (OPTIONS) y agrega las cabeceras CORS
func ConCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origen := r.Header.Get("Origin")
        
        // Si el origen está permitido, lo devolvemos dinámicamente
        if origenesPermitidos[origen] {
            w.Header().Set("Access-Control-Allow-Origin", origen)
        } else {
            // Comodín permisivo para desarrollo local en contenedores
            w.Header().Set("Access-Control-Allow-Origin", "*")
        }

        w.Header().Set("Vary", "Origin")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
package main

import ("encoding/json"; "log"; "net/http"; "os"; "time"; "fmt")

func ApiHandler(w http.ResponseWriter, r *http.Request) {
   // 1. Configurar CORS para permitir peticiones desde el frontend (Vue / Docker)
   w.Header().Set("Access-Control-Allow-Origin", "*")
   w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
   w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
   w.Header().Set("Content-Type", "application/json")

   // 2. Responder inmediatamente si es una petición OPTIONS (Preflight del navegador)
   if r.Method == http.MethodOptions {
      w.WriteHeader(http.StatusNoContent)
      return
   }

   // Recuperación global de panics para esta petición
   defer func() {
      if rec := recover(); rec != nil {
         log.Printf("[PANIC RECOVER]: %v", rec)
         w.WriteHeader(http.StatusInternalServerError)
         json.NewEncoder(w).Encode(map[string]string {
            "ERROR": fmt.Sprintf("Error interno del servidor: %v", rec),
         })
      }
   }()   

   if r.Method != http.MethodPost {
      w.WriteHeader(http.StatusMethodNotAllowed)
      json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
      return
   }

   var paData map[string]any
   if err := json.NewDecoder(r.Body).Decode(&paData); err != nil {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"error": "JSON de entrada inválido"})
      return
   }
   id, ok := paData["ID"].(string)
   if !ok || id == "" {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"error": "ID no definido"})
      return
   }
   fmt.Println("ID:", id, "-", time.Now().Format("2006-01-02 15:04:05.00"))
   // Mesa de Partes
   if id[:3] == "MPV" {
      ltInicio := time.Now()
      HMesaPartes(w, r, paData)
      fmt.Println("Tiempo transcurrido:", time.Since(ltInicio))
      return
   }   
   //var lcRpta string
   var laData []byte
   var err error
   if id == "LOGIN" || id == "LOGINE" {
      ltInicio := time.Now()
      lo := SLogin{Data: paData}
      laData, err = lo.omLogin()
      fmt.Println("Tiempo transcurrido:", time.Since(ltInicio))
   } else if id == "ICLAVE" {
      ltInicio := time.Now()
      lo := SClave{Data: paData}
      laData, err = lo.omInicializarClaves()
      fmt.Println("Tiempo transcurrido:", time.Since(ltInicio))
   /*   
   } else if id == "CLAVE" {
      lo := SClave{Data: laData}
      lcRpta, err = lo.omCifrarClave()
   } else if id == "CLAVE1" {
      lo := SClave{Data: laData}
      lcRpta, err = lo.omCompararClave()
   } else {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"ERROR": "ID no reconocido"})
      return
   */   
   }
   if err != nil {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"ERROR": err.Error()})
      return
   }
   w.WriteHeader(http.StatusOK)
   //w.Write([]byte(lcRpta))
   w.Write(laData)
}

/*
func ApiHandler_old(w http.ResponseWriter, r *http.Request) {
   w.Header().Set("Content-Type", "application/json")

   if r.Method != http.MethodPost {
      w.WriteHeader(http.StatusMethodNotAllowed)
      json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
      return
   }

   var datosEntrada map[string]any
   if err := json.NewDecoder(r.Body).Decode(&datosEntrada); err != nil {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"error": "JSON de entrada inválido"})
      return
   }

   id, ok := datosEntrada["ID"].(string)
   if !ok || id == "" {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"error": "ID no definido"})
      return
   }

   var resultadoJSON string
   var err error

   if id == "LOGIN" {
      loginService := SLogin{Datos: datosEntrada}
      resultadoJSON, err = loginService.omLogin()
   } else if id == "CLAVE" {
      claveService := SClave{Datos: datosEntrada}
      resultadoJSON, err = claveService.omCifrarClave()
   } else {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"error": "ID no reconocido"})
      return
   }
   if err != nil {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
      return
   }
   w.WriteHeader(http.StatusOK)
   w.Write([]byte(resultadoJSON))
}
*/

func main() {
   if err := ConnectDB(); err != nil {
      log.Fatalf("No se pudo conectar a la BD: %v", err)
   }
   defer poPool.Close()
   
   http.HandleFunc("/", ApiHandler)
   
   // Leer el puerto asignado por Render o usar 8080 por defecto
   port := os.Getenv("PORT")
   if port == "" {
      port = "8080"
   }

   log.Println("Servidor API REST escuchando en el puerto " + port)
   if err := http.ListenAndServe(":"+port, nil); err != nil {
      log.Fatalf("Error en el servidor: %v", err)
   }
}
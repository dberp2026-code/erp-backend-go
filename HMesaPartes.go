// HMesaPartes.go
package main

import ("encoding/json"; "net/http"; "fmt")

// Manejador secundario especializado en Mesa de Partes
func HMesaPartes(w http.ResponseWriter, r *http.Request, p_aData map[string]any) {
   var laData []byte
   var err error
   //if id == "MPV1010i" {
   if p_aData["ID"].(string) == "MPV1010i" {
      lo := SMesaPartes{Data: p_aData}
      laData, err = lo.omInitRegistrarExpediente()
   } else if p_aData["ID"].(string) == "MPV1010g" {
      fmt.Println("111")
      lo := SMesaPartes{Data: p_aData}
      laData, err = lo.omGrabarExpediente()
   } else if p_aData["ID"].(string) == "MPV1020i" {
      lo := SMesaPartes{Data: p_aData}
      laData, err = lo.omInitEnviarExpedientes()
   } else {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"ERROR": "ID de Mesa de Partes no reconocido"})
      return
   }
   if err != nil {
      w.WriteHeader(http.StatusBadRequest)
      json.NewEncoder(w).Encode(map[string]string{"ERROR": err.Error()})
      return
   }
   w.WriteHeader(http.StatusOK)
   w.Write(laData)
}
package main

import ("crypto/md5"; "encoding/json"; "errors"; "fmt"; "math/rand"; "regexp"; "slices"; "context"; "encoding/hex")

// Estructura única: "Data" recibe cualquier etiqueta y también devuelve las que convengan
type SClave struct {
   Data map[string]any `json:"data"`
}

// ----------------------------------------------
// Cifra la clave
// 2026-09-01 FPM Creacion
// ----------------------------------------------
func (so *SClave) omCifrarClave() ([]byte, error) {
   if err := so.mxValParam(); err != nil {
      return nil, err
   }
   laData := map[string]any{"CCLAVE": so.Data["CCLAVE"].(string), "CSEED": fmt.Sprintf("%02x", rand.Intn(256))}
   lcClaveC, err := so.mxCifrarClave(laData)
   if err != nil {
      return nil, err
   }
   laData = map[string]any{"CCLAVE": lcClaveC}
   paData, err := json.Marshal(laData)
   if err != nil {
      return nil, err
   }
   return paData, nil
}

// Valida que existan los campos obligatorios en so.Data y que cumplan el formato esperado
func (so *SClave) mxValParam() error {
   _, llOk := so.Data["CCLAVE"].(string)
   if !llOk || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(so.Data["CCLAVE"].(string)) {
      return errors.New("CLAVE NO DEFINIDA O INVÁLIDA")
   }
   return nil
}

func (so *SClave) mxCifrarClave(p_Data map[string]any) (string, error) {
   lcClave := fmt.Sprintf("%x", md5.Sum([]byte(p_Data["CSEED"].(string) + p_Data["CCLAVE"].(string))))
   lcClave1 := []byte(lcClave)
   slices.Reverse(lcClave1)
   lcClave = string(lcClave1)
   lcClave = p_Data["CSEED"].(string) + lcClave[len(lcClave)-30:]
   return lcClave, nil
}

// ----------------------------------------------
// Comparar clave
// 2026-09-01 FPM Creacion
// ----------------------------------------------
func (so *SClave) omCompararClave() (error) {
   if err := so.mxValParamCompararClave(); err != nil {
      return err
   }
   if err := so.mxCompararClave(); err != nil {
      return err
   }
   return nil
}

// Valida que existan los campos obligatorios en so.Data y que cumplan el formato esperado
func (so *SClave) mxValParamCompararClave() error {
   _, llOk := so.Data["CCLAVE"].(string)
   if !llOk || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(so.Data["CCLAVE"].(string)) {
      return errors.New("CLAVE NO DEFINIDA O INVÁLIDA")
   }
   _, llOk = so.Data["CCLAVEC"].(string)
   if !llOk || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(so.Data["CCLAVEC"].(string)) {
      return errors.New("CLAVE CIFRADA NO DEFINIDA O INVÁLIDA")
   }
   return nil
}

func (so *SClave) mxCompararClave() error {
   laData := map[string]any{"CCLAVE": so.Data["CCLAVE"].(string), "CSEED": so.Data["CCLAVEC"].(string)[:2]}
   lcClaveC, err := so.mxCifrarClave(laData)
   if err != nil {
      return err
   }
   //fmt.Println(lcClaveC, so.Data["CCLAVEC"])
   if lcClaveC != so.Data["CCLAVEC"] {
      return errors.New("CLAVE ERRADA")
   }
   return nil
}

// ----------------------------------------------
// Funcion que actualiza todos las claves
// 2026-09-01 FPM Creacion
// ----------------------------------------------
func (so *SClave) omInicializarClaves() ([]byte, error) {
   if err := so.mxInicializarClaves(); err != nil {
      return nil, err
   }
   return []byte(`{"OK": "OK"}`), nil
}

func (so *SClave) mxInicializarClaves() error {
   loCtx := context.Background()
   const lnLote = 10000
   lnGrupo := 0
   lcUltDni := ""
   for {
       lnGrupo++
       fmt.Println("Grupo:", lnGrupo)
       // 1. BEGIN por lote
       tx, err := poPool.Begin(loCtx)
       if err != nil {
          return err
       }
       lcSql := "SELECT cNroDni FROM S01MPER WHERE cNroDni > $1 ORDER BY cNroDni LIMIT $2"
       RS, err := tx.Query(loCtx, lcSql, lcUltDni, lnLote)
       if err != nil {
          tx.Rollback(loCtx)
          return err
       }
       var laNroDni []string
       for RS.Next() {
           laTmp, err := RS.Values()
           if err != nil {
              RS.Close()
              tx.Rollback(loCtx)
              return err
           }
           laNroDni = append(laNroDni, laTmp[0].(string))
       }
       if err := RS.Err(); err != nil {
          RS.Close()
          tx.Rollback(loCtx)
          return err
       }
       RS.Close()
       // No hay más filas: terminamos
       if len(laNroDni) == 0 {
          tx.Rollback(loCtx) // no hay nada que confirmar
          break
       }
       for _, lcNroDni := range laNroDni {
           lcSeed := fmt.Sprintf("%02x", rand.Intn(256))
           sum := md5.Sum([]byte(lcNroDni))       // [16]byte
           lcClave := hex.EncodeToString(sum[:])
           laData := map[string]any{"CCLAVE": lcClave, "CSEED": lcSeed}
           lcClave, err := so.mxCifrarClave(laData)
           if err != nil {
              tx.Rollback(loCtx)
              return err
           }
           lcSql = "UPDATE S01MPER SET cClave = $1 WHERE cNroDni = $2"
           _, err = tx.Exec(loCtx, lcSql, lcClave, lcNroDni)
           if err != nil {
              tx.Rollback(loCtx)
              return err
           }
       }
       // 3. COMMIT del lote
       if err := tx.Commit(loCtx); err != nil {
          return err
       }
       // Avanzamos el cursor al último DNI de este lote
       lcUltDni = laNroDni[len(laNroDni)-1]
       // Si el lote vino incompleto, ya no hay más filas
       if len(laNroDni) < lnLote {
          break
       }
   }
   return nil
}
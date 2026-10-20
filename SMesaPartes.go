package main

import ("crypto/md5"; "encoding/json"; "errors"; "fmt"; "math/rand"; "regexp"; "slices"; "context"; "math"; "github.com/jackc/pgx/v5"; "strings")
   
// Estructura única: "Data" recibe cualquier etiqueta y también devuelve las que convengan
type SMesaPartes struct {
   Data map[string]any `json:"data"`
}

func (so *SMesaPartes) mxValParamDNI() error {
   if lcNroDni, llOk := so.Data["CNRODNI"].(string); !llOk || !regexp.MustCompile(`^[0-9]{8}$`).MatchString(lcNroDni) {
      return errors.New("NÚMERO DE DNI NO DEFINIDO O INVÁLIDO")
   }
   return nil
}

// ----------------------------------------------
// Init registro de expediente
// 2026-09-01 FPM Creacion
// ----------------------------------------------
func (so *SMesaPartes) omInitRegistrarExpediente() ([]byte, error) {
   if err := so.mxValParam(); err != nil {
      return nil, err
   }
   paData, err := so.mxInitRegistrarExpediente()
   if err != nil {
      return nil, err
   }
   return paData, nil
}

func (so *SMesaPartes) mxValParam() (error) {
   return nil
}

func (so *SMesaPartes) mxInitRegistrarExpediente() ([]byte, error) {
   laDatos := []map[string]string{}
   lcSql := "SELECT LEFT(cCodigo, 2), cDescri FROM V_S01TTAB WHERE cCodTab = '508'"
   RS, err := poPool.Query(context.TODO(), lcSql)
   if err != nil {
      return nil, err
   }
   defer RS.Close()
   for RS.Next() {
       laTmp, err := RS.Values()
       if err != nil {
          return nil, err
       }
       laFila := map[string]string{"CCODIGO": laTmp[0].(string), "CDESCRI": laTmp[1].(string),}
       laDatos = append(laDatos, laFila)
   }
   if err := RS.Err(); err != nil {
      return nil, err
   }
   return json.Marshal(laDatos)
}

// ----------------------------------------------
// Grabar expediente
// 2026-09-28 FPM Creacion
// ----------------------------------------------
func (so *SMesaPartes) omGrabarExpediente() ([]byte, error) {
   fmt.Println("222")
   if err := so.mxValParamGrabarExpediente(); err != nil {
      return nil, err
   }
   fmt.Println("333")
   poTx, err := poPool.Begin(context.TODO())
   if err != nil {
      return nil, err
   }
   defer poTx.Rollback(context.TODO())
   fmt.Println("444")
   paData, err := so.mxGrabarExpediente(poTx)
   if err != nil {
      return nil, err
   }
   fmt.Println("555")
   if err := poTx.Commit(context.TODO()); err != nil {
      return nil, err
   }
   fmt.Println(string(paData))
   return paData, nil
}

func (so *SMesaPartes) mxValParamGrabarExpediente() error {
   if err := so.mxValParamDNI(); err != nil {
      return err
   }
   if lcTipo, llOk := so.Data["CTIPO"].(string); !llOk || !regexp.MustCompile(`^[0-9]{2}$`).MatchString(lcTipo) {
      return errors.New("TIPO DE EXPEDIENTE NO DEFINIDO O INVÁLIDO")
   }
   if lnIdAlum, llOk := so.Data["NIDALUM"].(float64); !llOk || lnIdAlum != math.Trunc(lnIdAlum) || lnIdAlum < 0 {
      return errors.New("ID DE ALUMNO NO DEFINIDO O INVÁLIDO")
   }
   if lcAsunto, llOk := so.Data["CASUNTO"].(string); !llOk || !regexp.MustCompile(`^[A-Za-z0-9 ÁÉÍÓÚáéíóúÑñ,.;:/-]{15,100}$`).MatchString(lcAsunto) {
      return errors.New("ASUNTO NO DEFINIDO O INVÁLIDO")
   }
   if lcLink, llOk := so.Data["MLINK"].(string); !llOk || !(lcLink == "" || (len(lcLink) <= 2000 && regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]{8,}$`).MatchString(lcLink))) {
      return errors.New("ENLACE DE ARCHIVO EXTERNO NO DEFINIDO O INVÁLIDO")
   }
   if lcAdjunt, llOk := so.Data["CADJUNT"].(string); !llOk || !regexp.MustCompile(`^[SN]$`).MatchString(lcAdjunt) {
      return errors.New("FLAG DE ARCHIVO ADJUNTO NO DEFINIDO O INVÁLIDO")
   }
   if lcTexto, llOk := so.Data["MTEXTO"].(string); !llOk || !regexp.MustCompile(`^[A-Za-z0-9 ÁÉÍÓÚáéíóúÑñ,.;:/-]{15,1000}$`).MatchString(lcTexto) {
      return errors.New("TEXTO (CONTENIDO DE SOLICITUD) NO DEFINIDO O INVÁLIDO")
   }
   return nil
}

func (so *SMesaPartes) mxGrabarExpediente(p_Tx pgx.Tx) ([]byte, error) {
   lcNroDni := so.Data["CNRODNI"].(string)
   lnIdAlum := int64(so.Data["NIDALUM"].(float64))
   lcSql := "SELECT cNroDni FROM S01MPER WHERE cNroDni = $1"
   if err := p_Tx.QueryRow(context.TODO(), lcSql, so.Data["CNRODNI"]).Scan(&lcNroDni); err != nil {
      if errors.Is(err, pgx.ErrNoRows) {
         return nil, errors.New("DNI NO EXISTE EN EL MAESTRO DE PERSONAS")
      }
      return nil, err
   }
   // Verifica Id de alumno
   lcEstado := ""
   lcSql = "SELECT cEstado FROM A01MALU WHERE nIdAlum = $1"
   if err := p_Tx.QueryRow(context.TODO(), lcSql, lnIdAlum).Scan(&lcEstado); err != nil {
      if errors.Is(err, pgx.ErrNoRows) {
         return nil, errors.New("ID NO EXISTE EN EL MAESTRO DE ALUMNOS")
      }
      return nil, err
   }
   if lcEstado != "A" {
      return nil, errors.New("ALUMNO NO ESTÁ ACTIVO")
   }
   // Saca el siguiente numero de expediente (el year sale de la base de datos)
   lnNumero, lnYear := 0, 0
   lcSql = `SELECT COALESCE(MAX(nNumero), 0), CAST(TO_CHAR(NOW(), 'YYYY') AS SMALLINT)
            FROM B01MEXP WHERE nYear = CAST(TO_CHAR(NOW(), 'YYYY') AS SMALLINT)`
   if err := p_Tx.QueryRow(context.TODO(), lcSql).Scan(&lnNumero, &lnYear); err != nil {
      return nil, err
   }
   lnNumero++
   lmBitaco, err := json.Marshal(map[string]string{"CESTADO": "A", "TMODIFI": "2026-09-28 23:33", "CCODUSU": "ZZZZ"})
   if err != nil {
      return nil, err
   }
   lcSql = `INSERT INTO B01MEXP (nNumero, cNroDni, nIdAlum, cTipo, cAsunto, mTexto, cAdjunt, mLink, mBitaco)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
   _, err = p_Tx.Exec(context.TODO(), lcSql, lnNumero, lcNroDni, lnIdAlum, so.Data["CTIPO"], so.Data["CASUNTO"],
                      so.Data["MTEXTO"], so.Data["CADJUNT"], so.Data["MLINK"], lmBitaco)
   if err != nil {
      fmt.Println(err)
      return nil, errors.New("NO SE PUDO REGISTRAR EXPEDIENTE")
   }
   lcNumero := fmt.Sprintf("%04d-%06d", lnYear, lnNumero)
   return json.Marshal(map[string]string{"CNUMERO": lcNumero})
}

// ----------------------------------------------
// Enviar expedientes
// 2026-09-28 FPM Creacion
// ----------------------------------------------
func (so *SMesaPartes) omInitEnviarExpedientes() ([]byte, error) {
   /*
   if err := so.mxValParamDNI(); err != nil {
      return nil, err
   }
   */
   paData, err := so.mxInitEnviarExpedientes()
   if err != nil {
      return nil, err
   }
   return paData, nil
}

func (so *SMesaPartes) mxInitEnviarExpedientes() ([]byte, error) {
   lcSql := `SELECT A.nIdExpe, TO_CHAR(A.tRecepc, 'YYYY'), A.nNumero, A.cNroDni, B.cNombre, A.cTipo, C.cDescri, A.cAsunto
             FROM B01MEXP A
             INNER JOIN S01MPER B ON B.cNroDni = A.cNroDni
             LEFT OUTER JOIN V_S01TTAB C ON C.cCodTab = '508' AND LEFT(C.cCodigo, 2) = A.cTipo
             WHERE A.cEstado = 'A'
             ORDER BY A.tRecepc ASC`
   RS, err := poPool.Query(context.Background(), lcSql)
   if err != nil {
      return nil, err
   }
   defer RS.Close()
   laDatos := []map[string]any{}
   for RS.Next() {
      laTmp, err := RS.Values()
      if err != nil {
         return nil, err
      }
      lcNumero := fmt.Sprintf("%s-%06d", laTmp[1], laTmp[2])
      lcNombre := strings.ReplaceAll(laTmp[4].(string), "/", " ")
      laDatos = append(laDatos, map[string]any{"NIDEXPE": laTmp[0], "CNUMERO": lcNumero, "CNRODNI": laTmp[3], "CNOMBRE": lcNombre,
                                               "CTIPO": laTmp[5], "CDESTIP": laTmp[6], "CASUNTO": laTmp[7]})
   }
   if err := RS.Err(); err != nil {
      return nil, err
   }
   return json.Marshal(map[string]any{"DATOS": laDatos})
}
/*
func (so *SMesaPartes) mxGrabarExpediente_old(p_Tx pgx.Tx) ([]byte, error) {
   lcNroDni := so.Data["CNRODNI"].(string)
   lnIdAlum := int64(so.Data["NIDALUM"].(float64))
   // Verifica DNI
   lcSql := "SELECT cNroDni FROM S01MPER WHERE cNroDni = $1"
   RS, err := p_Tx.Query(context.TODO(), lcSql, lcNroDni)
   if err != nil {
      return nil, err
   }
   llOk := RS.Next()
   RS.Close()
   if !llOk {
      return nil, errors.New("DNI NO EXISTE EN EL MAESTRO DE PERSONAS")
   }
   // Verifica Id de alumno
   lcSql = "SELECT cEstado FROM A01MALU WHERE nIdAlum = $1"
   RS, err = p_Tx.Query(context.TODO(), lcSql, lnIdAlum)
   if err != nil {
      return nil, err
   }
   llOk = RS.Next()
   RS.Close()
   if !llOk {
      return nil, errors.New("ID NO EXISTE EN EL MAESTRO DE ALUMNOS")
   }
   // Saca el siguiente numero de expediente (el year sale de la base de datos)
   lnNumero, lnYear := 0, 0
   lcSql = `SELECT COALESCE(MAX(nNumero), 0), CAST(TO_CHAR(NOW(), 'YYYY') AS SMALLINT)
            FROM B01MEXP WHERE nYear = CAST(TO_CHAR(NOW(), 'YYYY') AS SMALLINT)`
   if err := p_Tx.QueryRow(context.TODO(), lcSql).Scan(&lnNumero, &lnYear); err != nil {
      return nil, err
   }
   lnNumero++
   lmBitaco := fxBitacora(map[string]string{}, map[string]string{"CESTADO": "A", "TMODIFI": "", "CCODUSU": "ZZZZ"})
   lcSql = `INSERT INTO B01MEXP (nNumero, cNroDni, nIdAlum, cTipo, cAsunto, mTexto, cAdjunt, mLink, mBitaco)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
   _, err = p_Tx.Exec(context.TODO(), lcSql, lnNumero, lcNroDni, lnIdAlum, so.Data["CTIPO"], so.Data["CASUNTO"],
                      so.Data["MTEXTO"], so.Data["CADJUNT"], so.Data["MLINK"], lmBitaco)
   if err != nil {
      fmt.Println(err)
      return nil, errors.New("NO SE PUDO REGISTRAR EXPEDIENTE")
   }
   lcNumero := fmt.Sprintf("%04d-%06d", lnYear, lnNumero)
   return json.Marshal(map[string]string{"CNUMERO": lcNumero})
}
*/
/*
// Valida que existan los camp.(string)os obligatorios en so.(string).Data y que cumplan el formato esperado
func (so *SMesaPartes) mxValParam() error {
   _, llOk := so.Data["CCLAVE"].(string)
   if !llOk || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(so.Data["CCLAVE"].(string)) {
      return errors.New("CLAVE NO DEFINIDA O INVÁLIDA")
   }
  : return nil
}
*/

func (so *SMesaPartes) mxCifrarClave(p_Data map[string]any) (string, error) {
   lcClave := fmt.Sprintf("%x", md5.Sum([]byte(p_Data["CSEED"].(string) + p_Data["CCLAVE"].(string))))
   // Para invertir md5
   lcClave1 := []byte(lcClave)
   slices.Reverse(lcClave1)
   lcClave = string(lcClave1)
   // Genera clave con semilla
   lcClave = p_Data["CSEED"].(string) + lcClave[len(lcClave)-30:]
   return lcClave, nil
}

// ----------------------------------------------
// Comparar clave
// 2026-09-01 FPM Creacion
// ----------------------------------------------
func (so *SMesaPartes) omCompararClave() (string, error) {
   if err := so.mxValParamCompararClave(); err != nil {
      return "", err
   }
   if err := so.mxCompararClave(); err != nil {
      return `{"ERROR": "CLAVE NO CONCUERDA"}`, err
   }
   return `{"OK": "OK"}`, nil
}

// Valida que existan los campos obligatorios en so.Data y que cumplan el formato esperado
func (so *SMesaPartes) mxValParamCompararClave() error {
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

func (so *SMesaPartes) mxCompararClave() error {
   laData := map[string]any{"CCLAVE": so.Data["CCLAVE"].(string), "CSEED": so.Data["CCLAVEC"].(string)[:2]}
   lcClaveC, err := so.mxCifrarClave(laData)
   if err != nil {
      return err
   }
   if lcClaveC != so.Data["CCLAVEC"] {
      return errors.New("CLAVE ERRADA")
   }
   return nil
}

// ----------------------------------------------
// Funcion que actualiza todos las claves
// 2026-09-01 FPM Creacion
// ----------------------------------------------
func (so *SMesaPartes) omActualizarClaves() (error) {
   if err := so.mxActualizarClaves(); err != nil {
      return err
   }
   return nil
}

func (so *SMesaPartes) mxActualizarClaves() error {
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
           laData := map[string]any{"CCLAVE": lcNroDni, "CSEED": lcSeed}
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
package main

import ("context"; "encoding/json"; "errors"; "regexp")
import ("fmt")

// Estructura única: "Data" recibe cualquier etiqueta y también devuelve las que convengan
type SLogin struct {
   Data map[string]any `json:"data"`
}

// ----------------------------------------------
// Funcion principal de login
// 2026-09-01 FPM Creacion
// ----------------------------------------------
func (so *SLogin) omLogin() ([]byte, error) {
   if err := so.mxValParam(); err != nil {
      return nil, err
   }
   if err := so.mxLogin(); err != nil {
      return nil, err
   }
   laData, err := json.Marshal(so.Data)
   if err != nil {
      return nil, err
   }
   return laData, nil
}

// Valida que existan los campos obligatorios en so.Data y que cumplan el formato esperado
func (so *SLogin) mxValParam() error {
   _, llOk := so.Data["CNRODNI"].(string)
   if !llOk || !regexp.MustCompile(`^[0-9]{8}$`).MatchString(so.Data["CNRODNI"].(string)) {
      return errors.New("DNI NO DEFINIDO O INVÁLIDO")
   }
   _, llOk = so.Data["CCLAVE"].(string)
   if !llOk || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(so.Data["CCLAVE"].(string)) {
      return errors.New("CLAVE NO DEFINIDA O INVÁLIDA")
   }
   return nil
}

func (so *SLogin) mxLogin() error {
   // Busca nombre y clave del usuario
   lcSql := "SELECT cNombre, cClave FROM S01MPER WHERE cNroDni = $1"
   RS, err := poPool.Query(context.Background(), lcSql, so.Data["CNRODNI"].(string))
   if err != nil {
      return err
   }   
   if !RS.Next() {
      RS.Close()
      return errors.New("DNI NO EXISTE")
   }
   laFila, err := RS.Values()
   RS.Close()
   if err != nil {
      return err
   }
   lcNombre := laFila[0]
   // 2. Verificar la clave
   lo := SClave{Data: map[string]any{"CCLAVE": so.Data["CCLAVE"].(string), "CCLAVEC": laFila[1]}}
   if err := lo.omCompararClave(); err != nil {
      return err
   }
   // 3. Traer las unidades académicas del alumno
   lcSql = "SELECT A.cCodEst, A.cUniAca, B.cNomUni FROM A01MEST A INNER JOIN A01MUAC B ON B.cUniAca = A.cUniAca WHERE A.cNroDni = $1"
   RS, err = poPool.Query(context.Background(), lcSql, so.Data["CNRODNI"].(string))
   if err != nil {
      return err
   }
   defer RS.Close()
   laDatos := []map[string]any{}
   for RS.Next() {
       laTmp, err := RS.Values()
       if err != nil {
          return err
       }
       laDatos = append(laDatos, map[string]any{"CCODALU": laTmp[0], "CUNIACA": laTmp[1], "CNOMUNI": laTmp[2],})
   }
   if err := RS.Err(); err != nil {
      return err
   }
   // Respuesta limpia, sin arrastrar campos del parámetro de entrada (ID, CCLAVE, etc.)
   so.Data = map[string]any{"CNOMBRE": lcNombre, "CNRODNI": so.Data["CNRODNI"].(string), "DATOS": laDatos,}
   fmt.Println("OK")
   return nil
}

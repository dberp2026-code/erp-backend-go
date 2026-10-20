package main

import ("context"; "log"; "os"; "time"; "github.com/jackc/pgx/v5"; "github.com/jackc/pgx/v5/pgxpool")

var poPool *pgxpool.Pool

func ConnectDB() error {
    // 1. Intenta leer la variable de entorno (para Render en la nube)
    dbURL := os.Getenv("DATABASE_URL")
    
    //2. Respaldo automático para pruebas locales si no hay variable en la terminal
    if dbURL == "" {
        dbURL = "postgresql://postgres.xbcuwcybhynusvxslgxy:Dberp20262026$@aws-0-us-east-1.pooler.supabase.com:6543/postgres?pgbouncer=true"
    }

    laConfig, err := pgxpool.ParseConfig(dbURL)
    if err != nil {
        return err
    }

    // Configuración del pool de conexiones (CORREGIDO SIN GUIONES Y CON MULTIPLICACIÓN)
    laConfig.MaxConns = 25
    laConfig.MinConns = 5
    laConfig.MaxConnLifetime = 5 * time.Minute
    laConfig.MaxConnIdleTime = 2 * time.Minute

    // FORZAR PROTOCOLO SIMPLE: Evita el error SQLSTATE 42P05
    laConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

    poPool, err = pgxpool.NewWithConfig(context.Background(), laConfig)
    if err != nil {
        return err
    }

    if err := poPool.Ping(context.Background()); err != nil {
        return err
    }

    log.Println("Conectado exitosamente a la base de datos (Supabase)")
    return nil
}

// Prueba para cambios en repositorio...
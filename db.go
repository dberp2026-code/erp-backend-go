package main

import ("context";"log";"os";"time"; "github.com/jackc/pgx/v5/pgxpool")

var poPool *pgxpool.Pool

func ConnectDB() error {
   // Intenta leer la variable de entorno de la nube (Render)
	dbURL := os.Getenv("DATABASE_URL")

	laConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return err
	}
	laConfig.MaxConns = 25
	laConfig.MinConns = 5
	laConfig.MaxConnLifetime = 5 * time.Minute
	laConfig.MaxConnIdleTime = 2 * time.Minute

	poPool, err = pgxpool.NewWithConfig(context.Background(), laConfig)
	if err != nil {
		return err
	}
	if err := poPool.Ping(context.Background()); err != nil {
		return err
	}
	log.Println("Conectado a la base de datos")
	return nil
}


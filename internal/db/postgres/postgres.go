package postgres

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

type Config struct {
	Host        string `env:"MF_AUTH_DB_HOST"           envDefault:"localhost"`
	Port        string `env:"MF_AUTH_DB_PORT"           envDefault:"5432"`
	User        string `env:"MF_AUTH_DB_USER"           envDefault:"mainflux"`
	Pass        string `env:"MF_AUTH_DB_PASS"           envDefault:"mainflux"`
	Name        string `env:"MF_AUTH_DB"                envDefault:"auth"`
	SSLMode     string `env:"MF_AUTH_DB_SSL_MODE"       envDefault:"disable"`
	SSLCert     string `env:"MF_AUTH_DB_SSL_CERT"       envDefault:""`
	SSLKey      string `env:"MF_AUTH_DB_SSL_KEY"        envDefault:""`
	SSLRootCert string `env:"MF_AUTH_DB_SSL_ROOT_CERT"  envDefault:""`
}

// SetupDB create connection to database and migrate
func SetupDB(cfg Config, migrations ...migrate.MemoryMigrationSource) (*sqlx.DB, error) {
	db, err := Connect(cfg)
	if err != nil {
		return nil, err
	}
	for _, migration := range migrations {
		if err := MigrateDB(db, migration); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate failure.
func Connect(cfg Config) (*sqlx.DB, error) {
	url := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass, cfg.SSLMode, cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert)

	db, err := sqlx.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func MigrateDB(db *sqlx.DB, migrations migrate.MemoryMigrationSource) error {
	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}

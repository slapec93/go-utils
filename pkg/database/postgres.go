package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	_ "github.com/lib/pq" // postgres lib
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"

	"gorm.io/driver/postgres" // postgres driver
	"gorm.io/gorm"
)

const (
	dbDialect = "postgres"
)

// DBConnectionParams ...
type DBConnectionParams struct {
	DatabaseURL string
	Host        string
	User        string
	DBName      string
	Password    string
	SSLMode     string
}

// InitializeDatabase ...
func InitializeDatabase(connParams DBConnectionParams) (*gorm.DB, func(*gorm.DB), error) {
	connString, err := connectionString(connParams, true)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	gormDB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to initialize database connection")
	}

	return gormDB, closeDB, nil
}

func closeDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get database: %s\n", err)
	}
	if db != nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Failed to close DB: %\n", err)
		}
	}
}

func (cp DBConnectionParams) validate() error {
	if cp.DatabaseURL != "" {
		return nil
	}
	if cp.Host == "" {
		return errors.New("No database host specified")
	}
	if cp.DBName == "" {
		return errors.New("No database name specified")
	}
	if cp.User == "" {
		return errors.New("No database user specified")
	}
	if cp.Password == "" {
		return errors.New("No database password specified")
	}
	return nil
}

func connectionString(defaultParams DBConnectionParams, withDB bool) (string, error) {
	connParams := defaultParams
	if connParams.DatabaseURL == "" {
		connParams.DatabaseURL = os.Getenv("DATABASE_URL")
		if connParams.DatabaseURL != "" {
			parsedURL, err := url.Parse(connParams.DatabaseURL)
			if err != nil {
				return "", err
			}
			connParams.Host = strings.Split(parsedURL.Host, ":")[0]
			connParams.User = parsedURL.User.Username()
			connParams.Password, _ = parsedURL.User.Password()
			connParams.DBName = strings.TrimPrefix(parsedURL.Path, "/")
		}
	}
	if connParams.Host == "" {
		connParams.Host = os.Getenv("DB_HOST")
	}
	if connParams.DBName == "" {
		connParams.DBName = os.Getenv("DB_NAME")
	}
	if connParams.User == "" {
		connParams.User = os.Getenv("DB_USER")
	}
	if connParams.Password == "" {
		connParams.Password = os.Getenv("DB_PWD")
	}
	if connParams.SSLMode == "" {
		connParams.SSLMode = os.Getenv("DB_SSL_MODE")
	}
	if err := connParams.validate(); err != nil {
		return "", err
	}
	connString := fmt.Sprintf("host=%s user=%s password=%s",
		connParams.Host, connParams.User, connParams.Password)
	if withDB {
		connString += " dbname=" + connParams.DBName
	}
	// optionals
	if connParams.SSLMode != "" {
		connString += " sslmode=" + connParams.SSLMode
	}
	return connString, nil
}

// RunMigrations ...
func RunMigrations() error {
	migrations := &migrate.FileMigrationSource{
		Dir: "db/migrations",
	}
	dbURL, err := connectionString(DBConnectionParams{}, true)
	if err != nil {
		return errors.WithMessage(err, "Failed to initialize database")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return errors.WithMessage(err, "Failed to open DB connection")
	}
	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("Failed to close DB: %s\n", err)
		}
	}()
	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return errors.WithMessage(err, "Failed to run migrations")
	}
	fmt.Printf("%d new migrations migrated\n", n)
	return nil
}

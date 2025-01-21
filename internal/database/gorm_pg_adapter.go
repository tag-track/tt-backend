package database

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GormPgAdapter implements DatabaseConnectorStrategy
type GormPgAdapter struct {
	host     string
	user     string
	password string
	port     int
	dbname   string
	timezone string

	db *gorm.DB
}

////////////////////////////////////////////////
// Constructors
////////////////////////////////////////////////

type GormPgAdapterConstructorOption func(adapter *GormPgAdapter)

func CreateGormPgAdapter(
	host string,
	user string,
	password string,
	port int,
	dbname string,
	opts ...GormPgAdapterConstructorOption,
) (*GormPgAdapter, error) {

	newAdapter := &GormPgAdapter{
		host:     host,
		user:     user,
		password: password,
		port:     port,
		dbname:   dbname,
		timezone: "Asia/Singapore",
	}

	for _, opt := range opts {
		opt(newAdapter)
	}

	return newAdapter, nil
}

func (g *GormPgAdapter) createDsnString() string {
	return fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=%v TimeZone=%v",
		g.host,
		g.user,
		g.password,
		g.dbname,
		g.port,
		g.timezone,
	)
}

func (g *GormPgAdapter) ensureDbConnection(ctx context.Context) error {

	if g.db == nil {
		if conErr := g.Connect(ctx); conErr != nil {
			return conErr
		}
	}

	return nil
}

func (g *GormPgAdapter) Connect(ctx context.Context) error {
	db, err := gorm.Open(postgres.Open(g.createDsnString()), &gorm.Config{})
	if err != nil {
		return err
	}
	g.db = db
	return nil
}

func (g *GormPgAdapter) Disconnect(ctx context.Context) error {
	if g.db == nil {
		return nil // Nothing to disconnect
	}

	// Get the underlying SQL DB
	sqlDB, err := g.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Close the database connection
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close the database connection: %w", err)
	}

	g.db = nil // Set to nil to indicate the connection is closed
	return nil
}

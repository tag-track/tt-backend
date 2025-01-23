package database

import (
	"Backend/internal/models"
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

func (g *GormPgAdapter) Migrate(ctx context.Context) error {
	if err := g.ensureDbConnection(ctx); err != nil {
		return err
	}

	if err := g.db.
		WithContext(ctx).
		AutoMigrate(
			&models.Entity{},
		); err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////
// DB Methods
////////////////////////////////////////////////

func (g *GormPgAdapter) CreateEntity(ctx context.Context, e *models.Entity) error {
	if err := g.ensureDbConnection(ctx); err != nil {
		return err
	}
	if res := g.db.WithContext(ctx).Create(e); res.Error != nil {
		return res.Error
	}
	return nil
}

// QueryTopLevel
// this method is to get top level entities with its direct children populated.
func (g *GormPgAdapter) QueryTopLevel(ctx context.Context) ([]*models.Entity, error) {
	if err := g.ensureDbConnection(ctx); err != nil {
		return nil, err
	}

	var entities []*models.Entity
	if err := g.db.
		Where("parent_id IS NULL").
		Preload("Children").
		Find(&entities).
		Error; err != nil {
		return nil, err
	}

	return entities, nil
}

func (g *GormPgAdapter) QueryById(ctx context.Context, id string) (*models.Entity, error) {
	if err := g.ensureDbConnection(ctx); err != nil {
		return nil, err
	}

	var entities models.Entity

	if err := g.db.
		Preload("Children").
		First(&entities, "id = ?", id).
		Error; err != nil {
		return nil, err
	}

	return &entities, nil
}

func (g *GormPgAdapter) QueryMultipleById(ctx context.Context, ids ...string) ([]*models.Entity, error) {
	if err := g.ensureDbConnection(ctx); err != nil {
		return nil, err
	}

	var entities []*models.Entity

	if err := g.db.
		Preload("Children").
		Where("id IN ?", ids).
		Find(&entities).
		Error; err != nil {
		return nil, err
	}

	return entities, nil

}

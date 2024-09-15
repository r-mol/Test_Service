package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"golang.yandex/hasql"
	"golang.yandex/hasql/checkers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

var DefaultClusterOptions = []hasql.ClusterOption{
	hasql.WithUpdateInterval(2 * time.Second),
	hasql.WithNodePicker(hasql.PickNodeRoundRobin()),
}

var DefaultTimeout = 30 * time.Second

type Client struct {
	cluster *hasql.Cluster
}

func NewClient(
	ctx context.Context, cfg *Config,
) (*Client, error) {
	connStrings := cfg.MakeConnStrings()
	hasqlNodes := make([]hasql.Node, 0, len(connStrings))

	for host, connString := range connStrings {
		connCfg, err := pgxpool.ParseConfig(connString)
		if err != nil {
			return nil, fmt.Errorf("error parsing config for host %s: %w", host, err)
		}

		connCfg.MaxConns = int32(cfg.MaxConn)
		connCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

		db := stdlib.OpenDB(*connCfg.ConnConfig)

		go func() {
			if err := db.Ping(); err != nil {
				log.Errorf("Postgres connection error for host %s: %s", host, err)
			} else {
				log.Infof("Successfully connected to host %s", host)
			}
		}()

		hasqlNodes = append(hasqlNodes, hasql.NewNode(host, db))
	}

	opts := DefaultClusterOptions
	cluster, err := hasql.NewCluster(hasqlNodes, checkers.PostgreSQL, opts...)
	if err != nil {
		return nil, fmt.Errorf("make hasql cluster: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()
	if _, err = cluster.WaitForPrimary(ctx); err != nil {
		return nil, fmt.Errorf("wait for premary: %w", err)
	}
	log.Infof("HI")
	client := &Client{
		cluster: cluster,
	}

	return client, nil
}

func (c *Client) Close() {
	if err := c.cluster.Close(); err != nil {
		log.Error("HASQL closing error: %s", err)
	}
}

func (c *Client) PrimaryGormDB() (*gorm.DB, error) {
	node := c.cluster.Primary()
	if node == nil {
		return nil, errors.New("postgres: primary node is unavailable")
	}

	return gormWrapper(node)
}

func (c *Client) StandbyPreferredGormDB() (*gorm.DB, error) {
	node := c.cluster.StandbyPreferred()
	if node == nil {
		return nil, errors.New("postgres: node is unavailable")
	}

	return gormWrapper(node)
}

func gormWrapper(node hasql.Node) (*gorm.DB, error) {
	db, err := gorm.Open(
		postgres.New(
			postgres.Config{
				Conn: node.DB(),
			},
		), &gorm.Config{
			SkipDefaultTransaction:                   true,
			PrepareStmt:                              false,
			AllowGlobalUpdate:                        true,
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	)
	if err != nil {
		return nil, err
	}
	db.Set("gorm:association_autocreate", false)
	db.Set("gorm:association_autoupdate", false)
	return db, nil
}

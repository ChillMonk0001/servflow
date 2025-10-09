package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Servflow/servflow/pkg/engine/actions"
	"github.com/Servflow/servflow/pkg/engine/integration"
	"github.com/google/uuid"
)

type storageIntegrations interface {
	integration.Integration
	Store(ctx context.Context, data map[string]interface{}, options map[string]string) error
}

type Config struct {
	IntegrationID     string                 `json:"integrationID"`
	Table             string                 `json:"table"`
	DatasourceOptions map[string]string      `json:"datasourceOptions"`
	Fields            map[string]interface{} `json:"fields"`
}

type Store struct {
	cfg *Config
	i   storageIntegrations
}

func (s *Store) Type() string {
	return "store"
}

func New(config Config) (*Store, error) {
	if config.IntegrationID == "" {
		return nil, errors.New("datasource is required")
	}
	if config.Table == "" {
		return nil, errors.New("table is required")
	}
	i, err := integration.GetIntegration(config.IntegrationID)
	if err != nil {
		return nil, err
	}

	u, ok := i.(storageIntegrations)
	if !ok {
		return nil, errors.New("integration does not implement storageIntegrations")
	}

	return &Store{
		cfg: &config,
		i:   u,
	}, nil
}

func (s *Store) Config() string {
	filtersStr, err := json.Marshal(s.cfg.Fields)
	if err != nil {
		return ""
	}
	return string(filtersStr)
}

func (s *Store) Execute(ctx context.Context, modifiedConfig string) (interface{}, error) {
	var item map[string]interface{}
	if err := json.Unmarshal([]byte(modifiedConfig), &item); err != nil {
		return "", nil
	}
	if item == nil {
		item = make(map[string]interface{})
	}

	_, ok := item["id"]
	if !ok {
		item["id"] = uuid.New().String()
	}
	err := s.i.Store(ctx, item, map[string]string{"collection": s.cfg.Table})
	if err != nil {
		return "", fmt.Errorf("error storing: %w", err)
	}
	return item, nil
}

func init() {
	if err := actions.RegisterAction("store", func(config json.RawMessage) (actions.ActionExecutable, error) {
		var cfg Config
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("error creating store action: %v", err)
		}
		return New(cfg)
	}); err != nil {
		panic(err)
	}
}

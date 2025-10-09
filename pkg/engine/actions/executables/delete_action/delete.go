//go:generate mockgen -source delete.go -destination delete_mock.go -package delete_action
package delete_action

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Servflow/servflow/pkg/engine/actions"
	"github.com/Servflow/servflow/pkg/engine/integration"
	"github.com/Servflow/servflow/pkg/engine/integration/integrations/filters"
)

type Delete struct {
	cfg               *Config
	deleteIntegration deleteImplementation
}

func (d *Delete) Config() string {
	filtersStr, err := json.Marshal(d.cfg.Filters)
	if err != nil {
		return ""
	}
	return string(filtersStr)
}

func (d *Delete) Type() string {
	return "delete"
}

type Config struct {
	IntegrationID     string            `json:"integrationID"`
	Filters           []filters.Filter  `json:"filters"`
	Table             string            `json:"table"`
	DatasourceOptions map[string]string `json:"datasourceOptions"`
}

type deleteImplementation interface {
	integration.Integration
	Delete(ctx context.Context, options map[string]string, filters ...filters.Filter) error
}

func New(config Config) (*Delete, error) {
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

	u, ok := i.(deleteImplementation)
	if !ok {
		return nil, errors.New("integration is not of type deleteImplementation")
	}
	return &Delete{
		cfg:               &config,
		deleteIntegration: u,
	}, nil
}

func (d *Delete) Execute(ctx context.Context, modifiedConfig string) (interface{}, error) {
	var filters []filters.Filter
	if err := json.Unmarshal([]byte(modifiedConfig), &filters); err != nil {
		return "", err
	}

	var ret interface{}
	err := d.deleteIntegration.Delete(ctx, map[string]string{"collection": d.cfg.Table}, filters...)
	if err != nil {
		return "", fmt.Errorf("delete with filters: %v", err)
	}
	return ret, nil
}

func init() {
	if err := actions.RegisterAction("delete", func(config json.RawMessage) (actions.ActionExecutable, error) {
		var cfg Config
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("error creating delete action: %v", err)
		}
		return New(cfg)
	}); err != nil {
		panic(err)
	}
}

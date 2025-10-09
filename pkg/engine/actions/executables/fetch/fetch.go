package fetch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Servflow/servflow/pkg/engine/actions"
	"github.com/Servflow/servflow/pkg/engine/integration"
	"github.com/Servflow/servflow/pkg/engine/integration/integrations/filters"
)

type Fetch struct {
	cfg               *Config
	fetchIntegrations fetchImplementation
}

func (f *Fetch) Type() string {
	return "fetch"
}

type fetchImplementation interface {
	integration.Integration
	Fetch(ctx context.Context, options map[string]string, filters ...filters.Filter) ([]map[string]interface{}, error)
}

type Config struct {
	IntegrationID     string            `json:"integrationID"`
	Filters           []filters.Filter  `json:"filters"`
	Table             string            `json:"table"`
	DatasourceOptions map[string]string `json:"datasourceOptions"`
	Single            bool              `json:"single"`
	ShouldFail        bool              `json:"shouldFail"`
}

func New(config Config) (*Fetch, error) {
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

	u, ok := i.(fetchImplementation)
	if !ok {
		return nil, errors.New("integration is not of type fetchImplementation")
	}
	return &Fetch{
		cfg:               &config,
		fetchIntegrations: u,
	}, nil
}

func (f *Fetch) Config() string {
	filtersStr, err := json.Marshal(f.cfg.Filters)
	if err != nil {
		return ""
	}
	return string(filtersStr)
}

func (f *Fetch) Execute(ctx context.Context, modifiedConfig string) (interface{}, error) {
	var filters []filters.Filter
	if err := json.Unmarshal([]byte(modifiedConfig), &filters); err != nil {
		return "", err
	}

	var ret interface{}
	resp, err := f.fetchIntegrations.Fetch(ctx, map[string]string{"collection": f.cfg.Table}, filters...)
	if err != nil {
		return "", fmt.Errorf("fetch with filters: %v", err)
	}
	ret = resp
	if len(resp) < 1 && !f.cfg.ShouldFail {
		return map[string]interface{}{}, nil
	} else if len(resp) < 1 && f.cfg.ShouldFail {
		return nil, fmt.Errorf("no data found")
	}
	if f.cfg.Single && len(resp) > 0 {
		ret = resp[0]
	}
	return ret, nil
}

func init() {
	if err := actions.RegisterAction("fetch", func(config json.RawMessage) (actions.ActionExecutable, error) {
		var cfg Config
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("error creating fetch action: %v", err)
		}
		return New(cfg)
	}); err != nil {
		panic(err)
	}
}

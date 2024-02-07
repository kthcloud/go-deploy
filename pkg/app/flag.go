package app

import (
	"context"
	"go-deploy/pkg/config"
	"go-deploy/pkg/workers/confirm"
	"go-deploy/pkg/workers/job_execute"
	"go-deploy/pkg/workers/logger"
	metricsWorker "go-deploy/pkg/workers/metrics"
	"go-deploy/pkg/workers/ping"
	"go-deploy/pkg/workers/repair"
	"go-deploy/pkg/workers/snapshot"
	"go-deploy/pkg/workers/status_update"
	"go-deploy/pkg/workers/synchronize"
)

// FlagDefinition represents a definition for a flag that is passed to the program's executable.
type FlagDefinition struct {
	Name         string
	ValueType    string
	FlagType     string
	Description  string
	DefaultValue interface{}
	PassedValue  interface{}
	Run          func(ctx context.Context, cancel context.CancelFunc)
}

// GetPassedValue returns the value passed to the flag.
func (flag *FlagDefinition) GetPassedValue() interface{} {
	return flag.PassedValue
}

type FlagDefinitionList []FlagDefinition

// IsPassed returns true if the flag was passed to the program.
func (list *FlagDefinitionList) IsPassed(name string) bool {
	for _, flag := range *list {
		if flag.Name == name {
			return flag.GetPassedValue() != interface{}(nil)
		}
	}

	return false
}

// GetPassedValue returns the value passed to the flag.
func (list *FlagDefinitionList) GetPassedValue(name string) interface{} {
	for _, flag := range *list {
		if flag.Name == name {
			return flag.GetPassedValue()
		}
	}

	return nil
}

// SetPassedValue sets the value passed to the flag.
func (list *FlagDefinitionList) SetPassedValue(name string, value interface{}) {
	for idx, flag := range *list {
		if flag.Name == name {
			(*list)[idx].PassedValue = value
			return
		}
	}
}

// AnyWorkerFlagsPassed returns true if any worker flags were passed to the program.
func (list *FlagDefinitionList) AnyWorkerFlagsPassed() bool {
	for _, flag := range *list {
		if flag.FlagType == "worker" && flag.GetPassedValue().(bool) {
			return true
		}
	}

	return false
}

// GetFlags returns a list of all flags that can be passed to the program.
func GetFlags() FlagDefinitionList {
	return []FlagDefinition{
		{
			Name:         "api",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start api server",
			DefaultValue: false,
			Run:          nil,
		},
		{
			Name:         "confirmer",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start confirmer",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				confirm.Setup(ctx)
			},
		},
		{
			Name:         "status-updater",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start status updater",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				status_update.Setup(ctx)
			},
		},
		{
			Name:         "synchronizer",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start synchronizer",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				synchronize.Setup(ctx)
			},
		},
		{
			Name:         "job-executor",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start job executor",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				job_execute.Setup(ctx)
			},
		},
		{
			Name:         "repairer",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start repairer",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				repair.Setup(ctx)
			},
		},
		{
			Name:         "pinger",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start pinger",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				ping.Setup(ctx)
			},
		},
		{
			Name:         "snapshotter",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start snapshotter",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				snapshot.Setup(ctx)
			},
		},
		{
			Name:         "metrics-updater",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start metrics updater",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				metricsWorker.Setup(ctx)
			},
		},
		{
			Name:         "logger",
			ValueType:    "bool",
			FlagType:     "worker",
			Description:  "start logger",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				logger.Setup(ctx)
			},
		},
		{
			Name:         "test-mode",
			ValueType:    "bool",
			FlagType:     "global",
			Description:  "run in test mode",
			DefaultValue: false,
			Run: func(ctx context.Context, _ context.CancelFunc) {
				config.Config.TestMode = true
				config.Config.MongoDB.Name = config.Config.MongoDB.Name + "-test"
			},
		},
	}
}

// Copyright 2019 Intel Corporation and Smart-Edge.com, Inc. All rights reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/smartedgemec/appliance-ce/pkg/config"
	"github.com/smartedgemec/appliance-ce/pkg/eaa"
	"github.com/smartedgemec/appliance-ce/pkg/logger"

	// Imports required to run agents
	"github.com/smartedgemec/appliance-ce/pkg/ela"
)

// ServiceStartFunction is func typedef for starting service
type ServiceStartFunction func(context.Context, string) error

// EdgeServices array contains function pointers to services start functions
var EdgeServices = []ServiceStartFunction{ela.Run, eaa.Run}

var log = logger.NewLogger("main")

var cfg mainConfig

type mainConfig struct {
	Log      logger.Config     `json:"log"`
	Services map[string]string `json:"services"`
}

func init() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "configs/appliance.json",
		"config file path")
	flag.Parse()

	err := config.LoadJSONConfig(cfgPath, &cfg)
	if err != nil {
		log.Errorf("Failed to load config: %s", err.Error())
		os.Exit(1)
	}

	err = logger.ConfigLogger(log, &cfg.Log)
	if err != nil {
		log.Errorf("Failed to configure the logger: %s", err.Error())
		os.Exit(1)
	}
}

func waitForServices(wg *sync.WaitGroup,
	errors <-chan error, cancel context.CancelFunc) bool {

	waitFinished := make(chan struct{})
	ret := true
	go func() {
		defer close(waitFinished)
		wg.Wait()
	}()

	for {
		select {
		case <-waitFinished:
			return ret
		case err := <-errors:
			if err != nil {
				log.Tracef("Cancelling services because of error"+
					" from one of the services: %#v", err)
				cancel()
				ret = false
			}
		}
	}
}

func runServices(services []ServiceStartFunction) bool {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Handle SIGINT and SIGTERM by calling cancel()
	// which is propagated to services
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-osSignals
		log.Infof("Received signal: %#v", sig)
		cancel()
	}()

	results := make(chan error)

	log.Infof("Starting services")
	for _, runner := range services {
		funcName := runtime.FuncForPC(reflect.ValueOf(runner).Pointer()).Name()
		srvName := funcName[:strings.LastIndex(funcName, ".")]

		log.Infof("Starting: %v", srvName)
		wg.Add(1)
		go func(wg *sync.WaitGroup, start ServiceStartFunction, cfg string) {
			defer wg.Done()
			err := start(ctx, cfg)
			results <- err
		}(&wg, runner, cfg.Services[srvName])
	}

	return waitForServices(&wg, results, cancel)
}

func main() {
	if !runServices(EdgeServices) {
		os.Exit(1)
	}

	log.Infof("Services stopped gracefully")
}

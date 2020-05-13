// -*- coding: utf-8 -*-
//
// Copyright (c) 2020, Christian E. Hopps
// All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"flag"
	"github.com/choppsv1/docker-network-p2p/driver"
	"github.com/choppsv1/docker-network-p2p/logging"
	"github.com/docker/go-plugins-helpers/network"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var Version = "unset"

func setupSignalHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logging.Info("Exiting due to signal")
		os.Exit(1)
	}()
}

func main() {

	debugEnv, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	debugPtr := flag.Bool("debug", false, "Enable debug logging")
	tracePtr := flag.Bool("trace", false, "Enable debug logging")
	flag.Parse()

	logging.GlbDebug = true || (*debugPtr || debugEnv)
	logging.GlbTrace = *tracePtr

	setupSignalHandler()

	logging.Info("Initializing p2p network driver: version %v", Version)
	d, err := driver.Init()
	if err != nil {
		logging.Panicf("Failed to start driver: %v", err)
	}

	logging.Debug("Registering with docker")
	h := network.NewHandler(d)
	if err = h.ServeUnix("p2p", 0); err != nil {
		logging.Panicf("Error during execution: %v", err)
	}

	logging.Info("Exiting p2p network driver")
}

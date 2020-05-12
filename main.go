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
	"github.com/choppsv1/p2p-network-driver/driver"
	"github.com/choppsv1/p2p-network-driver/logging"
	"github.com/docker/go-plugins-helpers/network"
	"os"
	"os/signal"
	"syscall"
)

const (
	version = "0.1.0"
)

func cleanup() {
	logging.Info("Exiting due to signal")
}

func main() {
	debugPtr := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	logging.GlbDebug = *debugPtr
	logging.GlbTrace = *debugPtr

	logging.Info("Starting p2p network driver version %v", version)

	d, err := driver.Init()
	if err != nil {
		logging.Panicf("Failed to start driver: %v", err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	logging.Debug("Registering with docker")

	h := network.NewHandler(d)
	if err = h.ServeUnix("p2p", 0); err != nil {
		logging.Panicf("Error during execution: %v", err)
	}

	logging.Info("Exiting p2p-network-driver driver")
}

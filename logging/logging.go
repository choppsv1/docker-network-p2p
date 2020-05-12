// -*- coding: utf-8 -*-
//
// May 12 2020, Christian E. Hopps <chopps@gmail.com>
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

package logging

import (
	"log"
	"os"
)

var traplogger = log.New(os.Stderr, "TRAP: P2P: ", log.Ldate|log.Ltime|log.Lmicroseconds)
var tlogger = log.New(os.Stderr, "TRACE: P2P: ", log.Ldate|log.Ltime|log.Lmicroseconds)
var dlogger = log.New(os.Stderr, "DEBUG: P2P: ", log.Ldate|log.Ltime|log.Lmicroseconds)
var logger = log.New(os.Stderr, "INFO: P2P: ", log.Ldate|log.Ltime|log.Lmicroseconds)
var wlogger = log.New(os.Stderr, "WARNING: P2P: ", log.Ldate|log.Ltime|log.Lmicroseconds)
var elogger = log.New(os.Stderr, "ERROR: P2P: ", log.Ldate|log.Ltime|log.Lmicroseconds)

var GlbDebug, GlbTrace bool

// Trace logs to the tracing logger if the given trace flag is set.
func Trace(format string, a ...interface{}) {
	if GlbTrace {
		tlogger.Printf(format, a...)
	}
}

// Trace logs to the tracing logger if the given trace flag is set.
func Debug(format string, a ...interface{}) {
	if GlbDebug {
		dlogger.Printf(format, a...)
	}
}

// Info logs to the info logger unconditionally.
func Info(format string, a ...interface{}) {
	logger.Printf(format, a...)
}

// Info logs to the info logger unconditionally.
func Warn(format string, a ...interface{}) {
	wlogger.Printf(format, a...)
}

// Info logs to the info logger unconditionally.
func Err(format string, a ...interface{}) {
	elogger.Printf(format, a...)
}

// Trap logs to the trap logger unconditionally.
func Trap(format string, a ...interface{}) {
	traplogger.Printf(format, a...)
}

// Panicf panics using the info logger and format string and args.
func Panicf(format string, a ...interface{}) {
	logger.Panicf(format, a...)
}

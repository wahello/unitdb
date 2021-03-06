/*
 * Copyright 2020 Saffat Technologies, Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package memdb

import (
	"time"
)

type _Options struct {
	logFilePath string

	// memdbSize sets maximum size of DB.
	memdbSize int64

	// bufferSize sets size of buffer to use for buffer pooling.
	bufferSize int64

	// logResetFlag flag to skips log recovery on DB open and reset WAL.
	logResetFlag bool

	logInterval time.Duration

	timeBlockDuration time.Duration
}

// Options it contains configurable options and flags for DB.
type Options interface {
	set(*_Options)
}

// fOption wraps a function that modifies options and flags into an
// implementation of the Options interface.
type fOption struct {
	f func(*_Options)
}

func (fo *fOption) set(o *_Options) {
	fo.f(o)
}

func newFuncOption(f func(*_Options)) *fOption {
	return &fOption{
		f: f,
	}
}

// WithDefaultOptions will open DB with some default values.
func WithDefaultOptions() Options {
	return newFuncOption(func(o *_Options) {
		if o.logFilePath == "" {
			o.logFilePath = "/tmp/unitdb"
		}
		if o.memdbSize == 0 {
			o.memdbSize = defaultMemdbSize
		}
		if o.bufferSize == 0 {
			o.bufferSize = defaultBufferSize
		}
		if o.logInterval == 0 {
			o.logInterval = 15 * time.Millisecond
		}
		if o.timeBlockDuration == 0 {
			o.timeBlockDuration = 1 * time.Second
		}
	})
}

// WithLogFilePath sets database directory for storing logs.
func WithLogFilePath(path string) Options {
	return newFuncOption(func(o *_Options) {
		o.logFilePath = path
	})
}

// WithMemdbSize sets max size of DB.
func WithMemdbSize(size int64) Options {
	return newFuncOption(func(o *_Options) {
		o.memdbSize = size
	})
}

// WithBufferSize sets max size of buffer to use for buffer pooling.
func WithBufferSize(size int64) Options {
	return newFuncOption(func(o *_Options) {
		o.bufferSize = size
	})
}

// WithLogReset flag to skip recovery on DB open and reset WAL.
func WithLogReset() Options {
	return newFuncOption(func(o *_Options) {
		o.logResetFlag = true
	})
}

// WithLogInterval sets interval for a time block. Block is pushed to the queue to write it to the log file.
func WithLogInterval(dur time.Duration) Options {
	return newFuncOption(func(o *_Options) {
		o.logInterval = dur
	})
}

// WithTimeBlockInterval sets interval for a time block. Block is pushed to the queue to write it to the log file.
func WithTimeBlockInterval(dur time.Duration) Options {
	return newFuncOption(func(o *_Options) {
		o.timeBlockDuration = dur
	})
}

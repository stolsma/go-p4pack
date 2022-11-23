// SPDX-FileCopyrightText: 2020-2022 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestDefaultLogger(t *testing.T) {
	Configure(&Config{})
	logger := GetLogger()
	assert.Equal(t, "github.com/stolsma/go-p4pack/pkg/logging", logger.Name())
	logger.Info("foo: bar")
	logger.Info("bar: baz")
	logger.Debug("baz: bar")
}

func TestLoggerConfig(t *testing.T) {
	// do json config tests
	jconfig := Config{}
	err := json.Unmarshal([]byte(jsonConfig), &jconfig)
	assert.NoError(t, err)
	err = Configure(&jconfig)
	assert.NoError(t, err)
	doTests(t)

	// do yaml config tests
	yconfig := Config{}
	err = yaml.Unmarshal([]byte(yamlConfig), &yconfig)
	assert.NoError(t, err)
	err = Configure(&yconfig)
	assert.NoError(t, err)
	doTests(t)
}

func doTests(t *testing.T) {
	// The root logger should be configured with INFO level
	logger := GetLogger()
	assert.Equal(t, InfoLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should be printed")

	// The "test" logger should inherit the INFO level from the root logger
	logger = GetLogger("test")
	assert.Equal(t, InfoLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should be printed")

	// The "test/1" logger should be configured with DEBUG level
	logger = GetLogger("test", "1")
	assert.Equal(t, DebugLevel, logger.GetLevel())
	logger.Debug("should be printed")
	logger.Info("should be printed")

	// The "test/1/2" logger should inherit the DEBUG level from "test/1"
	logger = GetLogger("test", "1", "2")
	assert.Equal(t, DebugLevel, logger.GetLevel())
	logger.Debug("should be printed")
	logger.Info("should be printed")

	// The "test" logger should still inherit the INFO level from the root logger
	logger = GetLogger("test")
	assert.Equal(t, InfoLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should be printed")

	// The "test/2" logger should be configured with WARN level
	logger = GetLogger("test", "2")
	assert.Equal(t, WarnLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should not be printed")
	logger.Warn("should be printed twice")

	// The "test/2/3" logger should be configured with INFO level
	logger = GetLogger("test", "2", "3")
	assert.Equal(t, InfoLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should be printed twice")
	logger.Warn("should be printed twice")

	// The "test/2/4" logger should inherit the WARN level from "test/2"
	logger = GetLogger("test", "2", "4")
	assert.Equal(t, WarnLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should not be printed")
	logger.Warn("should be printed twice")

	// The "test/2" logger level should be changed to DEBUG
	logger = GetLogger("test/2")
	logger.SetLevel(DebugLevel)
	assert.Equal(t, DebugLevel, logger.GetLevel())
	logger.Debug("should be printed")
	logger.Info("should be printed twice")
	logger.Warn("should be printed twice")

	// The "test/2/3" logger should not inherit the change to the "test/2" logger since its level has been explicitly set
	logger = GetLogger("test/2/3")
	assert.Equal(t, InfoLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should be printed twice")
	logger.Warn("should be printed twice")

	// The "test/2/4" logger should inherit the change to the "test/2" logger since its level has not been explicitly set
	// The "test/2/4" logger should not output DEBUG messages since the output level is explicitly set to WARN
	logger = GetLogger("test/2/4")
	assert.Equal(t, DebugLevel, logger.GetLevel())
	logger.Debug("should be printed")
	logger.Info("should be printed twice")
	logger.Warn("should be printed twice")

	// The "test/3" logger should be configured with INFO level
	// The "test/3" logger should write to multiple outputs
	logger = GetLogger("test/3")
	assert.Equal(t, InfoLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should be printed")
	logger.Warn("should be printed twice")

	// The "test/3/4" logger should inherit INFO level from "test/3"
	// The "test/3/4" logger should inherit multiple outputs from "test/3"
	logger = GetLogger("test/3/4")
	assert.Equal(t, InfoLevel, logger.GetLevel())
	logger.Debug("should not be printed")
	logger.Info("should be printed")
	logger.Warn("should be printed twice")

	// logger = GetLogger("test", "kafka")
	// assert.Equal(t, InfoLevel, logger.GetLevel())
}

const jsonConfig = `{
"loggers": {
  "root": {
    "level": "info",
    "output": {
      "stdout": {
        "sink": "stdout"
      },
      "file": {
        "sink": "file"
      }
    }
  },
  "test/1": {
    "level": "debug"
  },
  "test/2": {
    "level": "warn",
    "output": {
      "stdout-1": {
        "sink": "stdout-1",
        "level": "info"
      }
    }
  },
  "test/2/3": {
    "level": "info"
  },
  "test/3": {
    "level": "info",
    "output": {
      "stdout": {
        "level": "info"
      },
      "stdout-1": {
        "sink": "stdout-1",
        "level": "warn"
      }
    }
  },
  "test/kafka": {
    "level": "info",
    "output": {
      "kafka": {
        "sink": "kafka-1"
      }
    }
  }
},
"sinks": {
  "stdout": {
    "type": "stdout",
    "encoding": "console",
    "stdout": {}
  },
  "kafka-1": {
    "type": "kafka",
    "encoding": "json",
    "kafka": {
      "brokers": [
        "127.0.0.1:9092"
      ],
      "topic": "traces",
      "key": "test"
    }
  },
  "stdout-1": {
    "type": "stdout",
    "encoding": "json",
    "stdout": {}
  },
  "file": {
    "type": "file",
    "encoding": "json",	
    "file": {
      "path": "test.log"
    }
  }
}
}`

const yamlConfig = `loggers:
  root:
    level: info
    output:
      stdout:
        sink: stdout
      file:
        sink: file
  test/1:
    level: debug
  test/2:
    level: warn
    output:
      stdout-1:
        sink: stdout-1
        level: info
  test/2/3:
    level: info
  test/3:
    level: info
    output:
      stdout:
        level: info
      stdout-1:
        sink: stdout-1
        level: warn
  test/kafka:
    level: info
    output:
      kafka:
        sink: kafka-1
sinks:
  stdout:
    type: stdout
    encoding: console
    stdout: {}
  kafka-1:
    type: kafka
    encoding: json
    kafka:
      brokers:
        - "127.0.0.1:9092"
      topic: "traces"
      key: "test"
  stdout-1:
    type: stdout
    encoding: json
    stdout: {}
  file:
    type: file
    encoding: json
    file:
      path: test.log
`

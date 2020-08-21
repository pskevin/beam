// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	"github.com/apache/beam/sdks/go/pkg/beam/core/util/reflectx"
	"github.com/apache/beam/sdks/go/pkg/beam/testing/passert"

	"context"
	"flag"
	"log"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/core/typex"

	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"

	// Imports to enable correct filesystem access and runner setup in LOOPBACK mode
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
	_ "github.com/apache/beam/sdks/go/pkg/beam/runners/universal"
)

var (
	expansionAddr = flag.String("expansion_addr", "", "Address of Expansion Service")
)

// formatFn is a DoFn that formats a word and its count as a string.
func formatFn(c int64) string {
	return fmt.Sprintf("%v", c)
}

func init() {
	beam.RegisterFunction(formatFn)
}

func main() {
	flag.Parse()
	beam.Init()

	if *expansionAddr == "" {
		log.Fatal("No expansion address provided")
	}

	p := beam.NewPipeline()
	s := p.Root()

	col := beam.CreateList(s, []int64{1, 2, 3, 4, 5, 6})

	outputType := typex.New(reflectx.Int64)
	namedOutputs := map[string]typex.FullType{"0": outputType, "1": outputType}

	// Using Cross-language Count from Python's test expansion service

	c := beam.CrossLanguageWithSource(s, "beam:transforms:xlang:test:partition", nil, *expansionAddr, col, namedOutputs)

	formatted0 := beam.ParDo(s, formatFn, c["0"])
	formatted1 := beam.ParDo(s, formatFn, c["1"])

	passert.Equals(s, formatted0, "2", "4", "6")
	passert.Equals(s, formatted1, "1", "3", "5")

	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}

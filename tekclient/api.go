// Copyright 2024 The Tektite Authors
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

package tekclient

import (
	"github.com/spirit-labs/tektite/types"
)

type Client interface {
	ExecuteStatement(statement string) error

	PrepareQuery(queryName string, tsl string) (PreparedQuery, error)

	GetPreparedQuery(queryName string) PreparedQuery

	ExecuteQuery(query string) (QueryResult, error)

	StreamExecuteQuery(query string) (chan StreamChunk, error)

	RegisterWasmModule(modulePath string) error

	UnregisterWasmModule(moduleName string) error

	Close()
}

type PreparedQuery interface {
	Name() string

	SetIntArg(index int, val int64)

	SetFloatArg(index int, val float64)

	SetBoolArg(index int, val bool)

	SetDecimalArg(index int, val types.Decimal)

	SetStringArg(index int, val string)

	SetBytesArg(index int, val []byte)

	SetTimestampArg(index int, val types.Timestamp)

	SetNullArg(index int)

	Execute() (QueryResult, error)

	StreamExecute() (chan StreamChunk, error)
}

type QueryResult interface {
	Meta() Meta

	ColumnCount() int

	RowCount() int

	Column(colIndex int) Column

	Row(rowIndex int) Row
}

type StreamChunk struct {
	Err error

	Chunk QueryResult
}

type Column interface {
	IsNull(rowIndex int) bool

	IntVal(rowIndex int) int64

	FloatVal(rowIndex int) float64

	BoolVal(rowIndex int) bool

	DecimalVal(rowIndex int) types.Decimal

	StringVal(rowIndex int) string

	BytesVal(rowIndex int) []byte

	TimestampVal(rowIndex int) types.Timestamp
}

type Row interface {
	IsNull(colIndex int) bool

	IntVal(colIndex int) int64

	FloatVal(colIndex int) float64

	BoolVal(colIndex int) bool

	DecimalVal(colIndex int) types.Decimal

	StringVal(colIndex int) string

	BytesVal(colIndex int) []byte

	TimestampVal(colIndex int) types.Timestamp
}

type Meta interface {
	ColumnNames() []string

	ColumnTypes() []types.ColumnType
}

type ResultChunk struct {
	QueryResult

	Err error
}

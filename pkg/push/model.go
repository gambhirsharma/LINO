// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package push

import (
	"encoding/json"
	"time"

	over "github.com/adrienaury/zeromdc"
	"github.com/rs/zerolog/log"
)

// Table from which to push data.
type Table interface {
	Name() string
	PrimaryKey() []string
	Columns() ColumnList
	Import(map[string]interface{}) (ImportedRow, *Error)
	GetColumn(name string) Column
}

// ColumnList is a list of columns.
type ColumnList interface {
	Len() uint
	Column(idx uint) Column
}

// Column of a table.
type Column interface {
	Name() string
	Export() string
	Import() string
	Length() int64
	LengthInBytes() bool
	Truncate() bool
	Preserve() string
}

// Plan describe how to push data
type Plan interface {
	FirstTable() Table
	RelationsFromTable(table Table) map[string]Relation
	Tables() []Table
}

// Relation between two tables.
type Relation interface {
	Name() string
	Parent() Table
	Child() Table
	OppositeOf(table Table) Table
}

// Value is an untyped data.
type Value interface{}

// Row of data.
type Row map[string]interface{}

// Cache is a dictionary for values.
type Cache map[Value]Value

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}

// StopIteratorError signal the end of iterator
type StopIteratorError struct{}

// ExecutionStats provides an overview of the work done
type ExecutionStats interface {
	GetInputLinesCount() int
	GetCreatedLinesCount() map[string]int
	GetDeletedLinesCount() map[string]int
	GetCommitsCount() int
	GetDuration() time.Duration

	ToJSON() []byte
}

type stats struct {
	InputLinesCount   int            `json:"inputLinesCount"`
	CreatedLinesCount map[string]int `json:"createdLinesCount"`
	DeletedLinesCount map[string]int `json:"deletedLinesCount"`
	CommitsCount      int            `json:"commitsCount"`
	Duration          time.Duration  `json:"duration"`
}

// Reset all statistics to zero
func Reset() {
	over.MDC().Set("stats", &stats{CreatedLinesCount: map[string]int{}, DeletedLinesCount: map[string]int{}})
}

// Compute current statistics and give a snapshot
func Compute() ExecutionStats {
	value, exists := over.MDC().Get("stats")
	if stats, ok := value.(ExecutionStats); exists && ok {
		return stats
	}
	log.Warn().Msg("Unable to compute statistics")
	return &stats{}
}

func (s *stats) ToJSON() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		log.Warn().Msg("Unable to read statistics")
	}
	return b
}

func (s *stats) GetCreatedLinesCount() map[string]int {
	return s.CreatedLinesCount
}

func (s *stats) GetInputLinesCount() int {
	return s.InputLinesCount
}

func (s *stats) GetDeletedLinesCount() map[string]int {
	return s.DeletedLinesCount
}

func (s *stats) GetCommitsCount() int {
	return s.CommitsCount
}

func (s *stats) GetDuration() time.Duration {
	return s.Duration
}

func IncCreatedLinesCount(table string) {
	stats := getStats()
	stats.CreatedLinesCount[table]++
}

func IncInputLinesCount() {
	stats := getStats()
	stats.InputLinesCount++
}

func IncCommitsCount() {
	stats := getStats()
	stats.CommitsCount++
}

func IncDeletedLinesCount(table string) {
	stats := getStats()
	stats.DeletedLinesCount[table]++
}

func SetDuration(duration time.Duration) {
	stats := getStats()
	stats.Duration = duration
}

func getStats() *stats {
	value, exists := over.MDC().Get("stats")
	if stats, ok := value.(*stats); exists && ok {
		return stats
	}
	log.Warn().Msg("Statistics uncorrectly initialized")
	return &stats{}
}

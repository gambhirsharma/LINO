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

package table

import (
	"bytes"
	"os"

	"github.com/cgi-fr/lino/pkg/table"
	"gopkg.in/yaml.v3"
)

// Version of the YAML strcuture.
const Version string = "v1"

// YAMLStructure of the file.
type YAMLStructure struct {
	Version string      `yaml:"version"`
	Tables  []YAMLTable `yaml:"tables,omitempty"`
}

// YAMLTable defines how to store a table in YAML format.
type YAMLTable struct {
	Name       string       `yaml:"name"`
	Keys       []string     `yaml:"keys"`
	Columns    []YAMLColumn `yaml:"columns,omitempty"`
	ExportMode string       `yaml:"export,omitempty"`
}

// YAMLColumn defines how to store a column in YAML format.
type YAMLColumn struct {
	Name     string     `yaml:"name"`
	Export   string     `yaml:"export,omitempty"`
	Import   string     `yaml:"import,omitempty"`
	DBInfo   YAMLDBInfo `yaml:"dbinfo,omitempty"`
	Preserve string     `yaml:"preserve,omitempty"`
}

// YAMLDBInfo defines how to store a column dbinfos in YAML format.
type YAMLDBInfo struct {
	Type      string `yaml:"type"`
	Length    int64  `yaml:"length,omitempty"`
	Size      int64  `yaml:"size,omitempty"`
	Precision int64  `yaml:"precision,omitempty"`
	ByteBased bool   `yaml:"bytes,omitempty"`
}

// YAMLStorage provides storage in a local YAML file
type YAMLStorage struct{}

// NewYAMLStorage create a new YAML storage
func NewYAMLStorage() *YAMLStorage {
	return &YAMLStorage{}
}

// List all tables stored in the YAML file
func (s YAMLStorage) List() ([]table.Table, *table.Error) {
	list, err := readFile()
	if err != nil {
		return nil, err
	}
	result := []table.Table{}

	for _, ym := range list.Tables {
		cols := []table.Column{}
		for _, ymc := range ym.Columns {
			cols = append(cols, table.Column{Name: ymc.Name, Export: ymc.Export, Import: ymc.Import, DBInfo: table.DBInfo(ymc.DBInfo), Preserve: ymc.Preserve})
		}

		exportMode := table.ExportModeOnly
		if ym.ExportMode == "all" {
			exportMode = table.ExportModeAll
		}

		m := table.Table{
			Name:       ym.Name,
			Keys:       ym.Keys,
			Columns:    cols,
			ExportMode: exportMode,
		}
		result = append(result, m)
	}

	return result, nil
}

// Store tables in the YAML file
func (s YAMLStorage) Store(tables []table.Table) *table.Error {
	list := YAMLStructure{
		Version: Version,
	}

	for _, r := range tables {
		cols := []YAMLColumn{}
		for _, rc := range r.Columns {
			cols = append(cols, YAMLColumn{Name: rc.Name, Export: rc.Export, Import: rc.Import, DBInfo: YAMLDBInfo(rc.DBInfo)})
		}

		yml := YAMLTable{
			Name:    r.Name,
			Keys:    r.Keys,
			Columns: cols,
		}

		list.Tables = append(list.Tables, yml)
	}

	if err := writeFile(&list); err != nil {
		return err
	}

	return nil
}

func readFile() (*YAMLStructure, *table.Error) {
	list := &YAMLStructure{
		Version: Version,
	}

	dat, err := os.ReadFile("tables.yaml")
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	err = yaml.Unmarshal(dat, list)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	if list.Version != Version {
		return nil, &table.Error{Description: "invalid version in ./tables.yaml (" + list.Version + ")"}
	}

	return list, nil
}

func writeFile(list *YAMLStructure) *table.Error {
	out := &bytes.Buffer{}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)

	err := enc.Encode(list)
	if err != nil {
		return &table.Error{Description: err.Error()}
	}

	err = os.WriteFile("tables.yaml", out.Bytes(), 0o600)
	if err != nil {
		return &table.Error{Description: err.Error()}
	}

	return nil
}

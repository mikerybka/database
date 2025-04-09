package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mikerybka/util"
)

type Server struct {
	Workdir string
	locks   map[string]*sync.Mutex
}

func (s *Server) writeFile(path, data string) error {
	path = filepath.Join(s.Workdir, path)
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(data), os.ModePerm)
}

func (s *Server) CreateDB(name string) error {
	err := os.MkdirAll(filepath.Join(s.Workdir, "config", name), os.ModePerm)
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(s.Workdir, "data", name), os.ModePerm)
}

func (s *Server) DeleteDB(name string) error {
	err := os.RemoveAll(filepath.Join(s.Workdir, "config", name))
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(s.Workdir, "data", name))
}

func (s *Server) CreateTable(dbName, tableName string) error {
	err := os.MkdirAll(filepath.Join(s.Workdir, "config", dbName, tableName), os.ModePerm)
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(s.Workdir, "data", dbName, tableName), os.ModePerm)
}
func (s *Server) DeleteTable(dbName, tableName string) error {
	err := os.RemoveAll(filepath.Join(s.Workdir, "config", dbName, tableName))
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(s.Workdir, "data", dbName, tableName))
}

func (s *Server) CreateColumn(dbName, tableName, columnName, columnType string) error {
	return os.WriteFile(filepath.Join(s.Workdir, "config", dbName, tableName, columnName), []byte(columnType), os.ModePerm)
}

func (s *Server) DeleteColumn(dbName, tableName, columnName string) error {
	mu, ok := s.locks[fmt.Sprintf("/%s/%s", dbName, tableName)]
	if !ok {
		mu = &sync.Mutex{}
	}
	mu.Lock()
	defer mu.Unlock()

	err := os.Remove(filepath.Join(s.Workdir, "config", dbName, tableName, columnName))
	if err != nil {
		return err
	}

	rows, err := s.ListAllRows(dbName, tableName)
	if err != nil {
		return err
	}
	for _, rowID := range rows {
		err = os.Remove(filepath.Join(s.Workdir, "data", dbName, tableName, rowID, columnName))
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *Server) AddRow(db, table string, columns map[string]string) (rowID string, err error) {
	mu, ok := s.locks[fmt.Sprintf("/%s/%s", db, table)]
	if !ok {
		mu = &sync.Mutex{}
	}
	mu.Lock()
	defer mu.Unlock()

	rowID, err = s.newID(db, table)
	if err != nil {
		return "", err
	}
	for k, v := range columns {
		err := s.writeFile(filepath.Join("data", db, table, k), v)
		if err != nil {
			return rowID, err
		}
	}
	return rowID, nil
}

func (s *Server) newID(db, table string) (id string, err error) {
	id = util.RandomID()
	for {
		_, err := os.Stat(filepath.Join(s.Workdir, "data", db, table, id))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return id, nil
			}
			return "", err
		} else {
			id = util.RandomID()
		}
	}
}
func (s *Server) UpdateRow(db, table, row string, columns map[string]string) error {
	mu, ok := s.locks[fmt.Sprintf("/%s/%s/%s", db, table, row)]
	if !ok {
		mu = &sync.Mutex{}
	}
	mu.Lock()
	defer mu.Unlock()

}
func (s *Server) DeleteRow(db, table, row string) error {
	mu, ok := s.locks[fmt.Sprintf("/%s/%s", db, table)]
	if !ok {
		mu = &sync.Mutex{}
	}
	mu.Lock()
	defer mu.Unlock()

}
func (s *Server) GetRowByID(dbID, tableID, rowID string) (row map[string]string, err error)
func (s *Server) ListAllRows(dbID, tableID string) (rowIDs []string, err error)
func (s *Server) ListRowsWhere(db, table, where string) (rowIDs []string, err error)
func (s *Server) ListDBs() (dbIDs []string, err error)
func (s *Server) ListTables(dbID string) (tableIDs []string, err error)
func (s *Server) ListColumns(dbID, tableID string) (columnIDs []string, err error)
func (s *Server) GetColumnType(dbID, tableID, columnID string) (t string, err error)

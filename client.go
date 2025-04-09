package database

func NewClient() *Client {
	return &Client{}
}

type Client struct {
	ServerAddr string
}

func (c *Client) CreateDB(name string) error
func (c *Client) DeleteDB(name string) error
func (c *Client) CreateTable(dbName, tableName string) error
func (c *Client) DeleteTable(dbName, tableName string) error
func (c *Client) CreateColumn(dbName, tableName, columnName, columnType string) error
func (c *Client) DeleteColumn(dbName, tableName, columnName string) error
func (c *Client) AddRow(db, table string, columns map[string]string) (string, error)
func (c *Client) UpdateRow(db, table, row string, columns map[string]string) error
func (c *Client) DeleteRow(db, table, row string) error
func (c *Client) GetRowByID(db, table, row string) (map[string]string, error)
func (c *Client) ListRowsWhere(db, table, where string) ([]string, error)
func (c *Client) ListDBs() ([]string, error)
func (c *Client) ListTables(db string) ([]string, error)
func (c *Client) ListColumns(db, table string) ([]string, error)

// func (c *Client) GetColumnType(db, table, column string) (string, error)

package addreality

import (
	"bytes"
	"database/sql"
	"errors"
	"html/template"
	"math"
)

const (
	PgSQLMaxLine   = math.MaxInt64
	PgSQLMaxParams = 65000
	MSSQLMaxLine   = 2100
	MSSQLMaxParams = math.MaxInt64

	MSSQLDriver = iota
	PgSQLDriver
)

var (
	ErrTooManyLineParams = errors.New("too many params in one line")
	ErrInvalidDriver     = errors.New("invalid driver type")
)

type BatchQuery struct {
	Query string
	Args  []interface{}
}

type InsertBuilder interface {
	Append(args ...interface{}) error
	ToSQL() ([]BatchQuery, error)
	GetMaxLine() int
	GetMaxParams() int
}

func NewInsertBuilder(driver int) (InsertBuilder, error) {
	var builder = &Builder{}

	switch driver {
	case MSSQLDriver:
		builder.MaxParams = MSSQLMaxParams
		builder.MaxLine = MSSQLMaxLine
	case PgSQLDriver:
		builder.MaxParams = PgSQLMaxParams
		builder.MaxLine = PgSQLMaxLine
	default:
		return nil, ErrInvalidDriver
	}

	builder.bqLen = 1

	return builder, nil
}

type Builder struct {
	MaxLine        int
	MaxParams      int
	curParamsCount int
	curLineCount   int
	bqLen          int
	BatchQueries   []BatchQuery
}

func (b *Builder) GetMaxLine() int {
	return b.MaxLine
}

func (b *Builder) GetMaxParams() int {
	return b.MaxParams
}

func (b *Builder) Append(args ...interface{}) error {
	if len(args) > b.MaxParams {
		return ErrTooManyLineParams
	}

	if (b.curParamsCount+len(args) > b.MaxParams) || (b.curLineCount == b.MaxLine) {
		b.curLineCount = 0
		b.curParamsCount = 0
		b.bqLen++
		b.BatchQueries = append(b.BatchQueries, BatchQuery{})
	}

	if b.BatchQueries == nil {
		b.BatchQueries = make([]BatchQuery, 1)
	}

	curbq := b.BatchQueries[b.bqLen-1]
	curbq.Args = append(curbq.Args, args)

	b.BatchQueries[b.bqLen-1] = curbq
	b.curLineCount++
	b.curParamsCount += len(args)

	return nil
}

func (b *Builder) ToSQL() ([]BatchQuery, error) {
	pattern := `
{{.Args}}
{{end}}
`

	tmpl, err := template.New("query").Parse(pattern)
	if err != nil {
		return nil, err
	}

	var buffer = bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, struct{ Args []interface{} }{args})
	if err != nil {
		return nil, err
	}

	return b.BatchQueries, nil
}

func BulkDevice(db *sql.DB) error {
	// rows := []struct {
	// 	Name       string
	// 	GroupID    uint
	// 	PlatformID uint
	// }{
	// 	{Name: "device1", GroupID: 1, PlatformID: 5281},
	// 	{Name: "device2", GroupID: 1, PlatformID: 5281},
	// 	{Name: "device3", GroupID: 1, PlatformID: 5281},
	// }
	//
	// var b InsertBuilder
	//
	// for _, r := range rows {
	// 	b.Append(r.Name, r.GroupID, r.PlatformID)
	// }
	//
	// batches, err := b.ToSQL()
	// if err != nil {
	// 	return err
	// }
	//
	// for _, b := range batches {
	// 	_, err := db.Exec(
	// 		"INSERT INTO devices ('name', 'group_id', 'platform_id') VALUES "+b.Query,
	// 		b.Args...,
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	//
	return nil
}

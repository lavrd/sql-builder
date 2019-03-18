package sqlbuilder

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"text/template"
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
	ToSQL() ([]*BatchQuery, error)
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
	BatchQueries   []*BatchQuery
	Delimiter      int
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

	b.Delimiter = len(args)

	if (b.curParamsCount+len(args) > b.MaxParams) || (b.curLineCount == b.MaxLine) {
		b.curLineCount = 0
		b.curParamsCount = 0
		b.bqLen++
		b.BatchQueries = append(b.BatchQueries, &BatchQuery{})
	}

	if b.BatchQueries == nil {
		b.BatchQueries = make([]*BatchQuery, 1)
	}

	curbq := b.BatchQueries[b.bqLen-1]
	if curbq == nil {
		curbq = &BatchQuery{}
	}
	curbq.Args = append(curbq.Args, args...)

	b.BatchQueries[b.bqLen-1] = curbq
	b.curLineCount++
	b.curParamsCount += len(args)

	return nil
}

func (b *Builder) ToSQL() ([]*BatchQuery, error) {
	pattern := `{{range .}}({{range .Line}}${{.}},{{end}}),{{end}}`

	tmpl, err := template.New("query").Parse(pattern)
	if err != nil {
		return nil, err
	}

	type Param struct {
		Line []int
	}

	var params []Param

	for _, bq := range b.BatchQueries {
		var line = make([]int, 0)

		for asi := range bq.Args {
			if asi%b.Delimiter == 0 && asi != 0 {
				params = append(params, Param{line})
				line = make([]int, 0)

			}
			line = append(line, asi+1)
		}

		params = append(params, Param{line})

		var buffer = bytes.NewBuffer([]byte{})
		err = tmpl.Execute(buffer, params)
		if err != nil {
			return nil, err
		}

		str := buffer.String()
		runes := []rune(str)
	loop:
		for i, char := range str {
			if i+1 != len(str) {
				cur := fmt.Sprintf("%c", char)
				next := fmt.Sprintf("%c", str[i+1])
				if next == ")" && cur == "," {
					runes[i] = ' '
					str = string(runes)
					goto loop
				}
			}
		}

		runes[len(runes)-1] = ' '
		bq.Query = string(runes)
	}

	return b.BatchQueries, nil
}

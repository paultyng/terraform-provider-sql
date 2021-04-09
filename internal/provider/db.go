package provider

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	// database drivers
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"

	// TODO: sqlite? need to use a pure go driver, i think this one is...
	// _ "modernc.org/sqlite"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type dbQueryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type dbExecer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (p *provider) connect(dsn string) error {
	var err error

	scheme, err := schemeFromURL(dsn)
	if err != nil {
		return err
	}

	switch scheme {
	case "postgres", "postgresql":
		// TODO: use consts for these driver names?
		p.Driver = "pgx"
	case "mysql":
		p.Driver = "mysql"
		dsn = strings.TrimPrefix(dsn, "mysql://")
		// TODO: multistatements? see go-migrate's implementation
		// https://github.com/golang-migrate/migrate/blob/master/database/mysql/mysql.go

		// TODO: also set parseTime=true https://github.com/go-sql-driver/mysql#parsetime
	case "sqlserver":
		p.Driver = "sqlserver"
	default:
		return fmt.Errorf("unexpected datasource name scheme: %q", scheme)
	}

	p.DB, err = sql.Open(string(p.Driver), dsn)
	if err != nil {
		return fmt.Errorf("unable to open database: %w", err)
	}

	// force this to zero, but let callers override config
	p.DB.SetMaxIdleConns(0)

	return nil
}

func schemeFromURL(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("a datasource name is required")
	}

	i := strings.Index(url, ":")

	// No : or : is the first character.
	if i < 1 {
		return "", fmt.Errorf("a scheme for datasource name is required")
	}

	return url[0:i], nil
}

func (p *provider) ValuesForRow(rows *sql.Rows) (map[string]tftypes.Value, map[string]tftypes.Type, error) {
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve column type: %w", err)
	}

	pointers := make([]interface{}, len(colTypes))
	row := map[string]struct {
		index int
		ty    tftypes.Type
		val   interface{}
	}{}

	for i, colType := range colTypes {
		name := colType.Name()
		if name == "?column?" {
			name = fmt.Sprintf("column%d", i)
		}

		ty, rty, err := p.typeAndValueForColType(colType)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to determine type for %q: %w", name, err)
		}

		val := reflect.New(rty)
		pointers[i] = val.Interface()

		row[name] = struct {
			index int
			ty    tftypes.Type
			val   interface{}
		}{i, ty, val.Interface()}
	}

	err = rows.Scan(pointers...)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to scan values: %w", err)
	}

	rowValues := map[string]tftypes.Value{}
	rowTypes := map[string]tftypes.Type{}
	for k, v := range row {
		val := v.val

		// unwrap sql types
		switch tv := val.(type) {
		case *sql.NullInt64:
			if !tv.Valid {
				val = nil
			} else {
				val = &tv.Int64
			}
		case *sql.NullInt32:
			if !tv.Valid {
				val = nil
			} else {
				val = &tv.Int32
			}
		case *sql.NullFloat64:
			if !tv.Valid {
				val = nil
			} else {
				val = &tv.Float64
			}
		case *sql.NullBool:
			if !tv.Valid {
				val = nil
			} else {
				val = &tv.Bool
			}
		case *sql.NullString:
			if !tv.Valid {
				val = nil
			} else {
				val = &tv.String
			}
		case *sql.NullTime:
			if !tv.Valid {
				val = nil
			} else {
				s := tv.Time.UTC().Format(time.RFC3339)
				val = &s
			}
		}

		rowValues[k] = tftypes.NewValue(
			v.ty,
			val,
		)
		rowTypes[k] = v.ty
	}

	return rowValues, rowTypes, nil
}

func (p *provider) typeAndValueForColType(colType *sql.ColumnType) (tftypes.Type, reflect.Type, error) {
	scanType := colType.ScanType()
	kind := scanType.Kind()

	switch p.Driver {
	case "sqlserver":
		switch dbName := colType.DatabaseTypeName(); dbName {
		case "UNIQUEIDENTIFIER":
			return tftypes.String, reflect.TypeOf((*sqlServerUniqueIdentifier)(nil)).Elem(), nil
		case "DECIMAL", "MONEY", "SMALLMONEY":
			// TODO: add diags about converting to numeric?
			return tftypes.String, reflect.TypeOf((*sql.NullString)(nil)).Elem(), nil
		}
	case "mysql":
		switch dbName := colType.DatabaseTypeName(); dbName {
		case "YEAR":
			return tftypes.Number, reflect.TypeOf((*sql.NullInt32)(nil)).Elem(), nil
		case "VARCHAR", "DECIMAL", "TIME", "JSON":
			return tftypes.String, reflect.TypeOf((*sql.NullString)(nil)).Elem(), nil
		case "DATE", "DATETIME":
			return tftypes.String, reflect.TypeOf((*sql.NullTime)(nil)).Elem(), nil
		}
	case "pgx":
		switch dbName := colType.DatabaseTypeName(); dbName {
		// 790 is the oid of money
		case "MONEY", "790":
			// TODO: add diags about converting to numeric?
			return tftypes.String, reflect.TypeOf((*sql.NullString)(nil)).Elem(), nil
		case "TIMESTAMPTZ", "TIMESTAMP", "DATE":
			return tftypes.String, reflect.TypeOf((*sql.NullTime)(nil)).Elem(), nil
		}
	}

	switch scanType {
	case reflect.TypeOf((*sql.NullInt64)(nil)).Elem(),
		reflect.TypeOf((*sql.NullInt32)(nil)).Elem(),
		reflect.TypeOf((*sql.NullFloat64)(nil)).Elem():
		return tftypes.Number, scanType, nil
	case reflect.TypeOf((*sql.NullString)(nil)).Elem():
		return tftypes.String, scanType, nil
	case reflect.TypeOf((*sql.NullBool)(nil)).Elem():
		return tftypes.Bool, scanType, nil
	case reflect.TypeOf((*sql.NullTime)(nil)).Elem():
		return tftypes.String, scanType, nil
	}

	// Force nullable typing for primitives
	switch kind {
	case reflect.String:
		return tftypes.String, reflect.TypeOf((*sql.NullString)(nil)).Elem(), nil
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int,
		reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		return tftypes.Number, reflect.TypeOf((*sql.NullInt64)(nil)).Elem(), nil
	case reflect.Uint64:
		// TODO: uint64 may be a problem in nullint64 if too large?
		return tftypes.Number, reflect.TypeOf((*sql.NullInt64)(nil)).Elem(), nil
	case reflect.Float32, reflect.Float64:
		return tftypes.Number, reflect.TypeOf((*sql.NullFloat64)(nil)).Elem(), nil
	case reflect.Bool:
		return tftypes.Bool, reflect.TypeOf((*sql.NullBool)(nil)).Elem(), nil
	}

	return nil, nil, fmt.Errorf("unexpected type for %q: %q (%s %s)", colType.Name(), colType.DatabaseTypeName(), kind, scanType)
}

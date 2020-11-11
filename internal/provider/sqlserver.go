package provider

import (
	"encoding/hex"
	"errors"
	"fmt"
)

type sqlServerUniqueIdentifier [16]byte

func (u *sqlServerUniqueIdentifier) Scan(v interface{}) error {
	reverse := func(b []byte) {
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}

	switch vt := v.(type) {
	case []byte:
		if len(vt) != 16 {
			return errors.New("mssql: invalid sqlServerUniqueIdentifier length")
		}

		var raw sqlServerUniqueIdentifier

		copy(raw[:], vt)

		reverse(raw[0:4])
		reverse(raw[4:6])
		reverse(raw[6:8])
		*u = raw

		return nil
	case string:
		if len(vt) != 36 {
			return errors.New("mssql: invalid sqlServerUniqueIdentifier string length")
		}

		b := []byte(vt)
		for i, c := range b {
			switch c {
			case '-':
				b = append(b[:i], b[i+1:]...)
			}
		}

		_, err := hex.Decode(u[:], []byte(b))
		return err
	default:
		return fmt.Errorf("mssql: cannot convert %T to sqlServerUniqueIdentifier", v)
	}
}

func (u *sqlServerUniqueIdentifier) ToTerraform5Value() (interface{}, error) {
	if u == nil {
		return nil, nil
	}

	s := fmt.Sprintf("%X-%X-%X-%X-%X", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return &s, nil
}

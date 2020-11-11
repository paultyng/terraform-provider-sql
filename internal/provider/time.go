package provider

import (
	"fmt"
	"reflect"
	"time"
)

type tfTime struct {
	time.Time
}

var tfTimeType = reflect.TypeOf((*tfTime)(nil)).Elem()

func (tf *tfTime) Scan(src interface{}) error {
	t, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("expected time.Time, got %T", src)
	}
	tf.Time = t
	return nil
}

func (tf *tfTime) ToTerraform5Value() (interface{}, error) {
	if tf == nil {
		return nil, nil
	}

	s := tf.Format(time.RFC3339)
	return &s, nil
}

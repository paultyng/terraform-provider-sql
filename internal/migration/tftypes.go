package migration

import "github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"

var (
	ListTFType = tftypes.List{
		ElementType: ValueTFType,
	}
	ValueTFType = tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":   tftypes.String,
			"up":   tftypes.String,
			"down": tftypes.String,
		},
	}
)

func (m Migration) Value() tftypes.Value {
	return tftypes.NewValue(ValueTFType, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, m.ID),
		"up":   tftypes.NewValue(tftypes.String, m.Up),
		"down": tftypes.NewValue(tftypes.String, m.Down),
	})
}

func List(migrations []Migration) tftypes.Value {
	values := []tftypes.Value{}
	for _, m := range migrations {
		values = append(values, m.Value())
	}

	return tftypes.NewValue(ListTFType, values)
}

func FromValue(v tftypes.Value) (Migration, error) {
	m := Migration{}

	var valueMap map[string]tftypes.Value
	err := v.As(&valueMap)
	if err != nil {
		return m, err
	}

	err = valueMap["id"].As(&m.ID)
	if err != nil {
		return m, err
	}

	err = valueMap["up"].As(&m.Up)
	if err != nil {
		return m, err
	}

	err = valueMap["down"].As(&m.Down)
	if err != nil {
		return m, err
	}

	return m, nil
}

func FromListValue(v tftypes.Value) ([]Migration, error) {
	var migrationValues []tftypes.Value
	err := v.As(&migrationValues)
	if err != nil {
		return nil, err
	}

	migrations := []Migration{}
	for _, mValue := range migrationValues {
		m, err := FromValue(mValue)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, nil
}

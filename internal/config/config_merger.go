package config

import "reflect"

func mergeOptionsServer(mergeInto *OptionsServer, newValues OptionsServer) {
	t := reflect.TypeOf(mergeInto).Elem()
	vInto := reflect.ValueOf(mergeInto).Elem()
	vNew := reflect.ValueOf(newValues)
	for i := 0; i < t.NumField(); i++ {
		fieldValueInto := vInto.Field(i)
		intoBeenSet := fieldValueInto.FieldByName("BeenSet")
		intoValue := fieldValueInto.FieldByName("Value")

		fieldValueNew := vNew.Field(i)
		newBeenSet := fieldValueNew.FieldByName("BeenSet")
		newValue := fieldValueNew.FieldByName("Value")

		if newBeenSet.Bool() {
			intoValue.Set(newValue)
			intoBeenSet.Set(newBeenSet)
		}
	}
}

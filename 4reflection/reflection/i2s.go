package main

import (
	"reflect"
	"fmt"
)

//type Simple struct {
//	ID       int
//	Username string
//	Active   bool
//}

func i2s(data interface{}, out interface{}) error {
	inputval := reflect.ValueOf(data)
	outputVal := reflect.ValueOf(out)

	var err error = nil

	//проверка, что пришло
	switch reflect.TypeOf(data).Kind() {
	case reflect.Invalid:
		return fmt.Errorf("No value")

	case reflect.Map:
		{
			if outputVal.Type().Kind() == reflect.Struct {
				return fmt.Errorf("bad map")
			}

			err = ParsingStruct(inputval, outputVal)

		}
	case reflect.Slice:
		{
			if inputval.Type().Kind() != reflect.Indirect(outputVal).Type().Kind() {
				return fmt.Errorf("In and Out types dont match")
			}

			IndValue := reflect.Indirect(inputval)

			sliceValue := reflect.MakeSlice(outputVal.Type().Elem(), IndValue.Len(), IndValue.Cap())

			for i := 0; i < IndValue.Len(); i++ {
				err = ParsingStruct(IndValue.Index(i).Elem(), sliceValue.Index(i))
			}
			outputVal.Elem().Set(sliceValue)
		}
	default:
		return fmt.Errorf("bad type")

	}

	//	fmt.Println(outVal.Type().Field(1))
	return err
}

func ParsingStruct(inputval, out reflect.Value) error {
	var err error = nil
	outVal := reflect.Indirect(out)

	for i := 0; i < outVal.NumField(); i++ {

		typeField := outVal.Type().Field(i)                                 //поле структуры
		InElem := inputval.MapIndex(reflect.ValueOf(typeField.Name)).Elem() //значение в мапе для ключа с таким именем поля структуры

		switch typeField.Type.Kind() {
		case reflect.Int:

			switch InElem.Type().Kind() {
			case reflect.Float64:

				outVal.Field(i).SetInt(int64(InElem.Float()))
			case reflect.Int:
				outVal.Field(i).SetInt(InElem.Int())
			case reflect.String:
				return fmt.Errorf("bad int")
			}

		case reflect.String:
			if InElem.Type().Kind() != reflect.String {
				return fmt.Errorf(" Bad string!")
			}

			outVal.Field(i).SetString(InElem.String())

		case reflect.Bool:
			if InElem.Type().Kind() != reflect.Bool {
				return fmt.Errorf(" Bad bool")
			}
			outVal.Field(i).SetBool(InElem.Bool())

		case reflect.Struct:
			if InElem.Type().Kind() == reflect.Bool {
				return fmt.Errorf("Bad struct")
			}

			strValue := reflect.New(typeField.Type)
			ParsingStruct(InElem, strValue)
			outVal.Field(i).Set(strValue.Elem())

		case reflect.Slice:
			if InElem.Type().Kind() != reflect.Slice {
				return fmt.Errorf("Struct вместо Slice!")
			}

			sliceValue := reflect.MakeSlice(typeField.Type, InElem.Len(), InElem.Cap())
			for i := 0; i < InElem.Len(); i++ {
				ParsingStruct(InElem.Index(i).Elem(), sliceValue.Index(i))
			}

			outVal.Field(i).Set(sliceValue)

		default:
			return fmt.Errorf("Can not parse this type")
		}

	}
	return err

}

// Got:
//                &main.Simple{ID:0, Username:"", Active:false}
//                Expected:
//                &main.Simple{ID:42, Username:"rvasily", Active:true}


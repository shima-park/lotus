package inject

import (
	"reflect"
)

type Receptor struct {
	StructName      string
	StructFieldName string
	InjectName      string
	ReflectType     string
}

func GetFuncReqAndRespReceptorList(f interface{}) ([]Receptor, []Receptor) {
	if reflect.TypeOf(f).Kind() != reflect.Func {
		panic("Interface must be a callable func")
	}

	t := reflect.TypeOf(f)

	var reqReceptors []Receptor
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)

		for argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		if argType.Kind() != reflect.Struct {
			continue
		}

		val := reflect.New(argType)

		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			continue
		}

		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			structField := typ.Field(i)
			injectName := structField.Tag.Get("inject")

			var tt reflect.Type
			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				tt = InterfaceOf(nilPtr.Interface())
			} else {
				tt = f.Type()
			}

			reqReceptors = append(reqReceptors, Receptor{
				StructName:      typ.Name(),
				StructFieldName: structField.Name,
				InjectName:      injectName,
				ReflectType:     tt.String(),
			})
		}
	}

	var respReceptors []Receptor
	for i := 0; i < t.NumOut(); i++ {
		outType := t.Out(i)

		if outType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}

		for outType.Kind() == reflect.Ptr {
			outType = outType.Elem()
		}

		if outType.Kind() != reflect.Struct {
			continue
		}

		val := reflect.New(outType)
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			continue
		}

		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			structField := typ.Field(i)
			injectName := structField.Tag.Get("inject")

			var tt reflect.Type
			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				tt = InterfaceOf(nilPtr.Interface())
			} else {
				tt = f.Type()
			}

			respReceptors = append(respReceptors, Receptor{
				StructName:      typ.Name(),
				StructFieldName: structField.Name,
				InjectName:      injectName,
				ReflectType:     tt.String(),
			})
		}
	}
	return reqReceptors, respReceptors
}

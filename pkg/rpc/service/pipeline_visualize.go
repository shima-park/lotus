package service

//import (
//	"io"
//	"reflect"
//
//	"github.com/olekukonko/tablewriter"
//	"github.com/pkg/errors"
//	"github.com/shima-park/lotus/pkg/common/inject"
//	"github.com/shima-park/lotus/pkg/executor"
//	"github.com/shima-park/lotus/pkg/processor"
//)
//
//func init() {
//	err := executor.AddVisualizer("term", TermVisualizer)
//	if err != nil {
//		panic(err)
//	}
//}
//
//func TermVisualizer(w io.Writer, executor executor.Executor) error {
//	printExecutorComponents(w, executor)
//	printExecutorProcessors(w, executor)
//	return nil
//}
//
//func printExecutorComponents(w io.Writer, p executor.Executor) {
//	table := tablewriter.NewWriter(w)
//	table.SetHeader([]string{
//		"Component Name", "Reflect Type", "Inject Name", "Raw Config", "Description",
//	})
//	table.SetRowLine(true)
//	for _, c := range p.ListComponents() {
//		arr := []string{
//			c.Name,
//			c.Component.Instance().Type().String(),
//			c.Component.Instance().Name(),
//			c.RawConfig,
//			c.Factory.Description(),
//		}
//
//		table.Rich(arr, []tablewriter.Colors{
//			tablewriter.Colors{},
//			tablewriter.Colors{},
//			tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor},
//			tablewriter.Colors{},
//			tablewriter.Colors{},
//		})
//
//	}
//	table.Render()
//}
//
//func printExecutorProcessors(w io.Writer, p executor.Executor) {
//	mdeErrs := filterMissingDependencyError(p.CheckDependence())
//
//	table := tablewriter.NewWriter(w)
//	table.SetHeader([]string{"Processor Name", //"Config",
//		"Struct name", "Field", "Reflect type", "Inject name"})
//	table.SetAutoMergeCells(true)
//	table.SetRowLine(true)
//	for _, p := range p.ListProcessors() {
//		requests, responses := getFuncReqAndRespReceptorList(p.Processor)
//
//		for _, req := range requests {
//			mdeErr := matchError(mdeErrs, req)
//			if mdeErr != nil {
//				table.Rich(
//					[]string{p.Name, req.StructName, req.StructFieldName, req.ReflectType, req.InjectName},
//					[]tablewriter.Colors{
//						tablewriter.Colors{},
//						tablewriter.Colors{},
//						tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
//						tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
//						tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
//					})
//			} else {
//				table.Rich(
//					[]string{p.Name, req.StructName, req.StructFieldName, req.ReflectType, req.InjectName},
//					[]tablewriter.Colors{
//						tablewriter.Colors{},
//						tablewriter.Colors{},
//						tablewriter.Colors{},
//						tablewriter.Colors{},
//						tablewriter.Colors{tablewriter.Normal, tablewriter.FgCyanColor},
//					})
//			}
//		}
//
//		for _, resp := range responses {
//			table.Rich(
//				[]string{p.Name, resp.StructName, resp.StructFieldName, resp.ReflectType, resp.InjectName},
//				[]tablewriter.Colors{
//					tablewriter.Colors{},
//					tablewriter.Colors{},
//					tablewriter.Colors{},
//					tablewriter.Colors{},
//					tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor},
//				})
//		}
//	}
//	table.Render()
//}
//
//func filterMissingDependencyError(errs []error) []executor.MissingDependencyError {
//	var mdeErrs []executor.MissingDependencyError
//	for _, err := range errs {
//		cause, ok := errors.Cause(err).(executor.MissingDependencyError)
//		if ok {
//			mdeErrs = append(mdeErrs, cause)
//		}
//	}
//	return mdeErrs
//}
//
//func matchError(mdeErrs []executor.MissingDependencyError, r Receptor) *executor.MissingDependencyError {
//	for _, mdeErr := range mdeErrs {
//		if mdeErr.Field == r.StructFieldName &&
//			mdeErr.ReflectType == r.ReflectType &&
//			mdeErr.InjectName == r.InjectName {
//			return &mdeErr
//		}
//	}
//	return nil
//}
//
//type Receptor struct {
//	StructName      string
//	StructFieldName string
//	InjectName      string
//	ReflectType     string
//}
//
//func getFuncReqAndRespReceptorList(f interface{}) ([]Receptor, []Receptor) {
//	if err := processor.Validate(f); err != nil {
//		return nil, nil
//	}
//
//	t := reflect.TypeOf(f)
//
//	var reqReceptors []Receptor
//	for i := 0; i < t.NumIn(); i++ {
//		argType := t.In(i)
//
//		for argType.Kind() == reflect.Ptr {
//			argType = argType.Elem()
//		}
//
//		if argType.Kind() != reflect.Struct {
//			continue
//		}
//
//		val := reflect.New(argType)
//
//		for val.Kind() == reflect.Ptr {
//			val = val.Elem()
//		}
//
//		if val.Kind() != reflect.Struct {
//			continue
//		}
//
//		typ := val.Type()
//
//		for i := 0; i < val.NumField(); i++ {
//			f := val.Field(i)
//			structField := typ.Field(i)
//			injectName := structField.Tag.Get("inject")
//
//			var tt reflect.Type
//			if f.Type().Kind() == reflect.Interface {
//				nilPtr := reflect.New(f.Type())
//				tt = inject.InterfaceOf(nilPtr.Interface())
//			} else {
//				tt = f.Type()
//			}
//
//			reqReceptors = append(reqReceptors, Receptor{
//				StructName:      typ.Name(),
//				StructFieldName: structField.Name,
//				InjectName:      injectName,
//				ReflectType:     tt.String(),
//			})
//		}
//	}
//
//	var respReceptors []Receptor
//	for i := 0; i < t.NumOut(); i++ {
//		outType := t.Out(i)
//
//		if outType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
//			continue
//		}
//
//		for outType.Kind() == reflect.Ptr {
//			outType = outType.Elem()
//		}
//
//		if outType.Kind() != reflect.Struct {
//			continue
//		}
//
//		val := reflect.New(outType)
//		for val.Kind() == reflect.Ptr {
//			val = val.Elem()
//		}
//
//		if val.Kind() != reflect.Struct {
//			continue
//		}
//
//		typ := val.Type()
//
//		for i := 0; i < val.NumField(); i++ {
//			f := val.Field(i)
//			structField := typ.Field(i)
//			injectName := structField.Tag.Get("inject")
//
//			var tt reflect.Type
//			if f.Type().Kind() == reflect.Interface {
//				nilPtr := reflect.New(f.Type())
//				tt = inject.InterfaceOf(nilPtr.Interface())
//			} else {
//				tt = f.Type()
//			}
//
//			respReceptors = append(respReceptors, Receptor{
//				StructName:      typ.Name(),
//				StructFieldName: structField.Name,
//				InjectName:      injectName,
//				ReflectType:     tt.String(),
//			})
//		}
//	}
//	return reqReceptors, respReceptors
//}

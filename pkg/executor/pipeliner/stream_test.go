package pipeliner

import (
	"reflect"
	"strings"
	"testing"
)

func TestStream(t *testing.T) {
	f := &Stream{processor: Processor{Name: "root"}}

	equalsSlice(t, travel(f, 0), []string{"root"})

	err := f.AppendByParentName("root", &Stream{processor: Processor{Name: "step1"}})
	handleErr(t, err)

	equalsSlice(t, travel(f, 0), []string{"root", "step1"})

	err = f.InsertBefore("step1", &Stream{processor: Processor{Name: "step0"}})
	handleErr(t, err)

	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1"})

	err = f.InsertAfter("step1", &Stream{processor: Processor{Name: "step2"}})
	handleErr(t, err)
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step2"})

	err = f.InsertAfter("step2", &Stream{processor: Processor{Name: "step3"}})
	handleErr(t, err)
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step2", "step3"})

	err = f.AppendByParentName("step1", &Stream{processor: Processor{Name: "step1.5"}})
	handleErr(t, err)
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step1.5", "step2", "step3"})

	step1, ok := f.Get("step1")
	equal(t, true, ok)
	equal(t, step1.childs[0].Name(), "step1.5")

	err = f.Delete("step0")
	handleErr(t, err)
	equalsSlice(t, travel(f, 0), []string{"root", "step1", "step1.5", "step2", "step3"})

	err = f.Delete("root")
	handleErr(t, err)
	equalsSlice(t, travel(f, 0), []string{"root"})
}

func equalsSlice(t *testing.T, actual, expected []string) {
	equal(t, strings.Join(actual, ","), strings.Join(expected, ","))
}

func equal(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func handleErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

// +build js

package js_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gopherjs/gopherjs/js"
)

var dummys = js.Global.Call("eval", `({
	someBool: true,
	someString: "abc\u1234",
	someInt: 42,
	someFloat: 42.123,
	someArray: [41, 42, 43],
	add: function(a, b) {
		return a + b;
	},
	mapArray: function(array, f) {
		var newArray = new Array(array.length);
		for (var i = 0; i < array.length; i++) {
			newArray[i] = f(array[i]);
		}
		return newArray;
	},
	toUnixTimestamp: function(d) {
		return d.getTime() / 1000;
	},
	testField: function(o) {
		return o.Field;
	},
	testMethod: function(o) {
		return o.Method(42);
	},
	isEqual: function(a, b) {
		return a === b;
	},
	call: function(f, a) {
		f(a);
	},
})`)

func TestBool(t *testing.T) {
	e := true
	o := dummys.Get("someBool")
	if v := o.Bool(); v != e {
		t.Errorf("expected %#v, got %#v", e, v)
	}
	if i := o.Interface().(bool); i != e {
		t.Errorf("expected %#v, got %#v", e, i)
	}
	if dummys.Set("otherBool", e); dummys.Get("otherBool").Bool() != e {
		t.Fail()
	}
}

func TestStr(t *testing.T) {
	e := "abc\u1234"
	o := dummys.Get("someString")
	if v := o.String(); v != e {
		t.Errorf("expected %#v, got %#v", e, v)
	}
	if i := o.Interface().(string); i != e {
		t.Errorf("expected %#v, got %#v", e, i)
	}
	if dummys.Set("otherString", e); dummys.Get("otherString").String() != e {
		t.Fail()
	}
}

func TestInt(t *testing.T) {
	e := 42
	o := dummys.Get("someInt")
	if v := o.Int(); v != e {
		t.Errorf("expected %#v, got %#v", e, v)
	}
	if i := int(o.Interface().(float64)); i != e {
		t.Errorf("expected %#v, got %#v", e, i)
	}
	if dummys.Set("otherInt", e); dummys.Get("otherInt").Int() != e {
		t.Fail()
	}
}

func TestFloat(t *testing.T) {
	e := 42.123
	o := dummys.Get("someFloat")
	if v := o.Float(); v != e {
		t.Errorf("expected %#v, got %#v", e, v)
	}
	if i := o.Interface().(float64); i != e {
		t.Errorf("expected %#v, got %#v", e, i)
	}
	if dummys.Set("otherFloat", e); dummys.Get("otherFloat").Float() != e {
		t.Fail()
	}
}

func TestUndefined(t *testing.T) {
	if dummys == js.Undefined || dummys.Get("xyz") != js.Undefined {
		t.Fail()
	}
}

func TestNull(t *testing.T) {
	var null js.Object
	dummys.Set("test", nil)
	if null != nil || dummys == nil || dummys.Get("test") != nil {
		t.Fail()
	}
}

func TestLength(t *testing.T) {
	if dummys.Get("someArray").Length() != 3 {
		t.Fail()
	}
}

func TestIndex(t *testing.T) {
	if dummys.Get("someArray").Index(1).Int() != 42 {
		t.Fail()
	}
}

func TestSetIndex(t *testing.T) {
	dummys.Get("someArray").SetIndex(2, 99)
	if dummys.Get("someArray").Index(2).Int() != 99 {
		t.Fail()
	}
}

func TestCall(t *testing.T) {
	var i int64 = 40
	if dummys.Call("add", i, 2).Int() != 42 {
		t.Fail()
	}
	if dummys.Call("add", js.Global.Call("eval", "40"), 2).Int() != 42 {
		t.Fail()
	}
}

func TestInvoke(t *testing.T) {
	var i int64 = 40
	if dummys.Get("add").Invoke(i, 2).Int() != 42 {
		t.Fail()
	}
}

func TestNew(t *testing.T) {
	if js.Global.Get("Array").New(42).Length() != 42 {
		t.Fail()
	}
}

type StructWithJsField1 struct {
	js.Object
	Length int                  `js:"length"`
	Slice  func(int, int) []int `js:"slice"`
}

type StructWithJsField2 struct {
	object js.Object            // to hide members from public API
	Length int                  `js:"length"`
	Slice  func(int, int) []int `js:"slice"`
}

type Wrapper1 struct {
	StructWithJsField1
	WrapperLength int `js:"length"`
}

type Wrapper2 struct {
	innerStruct   *StructWithJsField2
	WrapperLength int `js:"length"`
}

func TestReadingJsField(t *testing.T) {
	a := StructWithJsField1{Object: js.Global.Get("Array").New(42)}
	b := &StructWithJsField2{object: js.Global.Get("Array").New(42)}
	wa := Wrapper1{StructWithJsField1: a}
	wb := Wrapper2{innerStruct: b}
	if a.Length != 42 || b.Length != 42 || wa.Length != 42 || wa.WrapperLength != 42 || wb.WrapperLength != 42 {
		t.Fail()
	}
}

func TestWritingJsField(t *testing.T) {
	a := StructWithJsField1{Object: js.Global.Get("Object").New()}
	b := &StructWithJsField2{object: js.Global.Get("Object").New()}
	a.Length = 42
	b.Length = 42
	if a.Get("length").Int() != 42 || b.object.Get("length").Int() != 42 {
		t.Fail()
	}
}

func TestCallingJsField(t *testing.T) {
	a := &StructWithJsField1{Object: js.Global.Get("Array").New(100)}
	b := &StructWithJsField2{object: js.Global.Get("Array").New(100)}
	a.SetIndex(3, 123)
	b.object.SetIndex(3, 123)
	f := a.Slice
	a2 := a.Slice(2, 44)
	b2 := b.Slice(2, 44)
	c2 := f(2, 44)
	if len(a2) != 42 || len(b2) != 42 || len(c2) != 42 || a2[1] != 123 || b2[1] != 123 || c2[1] != 123 {
		t.Fail()
	}
}

func TestUnboxing(t *testing.T) {
	a := StructWithJsField1{Object: js.Global.Get("Object").New()}
	b := &StructWithJsField2{object: js.Global.Get("Object").New()}
	if !dummys.Call("isEqual", a, a.Object).Bool() || !dummys.Call("isEqual", b, b.object).Bool() {
		t.Fail()
	}
	wa := Wrapper1{StructWithJsField1: a}
	wb := Wrapper2{innerStruct: b}
	if !dummys.Call("isEqual", wa, a.Object).Bool() || !dummys.Call("isEqual", wb, b.object).Bool() {
		t.Fail()
	}
}

func TestBoxing(t *testing.T) {
	o := js.Global.Get("Object").New()
	dummys.Call("call", func(a StructWithJsField1) {
		if a.Object != o {
			t.Fail()
		}
	}, o)
	dummys.Call("call", func(a *StructWithJsField2) {
		if a.object != o {
			t.Fail()
		}
	}, o)
	dummys.Call("call", func(a Wrapper1) {
		if a.Object != o {
			t.Fail()
		}
	}, o)
	dummys.Call("call", func(a Wrapper2) {
		if a.innerStruct.object != o {
			t.Fail()
		}
	}, o)
}

func TestFunc(t *testing.T) {
	a := dummys.Call("mapArray", []int{1, 2, 3}, func(e int64) int64 { return e + 40 })
	b := dummys.Call("mapArray", []int{1, 2, 3}, func(e ...int64) int64 { return e[0] + 40 })
	if a.Index(1).Int() != 42 || b.Index(1).Int() != 42 {
		t.Fail()
	}

	add := dummys.Get("add").Interface().(func(...interface{}) js.Object)
	var i int64 = 40
	if add(i, 2).Int() != 42 {
		t.Fail()
	}
}

func TestDate(t *testing.T) {
	d := time.Date(2013, time.August, 27, 22, 25, 11, 0, time.UTC)
	if dummys.Call("toUnixTimestamp", d).Int() != int(d.Unix()) {
		t.Fail()
	}

	d2 := js.Global.Get("Date").New(d.UnixNano() / 1000000).Interface().(time.Time)
	if !d2.Equal(d) {
		t.Fail()
	}
}

func TestEquality(t *testing.T) {
	if js.Global.Get("Array") != js.Global.Get("Array") || js.Global.Get("Array") == js.Global.Get("String") {
		t.Fail()
	}
	type S struct{ js.Object }
	o1 := js.Global.Get("Object").New()
	o2 := js.Global.Get("Object").New()
	a := S{o1}
	b := S{o1}
	c := S{o2}
	if a != b || a == c {
		t.Fail()
	}
}

func TestThis(t *testing.T) {
	dummys.Set("testThis", func(_ string) { // string argument to force wrapping
		if js.This != dummys {
			t.Fail()
		}
	})
	dummys.Call("testThis", "")
}

func TestArguments(t *testing.T) {
	dummys.Set("testArguments", func() {
		if len(js.Arguments) != 3 || js.Arguments[1].Int() != 1 {
			t.Fail()
		}
	})
	dummys.Call("testArguments", 0, 1, 2)
}

func TestSameFuncWrapper(t *testing.T) {
	a := func(_ string) {} // string argument to force wrapping
	b := func(_ string) {} // string argument to force wrapping
	if !dummys.Call("isEqual", a, a).Bool() || dummys.Call("isEqual", a, b).Bool() {
		t.Fail()
	}
	if !dummys.Call("isEqual", somePackageFunction, somePackageFunction).Bool() {
		t.Fail()
	}
	if !dummys.Call("isEqual", (*T).someMethod, (*T).someMethod).Bool() {
		t.Fail()
	}
	t1 := &T{}
	t2 := &T{}
	if !dummys.Call("isEqual", t1.someMethod, t1.someMethod).Bool() || dummys.Call("isEqual", t1.someMethod, t2.someMethod).Bool() {
		t.Fail()
	}
}

func somePackageFunction(_ string) {
}

type T struct{}

func (t *T) someMethod() {
	println(42)
}

func TestError(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fail()
		}
		if _, ok := err.(error); !ok {
			t.Fail()
		}
		jsErr, ok := err.(*js.Error)
		if !ok || jsErr.Get("stack") == js.Undefined {
			t.Fail()
		}
	}()
	js.Global.Get("notExisting").Call("throwsError")
}

type F struct {
	Field int
}

func TestExternalizeField(t *testing.T) {
	if dummys.Call("testField", map[string]int{"Field": 42}).Int() != 42 {
		t.Fail()
	}
	if dummys.Call("testField", F{42}).Int() != 42 {
		t.Fail()
	}
}

type M struct {
	x int
}

func (m *M) Method(x int) {
	m.x = x
}

func TestMakeWrapper(t *testing.T) {
	m := &M{}
	dummys.Call("testMethod", js.MakeWrapper(m))
	if m.x != 42 {
		t.Fail()
	}
}

func TestReflection(t *testing.T) {
	o := js.Global.Call("eval", "({ answer: 42 })")
	if reflect.ValueOf(o).Interface().(js.Object) != o {
		t.Fail()
	}

	type S struct {
		Field js.Object
	}
	s := S{o}

	v := reflect.ValueOf(&s).Elem()
	if v.Field(0).Interface().(js.Object).Get("answer").Int() != 42 {
		t.Fail()
	}
	if v.Field(0).MethodByName("Get").Call([]reflect.Value{reflect.ValueOf("answer")})[0].Interface().(js.Object).Int() != 42 {
		t.Fail()
	}
	v.Field(0).Set(reflect.ValueOf(js.Global.Call("eval", "({ answer: 100 })")))
	if s.Field.Get("answer").Int() != 100 {
		t.Fail()
	}

	if fmt.Sprintf("%+v", s) != "{Field:[object Object]}" {
		t.Fail()
	}
}

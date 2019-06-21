package tre_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/g7r/tre"
)

type w struct{}

type testCase struct {
	t    interface{}
	re   interface{}
	T, U, T1, T2, T3 interface{}
}

func typeOf(proto interface{}) reflect.Type {
	if proto == nil {
		return nil
	}

	ct := reflect.TypeOf(proto)
	if ct.Kind() == reflect.Map && ct.Key() == reflect.TypeOf(w{}) {
		ct = ct.Elem()
	}
	return ct
}

func (c *testCase) assertMatch(t *testing.T) {
	p, ok := tre.MatchType(typeOf(c.t), c.re)
	if !ok {
		t.Errorf("match failed: %s to %T", typeOf(c.t), c.re)
	}

	matchPlaceholder := func(ct interface{}, pt reflect.Type, name string) {
		if typeOf(ct) != pt {
			t.Logf("at %s to %T", typeOf(c.t), c.re)
			t.Errorf("invalid capture %s: expected %s but got %s", name, typeOf(ct), pt)
		}
	}

	matchPlaceholder(c.T, p.T, "T")
	matchPlaceholder(c.U, p.U, "U")
	matchPlaceholder(c.T1, p.T1, "T1")
	matchPlaceholder(c.T2, p.T2, "T2")
	matchPlaceholder(c.T3, p.T3, "T3")
}

func (c *testCase) assertNoMatch(t *testing.T) {
	if _, ok := tre.MatchType(reflect.TypeOf(c.t), c.re); ok {
		t.Errorf("match shouldn't succeed: %T to %T", c.t, c.re)
	}
}

func TestMatchType(t *testing.T) {
	for _, c := range []testCase{
		{t: map[w]int{}, re: 0},
		{t: map[w]int{}, re: map[tre.G]int{}},
		{t: map[w]int{}, re: map[tre.Or]func(int, string){}},
		{t: map[w]string{}, re: map[tre.Or]func(int, string){}},
		{t: map[w]func(int){}, re: func(map[tre.Or]func(int, string)) {}},
		{t: map[w]func(){}, re: func(map[tre.ZeroOrOne]func(context.Context)) {}},
		{t: map[w]func(context.Context){}, re: func(map[tre.ZeroOrOne]context.Context) {}},
		{t: map[w]func(int, bool, string, uint){}, re: func(map[tre.ZeroOrMore]tre.Any) {}},
		{t: map[w]func(context.Context, *reflect.Value){}, re: map[tre.G]func(context.Context, *tre.Struct) map[tre.ZeroOrOne]error{}},
		{t: map[w]func(context.Context, *reflect.Value) error{}, re: map[tre.G]func(context.Context, *tre.Struct) map[tre.ZeroOrOne]error{}},
	} {
		c.assertMatch(t)
	}

	for _, c := range []testCase{
		{t: map[w]func(int, string){}, re: func(tre.T, tre.T) {}},
		{t: map[w]func(){}, re: func(tre.T, map[tre.ZeroOrOne]tre.T) {}},
		{t: map[w]func(){}, re: func(map[tre.ZeroOrOne]tre.T, tre.T) {}},
		{t: map[w]func(int, string){}, re: func(tre.T, map[tre.ZeroOrOne]tre.T) {}},
		{t: map[w]func(string){}, re: func(*tre.T) {}},
		{t: map[w]func(int, string, bool){}, re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2)) {}},
		{t: map[w]func(*reflect.Value) error{}, re: map[tre.G]func(context.Context, *tre.Struct) map[tre.ZeroOrOne]error{}},
		{t: map[w]func(context.Context, reflect.Value) error{}, re: map[tre.G]func(context.Context, *tre.Struct) map[tre.ZeroOrOne]error{}},
		{t: map[w]func(int, string, map[int]string){}, re: func(tre.T, tre.U, map[tre.U]tre.T) {}},
		{t: map[w]func(int) string{}, re: (func(tre.T) tre.T)(nil)},
	} {
		c.assertNoMatch(t)
	}
}

func TestAssignableTypesDontMatch(t *testing.T) {
	type emptyStruct struct{}

	for _, c := range []testCase{
		{t: map[w]int{}, re: map[tre.G]interface{}{}},
		{t: map[w]interface{}{}, re: 0},
		{t: map[w]emptyStruct{}, re: struct{}{}},
		{t: map[w]struct{}{}, re: emptyStruct{}},
	} {
		c.assertNoMatch(t)
	}
}

func TestAssignableTo(t *testing.T) {
	type emptyStruct struct{}

	for _, c := range []testCase{
		{t: map[w]int{}, re: map[tre.AssignableTo]interface{}{}},
		{t: map[w]emptyStruct{}, re: map[tre.AssignableTo]struct{}{}},
		{t: map[w]struct{}{}, re: map[tre.AssignableTo]emptyStruct{}},
	} {
		c.assertMatch(t)
	}
}

func TestAssignableFrom(t *testing.T) {
	type emptyStruct struct{}

	for _, c := range []testCase{
		{t: map[w]interface{}{}, re: map[tre.AssignableFrom]int{}},
		{t: map[w]emptyStruct{}, re: map[tre.AssignableTo]struct{}{}},
		{t: map[w]struct{}{}, re: map[tre.AssignableTo]emptyStruct{}},
	} {
		c.assertMatch(t)
	}
}

func TestCapture(t *testing.T) {
	type recFunc func() recFunc

	for _, c := range []testCase{
		{
			t:  map[w]func(int){},
			re: func(tre.T) {},
			T:  map[w]int{},
		},
		{
			t:  map[w]func(int, int){},
			re: func(tre.T, tre.T) {},
			T:  map[w]int{},
		},
		{
			t:  map[w]func(int, string, map[string]int){},
			re: func(tre.T, tre.U, map[tre.U]tre.T) {},
			T:  map[w]int{},
			U:  "",
		},
		{
			t:  map[w]func(int) int{},
			re: (func(tre.T) tre.T)(nil),
			T:  map[w]int{},
		},
		{
			t:  map[w]func(recFunc){},
			re: map[tre.G]func(func() tre.T){},
			T:  map[w]recFunc{},
		},
		{
			t:  map[w]func(int, int){},
			re: func(tre.T, tre.T) {},
			T:  map[w]int{},
		},
		{
			t: map[w]func(int){},
			re: func(map[tre.ZeroOrOne]tre.T, tre.T) {},
			T: map[w]int{},
		},
		{
			t: map[w]func(string, string){},
			re: func(map[tre.ZeroOrOne]tre.T, tre.T) {},
			T: map[w]string{},
		},
		{
			t: map[w]func(context.Context, int){},
			re: func(map[tre.ZeroOrOne]context.Context, tre.T) {},
			T: map[w]int{},
		},
		{
			t: map[w]func(int){},
			re: func(map[tre.ZeroOrOne]context.Context, tre.T) {},
			T: map[w]int{},
		},
		{
			t: map[w]func(*int){},
			re: func(*tre.T) {},
			T: map[w]int{},
		},
		{
			t: map[w]func(*string){},
			re: func(*tre.T) {},
			T: map[w]string{},
		},
		{
			t: map[w]func(int, bool, string){},
			re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2, tre.T3)) {},
			T1: map[w]int{},
			T2: map[w]bool{},
			T3: map[w]string{},
		},
		{
			t: map[w]func(int, bool, string, bool){},
			re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2, tre.T3)) {},
			T1: map[w]int{},
			T2: map[w]bool{},
			T3: map[w]string{},
		},
		{
			t: map[w]func(int){},
			re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2, tre.T3)) {},
			T1: map[w]int{},
		},
		{
			t: map[w]func(){},
			re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2, tre.T3)) {},
		},
		{
			t: map[w]func(int, int, string, bool){},
			re: func(map[tre.ZeroOrMore]tre.T, map[tre.ZeroOrMore]tre.Any) {},
			T: map[w]int{},
		},
	} {
		c.assertMatch(t)
	}
}

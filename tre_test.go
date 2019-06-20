package tre_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/g7r/tre"
)

func TestMatchType(t *testing.T) {
	type testCase struct {
		value reflect.Type
		re    interface{}
	}

	for _, c := range []testCase{
		{value: reflect.TypeOf(0), re: 0},
		{value: reflect.TypeOf(0), re: map[tre.G]int{}},
		{value: reflect.TypeOf(0), re: map[tre.Or]func(int, string){}},
		{value: reflect.TypeOf(""), re: map[tre.Or]func(int, string){}},
		{value: reflect.TypeOf(func(int) {}), re: func(map[tre.Or]func(int, string)) {}},
		{value: reflect.TypeOf(func() {}), re: func(map[tre.ZeroOrOne]func(context.Context)) {}},
		{value: reflect.TypeOf(func(context.Context) {}), re: func(map[tre.ZeroOrOne]context.Context) {}},
		{value: reflect.TypeOf(func(int, int) {}), re: func(tre.T, tre.T) {}},
		{value: reflect.TypeOf(func(int) {}), re: func(map[tre.ZeroOrOne]tre.T, tre.T) {}},
		{value: reflect.TypeOf(func(string, string) {}), re: func(map[tre.ZeroOrOne]tre.T, tre.T) {}},
		{value: reflect.TypeOf(func(context.Context, int) {}), re: func(map[tre.ZeroOrOne]context.Context, tre.T) {}},
		{value: reflect.TypeOf(func(int) {}), re: func(map[tre.ZeroOrOne]context.Context, tre.T) {}},
		{value: reflect.TypeOf(func(*int) {}), re: func(*tre.T) {}},
		{value: reflect.TypeOf(func(*string) {}), re: func(*tre.T) {}},
		{value: reflect.TypeOf(func(int, bool, string) {}), re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2, tre.T3)) {}},
		{value: reflect.TypeOf(func(int, bool, string, bool) {}), re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2, tre.T3)) {}},
		{value: reflect.TypeOf(func(int) {}), re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2, tre.T3)) {}},
		{value: reflect.TypeOf(func() {}), re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2, tre.T3)) {}},
		{value: reflect.TypeOf(func(int, bool, string, uint) {}), re: func(map[tre.ZeroOrMore]tre.Any) {}},
		{value: reflect.TypeOf(func(context.Context, *reflect.Value) {}), re: map[tre.G]func(context.Context, *tre.Struct) map[tre.ZeroOrOne]error{}},
		{value: reflect.TypeOf((func(context.Context, *reflect.Value) error)(nil)), re: map[tre.G]func(context.Context, *tre.Struct) map[tre.ZeroOrOne]error{}},
		{value: reflect.TypeOf((func(int, int, string, bool))(nil)), re: func(map[tre.ZeroOrMore]tre.T, map[tre.ZeroOrMore]tre.Any) {}},
	} {
		if _, ok := tre.MatchType(c.value, c.re); !ok {
			t.Errorf("match failed: %s to %T", c.value, c.re)
		}
	}

	for _, c := range []testCase{
		{value: reflect.TypeOf(func(int, string) {}), re: func(tre.T, tre.T) {}},
		{value: reflect.TypeOf(func() {}), re: func(tre.T, map[tre.ZeroOrOne]tre.T) {}},
		{value: reflect.TypeOf(func() {}), re: func(map[tre.ZeroOrOne]tre.T, tre.T) {}},
		{value: reflect.TypeOf(func(int, string) {}), re: func(tre.T, map[tre.ZeroOrOne]tre.T) {}},
		{value: reflect.TypeOf(func(string) {}), re: func(*tre.T) {}},
		{value: reflect.TypeOf(func(int, string, bool) {}), re: func(map[tre.ZeroOrMore]map[tre.Or]func(tre.T1, tre.T2)) {}},
		{value: reflect.TypeOf((func(*reflect.Value) error)(nil)), re: map[tre.G]func(context.Context, *tre.Struct) map[tre.ZeroOrOne]error{}},
		{value: reflect.TypeOf((func(context.Context, reflect.Value) error)(nil)), re: map[tre.G]func(context.Context, *tre.Struct) map[tre.ZeroOrOne]error{}},
		{value: reflect.TypeOf((func(int, string, map[int]string))(nil)), re: func(tre.T, tre.U, map[tre.U]tre.T) {}},
		{value: reflect.TypeOf((func(int) string)(nil)), re: (func(tre.T) tre.T)(nil)},
	} {
		if _, ok := tre.MatchType(c.value, c.re); ok {
			t.Errorf("match shouldn't succeed: %s to %T", c.value, c.re)
		}
	}
}

func TestCapture(t *testing.T) {
	type recFunc func() recFunc

	type testCase struct {
		value interface{}
		re    interface{}
		t     reflect.Type
		u     reflect.Type
	}

	for _, c := range []testCase{
		{
			value: (func(int))(nil),
			re:    func(tre.T) {},
			t:     reflect.TypeOf(0),
		},
		{
			value: (func(int, int))(nil),
			re:    func(tre.T, tre.T) {},
			t:     reflect.TypeOf(0),
		},
		{
			value: (func(int, string, map[string]int))(nil),
			re:    func(tre.T, tre.U, map[tre.U]tre.T) {},
			t:     reflect.TypeOf(0),
			u:     reflect.TypeOf(""),
		},
		{
			value: (func(int) int)(nil),
			re:    (func(tre.T) tre.T)(nil),
			t:     reflect.TypeOf(0),
		},
		{
			value: (func(recFunc))(nil),
			re:    (func(func() tre.T))(nil),
			t:     reflect.TypeOf(recFunc(nil)),
		},
	} {
		p, ok := tre.MatchType(reflect.TypeOf(c.value), c.re)
		if !ok {
			t.Errorf("match failed: %T to %T", c.value, c.re)
		}

		if c.t != p.T {
			t.Errorf("invalid capture T: expected %s but got %s", c.t, p.T)
		}

		if c.u != p.U {
			t.Errorf("invalid capture U: expected %s but got %s", c.u, p.U)
		}
	}
}


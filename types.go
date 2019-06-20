package tre

import (
	"reflect"
)

type (
	// надо ли?
	G         struct{}
	Or        struct{}
	Any       struct{}
	Struct    struct{}
	Interface struct{}

	T  struct{}
	U  struct{}
	V  struct{}
	T1 struct{}
	T2 struct{}
	T3 struct{}
	T4 struct{}
	T5 struct{}
	T6 struct{}
	T7 struct{}
	T8 struct{}

	ZeroOrOne  struct{}
	ZeroOrMore struct{}
	OneOrMore  struct{}
)

var (
	gType         = reflect.TypeOf(G{})
	orType        = reflect.TypeOf(Or{})
	anyType       = reflect.TypeOf(Any{})
	structType    = reflect.TypeOf(Struct{})
	interfaceType = reflect.TypeOf(Interface{})

	tType  = reflect.TypeOf(T{})
	uType  = reflect.TypeOf(U{})
	vType  = reflect.TypeOf(V{})
	t1Type = reflect.TypeOf(T1{})
	t2Type = reflect.TypeOf(T2{})
	t3Type = reflect.TypeOf(T3{})
	t4Type = reflect.TypeOf(T4{})
	t5Type = reflect.TypeOf(T5{})
	t6Type = reflect.TypeOf(T6{})
	t7Type = reflect.TypeOf(T7{})
	t8Type = reflect.TypeOf(T8{})

	zeroOrOneType  = reflect.TypeOf(ZeroOrOne{})
	zeroOrMoreType = reflect.TypeOf(ZeroOrMore{})
	oneOrMoreType  = reflect.TypeOf(OneOrMore{})
)

func isMatcherType(t reflect.Type) (reflect.Type, bool) {
	if t.Kind() == reflect.Map {
		switch t.Key() {
		case gType, orType, zeroOrOneType, zeroOrMoreType, oneOrMoreType:
			return t.Key(), true
		}
	}

	switch t {
	case anyType, tType, uType, vType,
		t1Type, t2Type, t3Type, t4Type, t5Type, t6Type, t7Type, t8Type,
		structType, interfaceType:
		return t, true
	}

	return nil, false
}

type placeholderBinding struct {
	t   reflect.Type
	val reflect.Type
}

type placeholderMap []placeholderBinding

func (m *placeholderMap) put(t reflect.Type, val reflect.Type) bool {
	for _, b := range *m {
		if b.t == t {
			return b.val == val
		}
	}

	*m = append(*m, placeholderBinding{t: t, val: val})
	return true
}

func getIn(t reflect.Type) []reflect.Type {
	in := make([]reflect.Type, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		in[i] = t.In(i)
	}
	return in
}

func getOut(t reflect.Type) []reflect.Type {
	in := make([]reflect.Type, t.NumOut())
	for i := 0; i < t.NumOut(); i++ {
		in[i] = t.Out(i)
	}
	return in
}

type matchContext struct {
	capturedTypes placeholderMap
}

func (ctx matchContext) fork() *matchContext {
	ctx.capturedTypes = append(placeholderMap(nil), ctx.capturedTypes...)
	return &ctx
}

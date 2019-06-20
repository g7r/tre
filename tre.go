package tre

import (
	"fmt"
	"reflect"
)

var matchState = &state{matchFn: func(*matchContext, reflect.Type) (*matchContext, bool) {
	return nil, false
}}

type state struct {
	matchFn   func(*matchContext, reflect.Type) (*matchContext, bool)
	out, out1 *state
}

type stateItem struct {
	s   *state
	ctx *matchContext
}

type stateList struct {
	s []stateItem
}

func (l *stateList) addState(ctx *matchContext, s *state) {
	if s == nil {
		return
	}

	for i, ss := range l.s {
		if ss.s == s {
			l.s[i].ctx = ctx
			return
		}
	}

	if s.matchFn == nil {
		l.addState(ctx.fork(), s.out)
		l.addState(ctx.fork(), s.out1)
	} else {
		l.s = append(l.s, stateItem{s: s, ctx: ctx})
	}
}

func startList(ctx *matchContext, s *state) stateList {
	var l stateList
	l.addState(ctx, s)
	return l
}

func step(clist stateList, t reflect.Type, nlist *stateList) {
	nlist.s = nil
	for _, s := range clist.s {
		mctx, ok := s.s.matchFn(s.ctx, t)
		if ok {
			nlist.addState(mctx, s.s.out)
		}
	}
}

func isMatch(l stateList) (*matchContext, bool) {
	for _, s := range l.s {
		if s.s == matchState {
			return s.ctx, true
		}
	}

	return nil, false
}

func matchNFA(ctx *matchContext, typeList []reflect.Type, nfa *state) (*matchContext, bool) {
	var clist, nlist stateList
	clist = startList(ctx, nfa)
	for _, t := range typeList {
		step(clist, t, &nlist)
		clist, nlist = nlist, clist
	}

	return isMatch(clist)
}

func compileMatchState(key, re reflect.Type, nextState *state) *state {
	switch key {
	case gType:
		return compileState(re.Elem(), nextState)

	case orType:
		argsType := re.Elem()
		if argsType.NumIn() == 0 {
			panic(fmt.Sprintf("'Or' without alternatives: %s", re))
		}

		var lastState *state
		for i := 0; i < argsType.NumIn(); i++ {
			curState := compileState(argsType.In(i), nextState)
			if lastState == nil {
				lastState = curState
			} else {
				lastState = &state{
					out:  lastState,
					out1: curState,
				}
			}
		}

		return lastState

	case zeroOrOneType:
		elemState := compileState(re.Elem(), nextState)
		return &state{
			out:  elemState,
			out1: nextState,
		}

	case zeroOrMoreType:
		altState := &state{}
		altState.out = compileState(re.Elem(), altState)
		altState.out1 = nextState
		return altState

	case oneOrMoreType:
		altState := &state{}
		elemState := compileState(re.Elem(), altState)
		altState.out = elemState
		altState.out1 = nextState
		return altState

	case anyType:
		return &state{
			out: nextState,
			matchFn: func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
				return ctx, true
			},
		}

	case structType:
		return &state{
			out: nextState,
			matchFn: func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
				if t.Kind() != reflect.Struct {
					return nil, false
				}

				return ctx, true
			},
		}

	case interfaceType:
		return &state{
			out: nextState,
			matchFn: func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
				if t.Kind() != reflect.Interface {
					return nil, false
				}

				return ctx, true
			},
		}

	case tType, uType, vType, t1Type, t2Type, t3Type, t4Type, t5Type, t6Type, t7Type, t8Type:
		return &state{
			out: nextState,
			matchFn: func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
				if !ctx.capturedTypes.put(key, t) {
					return nil, false
				}

				return ctx, true
			},
		}
	}

	panic(fmt.Sprintf("unexpected re: %s", re))
}

func compileState(reType reflect.Type, nextState *state) *state {
	if key, isMatcher := isMatcherType(reType); isMatcher {
		return compileMatchState(key, reType, nextState)
	}

	var s state
	var matchFn func(*matchContext, reflect.Type) (*matchContext, bool)
	switch reType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Bool, reflect.String, reflect.Complex64, reflect.Complex128, reflect.Float32, reflect.Float64,
		reflect.Struct, reflect.Interface:

		matchFn = func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
			if t != reType {
				return nil, false
			}

			return ctx, true
		}

	case reflect.Slice, reflect.Array, reflect.Ptr:
		elemNFA := compileNFA([]reflect.Type{reType.Elem()})
		matchFn = func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
			return matchNFA(ctx, []reflect.Type{t.Elem()}, elemNFA)
		}

	case reflect.Map:
		keyNFA := compileNFA([]reflect.Type{reType.Key()})
		elemNFA := compileNFA([]reflect.Type{reType.Elem()})
		matchFn = func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
			mctx1, ok1 := matchNFA(ctx, []reflect.Type{t.Key()}, keyNFA)
			if !ok1 {
				return nil, false
			}

			mctx2, ok2 := matchNFA(mctx1, []reflect.Type{t.Elem()}, elemNFA)
			if !ok2 {
				return nil, false
			}

			return mctx2, true
		}

	case reflect.Func:
		inNFA := compileNFA(getIn(reType))
		outNFA := compileNFA(getOut(reType))
		matchFn = func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
			mctx1, ok1 := matchNFA(ctx, getIn(t), inNFA)
			if !ok1 {
				return nil, false
			}

			mctx2, ok2 := matchNFA(mctx1, getOut(t), outNFA)
			if !ok2 {
				return nil, false
			}

			return mctx2, true
		}

	case reflect.Chan:
		elemNFA := compileNFA([]reflect.Type{reType.Elem()})
		matchFn = func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
			if t.ChanDir() != reType.ChanDir() {
				return nil, false
			}

			mctx, ok := matchNFA(ctx, []reflect.Type{t.Elem()}, elemNFA)
			return mctx, ok
		}
	}

	s.out = nextState
	s.matchFn = func(ctx *matchContext, t reflect.Type) (*matchContext, bool) {
		if t.Kind() != reType.Kind() {
			return nil, false
		}

		return matchFn(ctx, t)
	}

	return &s
}

func compileNFA(reList []reflect.Type) *state {
	if len(reList) == 0 {
		return matchState
	}

	return compileState(reList[0], compileNFA(reList[1:]))
}

type Placeholders struct {
	T, U, V, T1, T2, T3, T4, T5, T6, T7, T8 reflect.Type
}

func MatchType(t reflect.Type, re interface{}) (Placeholders, bool) {
	s := compileNFA([]reflect.Type{reflect.TypeOf(re)})
	ctx, ok := matchNFA(&matchContext{}, []reflect.Type{t}, s)
	if !ok {
		return Placeholders{}, false
	}

	if len(ctx.capturedTypes) == 0 {
		return Placeholders{}, true
	}

	p := Placeholders{}
	for _, ct := range ctx.capturedTypes {
		switch ct.t {
		case tType:
			p.T = ct.val
		case uType:
			p.U = ct.val
		case vType:
			p.V = ct.val
		case t1Type:
			p.T1 = ct.val
		case t2Type:
			p.T2 = ct.val
		case t3Type:
			p.T3 = ct.val
		case t4Type:
			p.T4 = ct.val
		case t5Type:
			p.T5 = ct.val
		case t6Type:
			p.T6 = ct.val
		case t7Type:
			p.T7 = ct.val
		case t8Type:
			p.T8 = ct.val
		}
	}

	return p, true
}

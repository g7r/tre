package tre

import (
	"fmt"
	"reflect"
)

var matchState = &state{matchFn: func(Placeholders, reflect.Type) (Placeholders, bool) {
	return noMatch()
}}

type state struct {
	matchFn   func(Placeholders, reflect.Type) (Placeholders, bool)
	out, out1 *state
}

type stateItem struct {
	*state
	p Placeholders
}

type stateList []stateItem

func (l *stateList) addState(s *state, p Placeholders) {
	if s == nil {
		return
	}

	if s.matchFn == nil {
		l.addState(s.out, p)
		l.addState(s.out1, p)
		return
	}

	si := stateItem{state: s, p: p}
	for _, ss := range *l {
		if ss == si {
			return
		}
	}

	*l = append(*l, si)
}

func startList(s *state, p Placeholders) stateList {
	var l stateList
	l.addState(s, p)
	return l
}

func isMatch(l stateList) (Placeholders, bool) {
	for _, s := range l {
		if s.state == matchState {
			return s.p, true
		}
	}

	return noMatch()
}

func matchNFA(p Placeholders, typeList []reflect.Type, nfa *state) (Placeholders, bool) {
	var nlist stateList
	clist := startList(nfa, p)
	for _, t := range typeList {
		for _, s := range clist {
			if mp, ok := s.matchFn(s.p, t); ok {
				//noinspection GoNilness
				nlist.addState(s.out, mp)
			}
		}

		clist, nlist = nlist, clist
		nlist = nlist[:0]
	}

	return isMatch(clist)
}

func noMatch() (Placeholders, bool) {
	return Placeholders{}, false
}

func compilePlaceholderState(key reflect.Type, pfFunc func(*Placeholders) *reflect.Type, nextState *state) *state {
	return &state{
		out: nextState,
		matchFn: func(p Placeholders, t reflect.Type) (Placeholders, bool) {
			if !p.put(key, t, pfFunc(&p)) {
				return noMatch()
			}

			return p, true
		},
	}
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
			matchFn: func(p Placeholders, t reflect.Type) (Placeholders, bool) {
				return p, true
			},
		}

	case structType:
		return &state{
			out: nextState,
			matchFn: func(p Placeholders, t reflect.Type) (Placeholders, bool) {
				if t.Kind() != reflect.Struct {
					return noMatch()
				}

				return p, true
			},
		}

	case interfaceType:
		return &state{
			out: nextState,
			matchFn: func(p Placeholders, t reflect.Type) (Placeholders, bool) {
				if t.Kind() != reflect.Interface {
					return noMatch()
				}

				return p, true
			},
		}

	case assignableToType:
		return &state{
			out: nextState,
			matchFn: func(p Placeholders, t reflect.Type) (Placeholders, bool) {
				if !t.AssignableTo(re.Elem()) {
					return noMatch()
				}

				return p, true
			},
		}

	case assignableFromType:
		return &state{
			out: nextState,
			matchFn: func(p Placeholders, t reflect.Type) (Placeholders, bool) {
				if !re.Elem().AssignableTo(t) {
					return noMatch()
				}

				return p, true
			},
		}

	case tType:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T }, nextState)
	case uType:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.U }, nextState)
	case vType:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.V }, nextState)
	case t1Type:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T1 }, nextState)
	case t2Type:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T2 }, nextState)
	case t3Type:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T3 }, nextState)
	case t4Type:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T4 }, nextState)
	case t5Type:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T5 }, nextState)
	case t6Type:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T6 }, nextState)
	case t7Type:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T7 }, nextState)
	case t8Type:
		return compilePlaceholderState(key, func(p *Placeholders) *reflect.Type { return &p.T8 }, nextState)
	}

	panic(fmt.Sprintf("unexpected re: %s", re))
}

func compileState(reType reflect.Type, nextState *state) *state {
	if key, isMatcher := isMatcherType(reType); isMatcher {
		return compileMatchState(key, reType, nextState)
	}

	var s state
	var matchFn func(Placeholders, reflect.Type) (Placeholders, bool)
	switch reType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Bool, reflect.String, reflect.Complex64, reflect.Complex128, reflect.Float32, reflect.Float64,
		reflect.Struct, reflect.Interface:

		matchFn = func(p Placeholders, t reflect.Type) (Placeholders, bool) {
			if t != reType {
				return noMatch()
			}

			return p, true
		}

	case reflect.Slice, reflect.Array, reflect.Ptr:
		elemNFA := compileNFA([]reflect.Type{reType.Elem()})
		matchFn = func(p Placeholders, t reflect.Type) (Placeholders, bool) {
			return matchNFA(p, []reflect.Type{t.Elem()}, elemNFA)
		}

	case reflect.Map:
		keyNFA := compileNFA([]reflect.Type{reType.Key()})
		elemNFA := compileNFA([]reflect.Type{reType.Elem()})
		matchFn = func(p Placeholders, t reflect.Type) (Placeholders, bool) {
			mp1, ok1 := matchNFA(p, []reflect.Type{t.Key()}, keyNFA)
			if !ok1 {
				return noMatch()
			}

			mp2, ok2 := matchNFA(mp1, []reflect.Type{t.Elem()}, elemNFA)
			if !ok2 {
				return noMatch()
			}

			return mp2, true
		}

	case reflect.Func:
		inNFA := compileNFA(getIn(reType))
		outNFA := compileNFA(getOut(reType))
		matchFn = func(p Placeholders, t reflect.Type) (Placeholders, bool) {
			mp1, ok1 := matchNFA(p, getIn(t), inNFA)
			if !ok1 {
				return noMatch()
			}

			mp2, ok2 := matchNFA(mp1, getOut(t), outNFA)
			if !ok2 {
				return noMatch()
			}

			return mp2, true
		}

	case reflect.Chan:
		elemNFA := compileNFA([]reflect.Type{reType.Elem()})
		matchFn = func(p Placeholders, t reflect.Type) (Placeholders, bool) {
			if t.ChanDir() != reType.ChanDir() {
				return noMatch()
			}

			return matchNFA(p, []reflect.Type{t.Elem()}, elemNFA)
		}
	}

	s.out = nextState
	s.matchFn = func(p Placeholders, t reflect.Type) (Placeholders, bool) {
		if t.Kind() != reType.Kind() {
			return noMatch()
		}

		return matchFn(p, t)
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

func (p *Placeholders) put(t reflect.Type, val reflect.Type, pt *reflect.Type) bool {
	if *pt != nil {
		return *pt == val
	}

	*pt = val
	return true
}

func MatchType(t reflect.Type, re interface{}) (Placeholders, bool) {
	s := compileNFA([]reflect.Type{reflect.TypeOf(re)})
	return matchNFA(Placeholders{}, []reflect.Type{t}, s)
}

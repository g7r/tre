# tre
Regular expressions on Go types

Every seasoned Go developer wrote some amount of this kind of ugly code:
```go
// handlerFn must be function.
// handlerFn may accept context.Context as a first parameter.
// handlerFn must accept payload as a last parameter.
// handlerFn must return result.
// handlerFn may return error.
// payload must be a pointer to struct.
func registerHandler(handlerFn interface{}) {
    t := reflect.TypeOf(handlerFn)
    if t.Kind() != reflect.Func {
        panic(fmt.Sprintf("handlerFn should be a func, got %s", t))
    }

    contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
    switch t.NumIn() {
    case 1:
        if t.In(0) == contextType {
            panic(fmt.Sprintf("handlerFn is missing payload parameter, got %s", t))
        }
    case 2:
        if t.In(0) != contextType {
            panic(fmt.Sprintf("handlerFn first parameter should be Context, got %s", t))
        }
    default:
        panic(fmt.Sprintf("handlerFn should accept 1 or 2 parameters, got %s", t))
    }

    payloadType := t.In(t.NumIn()-1)
    
    if payloadType.Kind() != reflect.Ptr || payloadType.Elem().Kind != reflect.Struct {
        panic(fmt.Sprintf("payload type should be a pointer to struct, got %s", t))
    }

    errorType := reflect.TypeOf((*error)(nil)).Elem()
    switch t.NumOut() {
    case 1:
        if t.Out(0) == errorType {
            panic(fmt.Sprintf("handlerFn is missing result having only error, got %s", t))
        }
    case 2:
        if t.Out(1) != errorType {
            panic(fmt.Sprintf("handlerFn should have only one result, got %s", t))
        }
    default:
        panic(fmt.Sprintf("handlerFn should return 1 or 2 result parameters, got %s", t))
    }

    resultType := t.Out(0)

    // do the actual work
}
```

Wow! Such a lot of tedious and not much easy to read code. We should have
a tool that will simplify such kind of type checks. Generics should be 
that tool. But not for Go. At least not yet. But we don't have to submissively
wait! Go language already provides us some means to add a little bit more
generics to it.

## The Problem

We need a library that will do all the dirty work for us. Actually, type
matching has a lot of similarities with regular expression. Let's take
regular expressions as a model for building type patterns and matching
them to actual types.

We'll need:
- exact type matches
- matches by kind
- wildcard matches
- quantifiers, like a zero-or-one, zero-or-many and one-or-many
- alternations
- sequences

But the most important problem is what kind of API our library should 
provide? Let's recall how to get an instance of `reflect.Type` in Go:
```go
valType := reflect.TypeOf(val)
ifaceType := reflect.TypeOf((*context.Context)(nil)).Elem()
```

What kind of API we can provide using `reflect.Type`:
```go
ok := MatchType(t, Func(
    ZeroOrOne(reflect.TypeOf((*context.Context)(nil)).Elem()),
    *Struct,
). Returns(
    AnyType,
    ZeroOrOne(reflect.TypeOf((*error)(nil)).Elem())
))
```

Tedious, burdensome and not that clear that we'd like to. We definitely
don't want our library API consumer to write code like this. If Go had generics
our API could look like below:
```go
// fictional Go
ok := MatchType(t, (func(ZeroOrOne<context.Context>, *Struct) (AnyType, ZeroOrOne<error>))(nil))
```

A much cleaner version. We use natural Go construct `func` to match a function!
Isn't that pretty! Fortunately for us, Go language already have built-in generic types that we
can (ab)use to build an API like that.

First thing we should do is to declare primitive matcher types like `AnyType`
and `Struct` that don't require parameters. Such generic Go types like pointer,
slice, array, chan, map and func can be naturally used to match pointers,
slices, arrays, channels, maps and functions.   

Next, we're going to abuse Go maps to express our matchers with parameters. 
Map generic type has two type parameters, key type and value type. We can 
use key type to express generic matcher type and value type as its parameter.
For example:
```go
map[ZeroOrOne]context.Context // matches context.Context or nothing
map[AssignableTo]struct{}     // matches any type that could be assigned to an empty struct
```

The only thing that left are multiple parameters matchers. Let's abuse
func for that purpose:
```go
map[Or]func(int, string, bool) // matches either int, string or bool
```

So this would be our DSL.

## The Solution

Let's rewrite the code from first example using `tre`:
```go
// handlerFn must match func(context.Context?, *Payload) (Result, error?).
// Payload must be a struct.
func registerHandler(handlerFn interface{}) {
    t := reflect.TypeOf(handlerFn)
    var pattern func(map[ZeroOrOne]context.Context, *T) (U, map[ZeroOrOne]error)
    ps, ok := MatchType(t, pattern)
    if !ok {
        panic(fmt.Sprintf("handlerFn should match pattern %T, got %s", pattern, t))
    }
    
    payloadType := reflect.PtrTo(ps.T)
    resultType := ps.U

    if ps.T.Elem().Kind() != reflect.Struct {
        panic(fmt.Sprintf("payload should be a pointer to struct, got %s", payloadType))
    }

    // do the actual work
}
```

Much simpler and clearer version.
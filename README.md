# gofsm (wip)

A lightweight, type-safe, concurrent-safe Extended Finite State Machine (EFSM) library for Go.
Easily build flexible and scalable FSMs, with rich context-aware transitions and a fluent builder API.

[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/10490/badge)](https://www.bestpractices.dev/projects/10490)

## 📦 Installation

```bash
go get github.com/akishichinibu/gofsm
```

## 🚀 Quick Start

```go
package main

import (
 "fmt"

 "github.com/akishichinibu/gofsm"
)

type Ctx struct {
 Bias int
}

func main() {
 sm, err := gofsm.NewEFSM[Ctx, int, int](func(b gofsm.EFSMBuilder[Ctx, int, int]) {
  b.From(1).On(2).ToConst(3)
  b.From(2).On(3).ToConst(5)
  b.From(5).On(10).To(func(ctx Ctx, from int, by int) (int, error) {
   return from + by + ctx.Bias, nil
  })
 })
 if err != nil {
  panic(err)
 }

 ctx := Ctx{Bias: -1}

// normal transit
 next, _ := sm.Transit(ctx, 1, 2)
 fmt.Println(next) // Output: 3

 next, _ = sm.Transit(ctx, 2, 3)
 fmt.Println(next) // Output: 5

 next, _ = sm.Transit(ctx, 5, 10)
 fmt.Println(next) // Output: 14 （5+10-1）

 // illegal transit
 _, err = sm.Transit(ctx, 3, 1)
 if err != nil {
  fmt.Println("error:", err) // Output: error: illegal transit from 3 by 1
 }
}
```

## 🔧 State and Operation Comparison Semantics

In `gofsm`, both `State` and `Operation` types must be `comparable`. And there are some important nuances, especially if you're using custom types.

### ✅ State: Recommended as Enum-like Value

Your FSM state must be a comparable type — typically a simple type like `string`, `int`, or a user-defined type which implemented the `fmt.Stringer`, and `gofsm` will use the return value of `String()` as the state value.

```go
type OrderStatus string

const (
    Pending  OrderStatus = "pending"
    Approved OrderStatus = "approved"
)

// or

type MyState struct{
    ID int
    Name string
}

var _ fmt.Stringer = &MyState{}

func (s *MyState) String() string {
    return fmt.Sprintf("%d:%s", s.ID, s.Name)
}
```

### ✅ Operation: Value-Based Matching

Operations (also known as events or inputs) are matched by exact value. This supports both simple constants and struct-based variants.

```go
type Add struct{ Diff int }

fsm.From("running").On(Add{Diff: 5}).To("done")

fsm.Transit("running", Add{Diff: 5}) // ✅ matches
fsm.Transit("running", Add{Diff: 3}) // ❌ does not match
```

#### Pointer vs Value

If you pass a pointer to a struct as an operation, `gofsm` will dereference it and use the `reflect.DeepEqual` to compare the value. This means that if you pass a pointer to a struct, it will be compared by value, not by reference.

```go
op := &Add{Diff: 10}
fsm.From("ready").On(op).To("done")

fsm.Transit("ready", &Add{Diff: 10})        // ✅ OK
fsm.Transit("ready", Add{Diff: 10})         // ✅ OK
```

#### `interface` as Operation

In most cases, you can define a interface that contains multiple operations, and use it as the operation type. This allows you to carry different parameters in the operation, and use type assertion to get the specific operation type.

```go
type Operation interface {
    MyOperation()
}

type Add struct{ Diff int }
func (Add) MyOperation() {}

type Sub struct{}
func (Sub) MyOperation() {}
```

## 📜 License

MIT License.

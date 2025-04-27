# gofsm (wip)

A lightweight, type-safe, concurrent-safe Extended Finite State Machine (EFSM) library for Go.
Easily build flexible and scalable FSMs, with rich context-aware transitions and a fluent API.

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

## 📜 License

MIT License.

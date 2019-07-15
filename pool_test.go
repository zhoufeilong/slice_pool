package common

import (
    "math/rand"
    "testing"

    "github.com/stretchr/testify/assert"
)

type CharOperator interface {
    Id() int64
}

type LMap1 struct {
    Slice    []CharOperator
    Idx      map[int64]int
    FreeList []int // 空余位置。
}

func newLMap1() *LMap1 {
    return &LMap1{
        Slice:    make([]CharOperator, 0),
        Idx:      make(map[int64]int),
        FreeList: make([]int, 0),
    }
}

func (lmap *LMap1) Len() int {
    return len(lmap.Idx)
}

func (lmap *LMap1) Get(id int64) CharOperator {
    if idx, present := lmap.Idx[id]; present {
        return lmap.Slice[idx]
    }
    return nil
}

func (lmap *LMap1) Del(id int64) {
    if idx, present := lmap.Idx[id]; present {
        delete(lmap.Idx, id)
        lmap.Slice[idx] = nil
        lmap.FreeList = append(lmap.FreeList, idx)
    } else {
        log.Panicf("id: %v idx not in idx map!", id)
    }
}


func (lmap *LMap1) Add(char CharOperator) {
    id := char.Id()
    if len(lmap.FreeList) > 0 {
        idx := lmap.FreeList[0]
        lmap.Slice[idx] = char
        lmap.Idx[id] = idx
        lmap.FreeList = lmap.FreeList[1:]
    } else {
        lmap.Slice = append(lmap.Slice, char)
        lmap.Idx[id] = len(lmap.Slice) - 1
    }
}

func (lmap *LMap1) Iter(f func(CharOperator)) {
    for _, c := range lmap.Slice {
        if c != nil {
            f(c)
        }
    }
}

type TestStruct struct {
    PoolInfo
    F1 int32
    F2 int64
    *Pvp
    *Pvp2
}

func (ts TestStruct) Id() int64 {
    return ts.F2
}

func (ts TestStruct) Release() {
    pool1.Free(ts.Pvp.PoolId)
}

type Pvp struct {
    PoolInfo
    Hh int32
}
type Pvp2 struct {
    PoolInfo
    Hh int32
}

var map1 *LMap
var pool1 *LPool
var poolTestNum int

func t1() {
    pool1 = newLPool(func(l int) interface{} { return make([]Pvp, l, l) }, func(slice interface{}, idx int) PoolOperator { return &(slice.([]Pvp)[idx]) })
    freeF := func(obj interface{}) { ptr := obj.(*TestStruct); ptr.Release() }
    map1 = newLMap(func(l int) interface{} { return make([]TestStruct, l, l) }, func(slice interface{}, idx int) PoolOperator { return &(slice.([]TestStruct)[idx]) }, freeF)
    for i := 0; i < poolTestNum; i++ {
        s := map1.New(int64(i)).(*TestStruct)
        *s = TestStruct{
            F1:  int32(i),
            F2:  int64(i),
            Pvp: pool1.New().(*Pvp),
        }
    }
}

func t2() {
    cnt := 0
    for j := 0; j < poolTestNum; j++ {
        index := rand.Int63n(int64(poolTestNum))
        obj := map1.Get(index)
        if obj != nil {
            cnt++
            obj1 := obj.(*TestStruct)
            if obj1.F1 > 10000 {
                continue
            }
        }
    }
    for j := 0; j < poolTestNum; j++ {
        map1.Del(int64(j))
    }
}

func t3() {
    cnt := 0
    fn := func(obj PoolOperator) {
        ts := obj.(*TestStruct)
        if ts.F2 < 0 {
            println(ts.F2)
        }
        cnt++
    }
    for i := 0; i < 10000; i++ {
        map1.Iter(fn)
   }
    println(cnt)
}

func t4(lamp *LMap1) {
    cnt := 0
    fn := func(char CharOperator) {
        cnt++
    }
    for i := 0; i < 10000; i++ {
        lamp.Iter(fn)
    }
    println(cnt)
}

func testF() {
    t1()
    t3()
    t2()
    lmap := newLMap1()
    for i := 0; i < poolTestNum; i++ {
        s := &TestStruct{
            F1: int32(i),
            F2: int64(i),
        }
        lmap.Add(s)
    }
    t4(lmap)
}

func ObjectTest() {
    poolTestNum = 10000
    testF()
}

func TestPool(t *testing.T) {
    t.Parallel()
    ObjectTest()
    assert.Equal(t, poolTestNum, pool1.FreeList.Len(), "should equal")
}

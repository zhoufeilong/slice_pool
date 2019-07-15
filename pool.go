package common

import "reflect"

/*
* 此pool不能并发使用,并发必须自己上锁
 */

type PoolInfo struct {
    PoolId int 
}

func (poolObj PoolInfo) SetPoolId(poolId int) {
    poolObj.PoolId = poolId
}

type PoolOperator interface {
    SetPoolId(int)
}

type FreeSlice struct {
    List  []int // 空余位置。
    Start int 
    End   int 
}

func (slice *FreeSlice) Len() int {
    return slice.End - slice.Start
}

func (slice *FreeSlice) Push(idx int) {
    l := len(slice.List)
    if slice.End < l { 
        slice.List[slice.End] = idx 
        slice.End++
    } else if l == slice.End-slice.Start {
        slice.List = append(slice.List, idx)
        slice.End++
    } else {
        slice.Start--
        slice.List[slice.Start] = idx 
    }   
}

func (slice *FreeSlice) Pop() int {
    idx := slice.List[slice.Start]
    slice.Start++
    return idx 
}

func newFreeSlice() *FreeSlice {
    return &FreeSlice{
        List: make([]int, 0),
    }
}

type LMap struct {
    Pool  *LPool
    Idx   map[int64]int
    FreeF func(obj interface{})
    CopyF func(ptr interface{}, structObj interface{})
}

type LPool struct {
    makeFn   func(l int) interface{}
    getI     func(interface{}, int) PoolOperator
    slice    interface{}
    Len      int
    Cap      int
    DataFree []bool
    FreeList *FreeSlice // 空余位置。
}

func newLPool(makeFn func(int) interface{}, getI func(interface{}, int) PoolOperator) *LPool {
    return &LPool{
        makeFn:   makeFn,
        getI:     getI,
        FreeList: newFreeSlice(),
    }
}

func (pool *LPool) Get(idx int) interface{} {
    return pool.getI(pool.slice, idx)
}

func (pool *LPool) Iter(f func(PoolOperator)) {
    for i := 0; i < pool.Len; i++ {
        obj := pool.getI(pool.slice, i)
        if pool.DataFree[i] {
            continue
        }
        f(obj)
    }
}

func (pool *LPool) New() interface{} {
    obj, _ := pool.newObj()
    return obj
}

func (pool *LPool) newObj() (interface{}, int) {
    var idx int
    if pool.FreeList.Len() > 0 {
        idx = pool.FreeList.Pop()
        pool.DataFree[idx] = false
    } else {
        l, needNew := nextCap(pool.Cap, pool.Len+1)
        if needNew {
            newSlice := pool.makeFn(l)
            if pool.slice != nil {
                reflect.Copy(reflect.ValueOf(newSlice), reflect.ValueOf(pool.slice))
            }
            pool.slice = newSlice
            pool.Cap = l
            free := make([]bool, l)
            copy(free, pool.DataFree)
            pool.DataFree = free
        }
        idx = pool.Len
        pool.Len++
    }
    obj := pool.getI(pool.slice, idx)
    obj.SetPoolId(idx)
    return obj, idx
}

func (pool *LPool) Free(poolId int) interface{} {
    pool.FreeList.Push(poolId)
    obj := pool.getI(pool.slice, poolId)
    pool.DataFree[poolId] = true
    pool.Len--
    return obj
}

func newLMap(makeFn func(int) interface{}, getI func(interface{}, int) PoolOperator, freeF func(interface{})) *LMap {
    return &LMap{
        Pool:  newLPool(makeFn, getI),
        Idx:   make(map[int64]int),
        FreeF: freeF,
    }
}
func (lmap *LMap) Get(id int64) interface{} {
    if idx, present := lmap.Idx[id]; present {
        return lmap.Pool.Get(idx)
    }
    return nil
}

func (lmap *LMap) Del(id int64) {
    if idx, present := lmap.Idx[id]; present {
        delete(lmap.Idx, id)
        obj := lmap.Pool.Free(idx)
        if lmap.FreeF != nil {
            lmap.FreeF(obj)
        }
    } else {
        log.Panicf("id: %v idx not in idx map!", id)
    }
}

//推荐先new出来，后面赋值，这样不会分配对象
func (lmap *LMap) New(id int64) interface{} {
    obj, idx := lmap.Pool.newObj()
    lmap.Idx[id] = idx
    return obj
}

func (lmap *LMap) Iter(f func(PoolOperator)) {
    lmap.Pool.Iter(f)
}

func nextCap(oldLen int, newLen int) (int, bool) {
    m := oldLen
    if newLen <= m {
        return m, false
    }
    if m == 0 {
        m = 2
    }
    for m < newLen {
        if oldLen < 1024 {
            m += m
        } else {
            m += m / 4
        }
    }
    return m, true
}

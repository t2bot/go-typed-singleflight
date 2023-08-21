package typedsf

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type TestVal string

const retVal TestVal = "testing"

func TestDo(t *testing.T) {
	callCount := 0
	workFn := func() (TestVal, error) {
		callCount++
		return retVal, nil
	}

	key := "key"
	g := new(Group[TestVal])
	r, err, shared := g.Do(key, workFn)
	if err != nil {
		t.Fatal(err)
	}
	if shared {
		t.Error("Expected a non-shared result")
	}
	if r != retVal {
		t.Errorf("Expected %s but got %s", retVal, r)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDoError(t *testing.T) {
	expectedErr := errors.New("this is expected")
	callCount := 0
	workFn := func() (TestVal, error) {
		callCount++
		return retVal, expectedErr
	}

	key := "key"
	g := new(Group[TestVal])
	r, err, shared := g.Do(key, workFn)
	if err != nil && !errors.Is(err, expectedErr) {
		t.Fatal(err)
	}
	if !errors.Is(err, expectedErr) {
		t.Error("Expected error, got none")
	}
	if shared {
		t.Error("Expected a non-shared result")
	}
	if r != retVal {
		t.Errorf("Expected %s but got %s", retVal, r)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDoShared(t *testing.T) {
	workWg1 := new(sync.WaitGroup)
	workWg2 := new(sync.WaitGroup)
	workCh := make(chan int, 1)
	callCount := 0
	workFn := func() (TestVal, error) {
		callCount++
		if callCount == 1 {
			workWg1.Done()
		}
		v := <-workCh
		workCh <- v
		time.Sleep(10 * time.Millisecond)
		return retVal, nil
	}

	key := "key"
	g := new(Group[TestVal])
	readFn := func() {
		defer workWg2.Done()
		workWg1.Done()
		r, err, _ := g.Do(key, workFn)
		if err != nil {
			t.Error(err)
		}
		if r != retVal {
			t.Errorf("Expected %s but got %s", retVal, r)
		}
	}

	const max = 10
	workWg1.Add(1)
	for i := 0; i < max; i++ {
		workWg1.Add(1)
		workWg2.Add(1)
		go readFn()
	}
	workWg1.Wait()
	workCh <- 1
	workWg2.Wait()
	if callCount <= 0 || callCount >= max {
		t.Errorf("Expected between 1 and %d calls, got %d", max-1, callCount)
	}
}

func TestDoChan(t *testing.T) {
	callCount := 0
	workFn := func() (TestVal, error) {
		callCount++
		return retVal, nil
	}

	key := "key"
	g := new(Group[TestVal])
	ch := g.DoChan(key, workFn)
	res := <-ch
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if res.Shared {
		t.Error("Expected a non-shared result")
	}
	if res.Val != retVal {
		t.Errorf("Expected %s but got %s", retVal, res.Val)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDoChanError(t *testing.T) {
	expectedErr := errors.New("this is expected")
	callCount := 0
	workFn := func() (TestVal, error) {
		callCount++
		return retVal, expectedErr
	}

	key := "key"
	g := new(Group[TestVal])
	ch := g.DoChan(key, workFn)
	res := <-ch
	if res.Err != nil && !errors.Is(res.Err, expectedErr) {
		t.Fatal(res.Err)
	}
	if !errors.Is(res.Err, expectedErr) {
		t.Error("Expected error, got none")
	}
	if res.Shared {
		t.Error("Expected a non-shared result")
	}
	if res.Val != retVal {
		t.Errorf("Expected %s but got %s", retVal, res.Val)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDoChanShared(t *testing.T) {
	workWg1 := new(sync.WaitGroup)
	workWg2 := new(sync.WaitGroup)
	workCh := make(chan int, 1)
	callCount := 0
	workFn := func() (TestVal, error) {
		callCount++
		if callCount == 1 {
			workWg1.Done()
		}
		v := <-workCh
		workCh <- v
		time.Sleep(10 * time.Millisecond)
		return retVal, nil
	}

	key := "key"
	g := new(Group[TestVal])
	readFn := func() {
		defer workWg2.Done()
		workWg1.Done()
		ch := g.DoChan(key, workFn)
		res := <-ch
		if res.Err != nil {
			t.Error(res.Err)
		}
		if res.Val != retVal {
			t.Errorf("Expected %s but got %s", retVal, res.Val)
		}
	}

	const max = 10
	workWg1.Add(1)
	for i := 0; i < max; i++ {
		workWg1.Add(1)
		workWg2.Add(1)
		go readFn()
	}
	workWg1.Wait()
	workCh <- 1
	workWg2.Wait()
	if callCount <= 0 || callCount >= max {
		t.Errorf("Expected between 1 and %d calls, got %d", max-1, callCount)
	}
}

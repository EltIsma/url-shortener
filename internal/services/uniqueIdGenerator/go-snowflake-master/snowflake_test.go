package snowflake

import (
	"errors"
	"sync"
	"testing"
	"time"

)

func TestID(t *testing.T) {
	id := ID()

	if id <= 0 {
		t.Error("The snowflake should't < 0.")
	}

	mp := make(map[uint64]bool)
	for i := 0; i < 100000; i++ {
		id, e := NextID()
		if e != nil {
			t.Error(e)
			continue
		}
		if _, ok := mp[id]; ok {
			t.Error("ID should't repeat", id)
			break
		}
		mp[id] = true
	}
}

func TestID_bitch(t *testing.T) {
	le := 100000
	ch := make(chan uint64, le)
	var wg sync.WaitGroup
	for i := 0; i < le; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := ID()
			ch <- id
		}()
	}
	wg.Wait()
	close(ch)

	mp := make(map[uint64]bool)
	for id := range ch {
		if _, ok := mp[id]; ok {
			t.Error("It should not be repeated")
			break
		}
		mp[id] = true
	}
	if len(mp) != le {
		t.Error("map length should be equal", le)
	}
}

func TestSetStartTime(t *testing.T) {
	t.Run("A nil time", func(tt *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				tt.Error("Should throw a error when start time is zero")
			} else if e.(string) != "The start time cannot be a zero value" {
				tt.Error("The error message should equal [The start time cannot be a zero value]")
			}
		}()
		var time time.Time
		SetStartTime(time)
	})

	t.Run("Start time too big", func(tt *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				tt.Error("Should throw a error when start time is too big")
			} else if e.(string) != "The s cannot be greater than the current millisecond" {
				tt.Error("The error message should equal [The s cannot be greater than the current millisecond]")
			}
		}()
		time := time.Date(2035, 1, 1, 1, 0, 0, 0, time.UTC)
		SetStartTime(time)
	})

	t.Run("Start time too small", func(tt *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				tt.Error("Should throw a error when starttime is too small")
			} else if e.(string) != "The maximum life cycle of the snowflake algorithm is 69 years" {
				tt.Error("The error message should equal [The maximum life cycle of the snowflake algorithm is 69 years]")
			}
		}()
		// because 2021-69 = 1952, set df time > 69 years to test.
		time := time.Date(1951, 1, 1, 1, 0, 0, 0, time.UTC)
		SetStartTime(time)
	})

	t.Run("Default start time", func(tt *testing.T) {
		defaultTime := time.Date(2008, 11, 10, 23, 0, 0, 0, time.UTC)
		defaultNano := defaultTime.UTC().UnixNano() / 1e6

		sid := ParseID(ID())
		currentTime := sid.Timestamp + uint64(defaultNano)

		nowNano := time.Now().UTC().UnixNano() / 1e6

		// approximate equality, Assuming that the program is completed in one second.
		if currentTime/1000 != uint64(nowNano)/1000 {
			t.Error("The timestamp should be equal")
		}
	})

	t.Run("Basic", func(tt *testing.T) {
		date := time.Date(2002, 1, 1, 1, 0, 0, 0, time.UTC)
		SetStartTime(date)

		nowNano := time.Now().UTC().UnixNano() / 1e6
		startNano := date.UTC().UnixNano() / 1e6
		df := nowNano - startNano

		sid := ParseID(ID())

		// approximate equality, Assuming that the program is completed in one second.
		if sid.Timestamp/1000 != uint64(df)/1000 {
			t.Error("The timestamp should be equal")
		}
	})
}

func TestSetMachineID(t *testing.T) {
	// first test,
	sid := ParseID(ID())
	if sid.MachineID != 0 {
		t.Error("MachineID should be equal 0")
	}

	t.Run("No Panic", func(tt *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				t.Error("An error should not be returned")
			}
		}()

		SetMachineID(1)
		id := ID()
		sid :=ParseID(id)

		if sid.MachineID != 1 {
			tt.Error("The machineID should be equal 1")
		}
	})

	t.Run("Panic", func(tt *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				tt.Error("Should throw a error")
			} else if err.(string) != "The machineID cannot be greater than 1023" {
				tt.Error("The error message should be eq 「The machineID cannot be greater than 1023」")
			}
		}()

		SetMachineID(1024)
	})

	SetMachineID(100)
	sid = ParseID(ID())
	if sid.MachineID != 100 {
		t.Error("MachineID should be equal 100")
	}
}

func TestSetSequenceResolver(t *testing.T) {
	SetSequenceResolver(func(c int64) (uint16, error) {
		return 100, nil
	})

	id := ID()
	sid := ParseID(id)

	if sid.Sequence != 100 {
		t.Error("The snowflake number part of sequence should be equal 100")
	}

	time.Sleep(time.Millisecond)

	id = ID()
	sid2 := ParseID(id)

	if sid2.Sequence != 100 {
		t.Error("The snowflake number part of sequence should be equal 100")
	}

	if sid2.Timestamp <= sid.Timestamp {
		t.Error("It should be bigger than the previous time")
	}
}

func TestNextID(t *testing.T) {
	_, err := NextID()
	if err != nil {
		t.Error(err)
		return
	}

	SetSequenceResolver(func(ms int64) (uint16, error) {
		return 0, errors.New("test error")
	})
	_, e := NextID()
	if e == nil {
		t.Error("Should be throw error")
	} else if e.Error() != "test error" {
		t.Error("NextID error message should be equal [test error]")
	}
}

func TestParseID(t *testing.T) {
	time := 101 << (MachineIDLength + SequenceLength)
	machineid := 1023 << SequenceLength
	seq := 999

	id := uint64(time | machineid | seq)

	d := ParseID(id)
	if d.Sequence != 999 {
		t.Error("Sequence should be equal 999")
	}

	if d.MachineID != 1023 {
		t.Error("MachineID should be equal 1023")
	}

	if d.Timestamp != 101 {
		t.Error("Timestamp should be equal 101")
	}
}

func TestSID_GenerateTime(t *testing.T) {
	SetSequenceResolver(AtomicResolver)
	a, e := NextID()
	if e != nil {
		t.Error(e)
		return
	}

	sid := ParseID(a)

	if sid.GenerateTime().UTC().Second() != time.Now().UTC().Second() {
		t.Error("The id generate time should be equal current time")
	}
}

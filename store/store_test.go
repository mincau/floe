package store

import "testing"

func TestLocalStore(t *testing.T) {
	ls, err := NewLocalStore("%tmp")
	if err != nil {
		t.Fatal(err)
	}
	testStore("local", t, ls)

	ms := NewMemStore()
	testStore("mem", t, ms)
}

func testStore(which string, t *testing.T, s Store) {
	key := "foo/bar"
	val := "test data"
	err := s.Save(key, val)
	if err != nil {
		t.Fatal(which, err)
	}

	var loadVal string
	err = s.Load(key, &loadVal)
	if err != nil {
		t.Fatal(which, err)
	}
	if loadVal != val {
		t.Errorf("%s bad load <%s> <%s>", which, loadVal, val)
	}

	var failVal int
	err = s.Load(key, &failVal)
	if err == nil {
		t.Fatal("mismatched val should have failed")
	}
	t.Log(err)
}

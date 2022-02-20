package periodic

import (
	"errors"
	"testing"
)

func TestCache_Get(t *testing.T) {
	cache := NewCache(testValueProvider{})
	getAndAssert(t, cache, "", false)
}

func TestCache_Update(t *testing.T) {
	cache := NewCache(testValueProvider{value: "foobar"})
	updateAndAssert(t, cache, false)
	getAndAssert(t, cache, "foobar", true)
}

func TestCache_UpdateError(t *testing.T) {
	cache := NewCache(testValueProvider{err: errors.New("boom")})
	updateAndAssert(t, cache, true)
	getAndAssert(t, cache, "", false)
}

type testValueProvider struct {
	value string
	err error
}

func (p testValueProvider) NewValue() (string, error) {
	return p.value, p.err
}

func getAndAssert(
	t *testing.T,
	cache *Cache,
	expectedValue string,
	expectedOK bool,
) {
	value, ok := cache.Get()
	if value != expectedValue {
		t.Fatalf(
			"Get() returned value == %v, expected %v",
			value,
			expectedValue,
		)
	}
	if ok != expectedOK {
		t.Fatalf("Get() returned ok == %v, expected %v", ok, expectedOK)
	}
}

func updateAndAssert(t *testing.T, cache *Cache, wantErr bool) {
	err := cache.Update()
	if (err != nil) != wantErr {
		t.Fatalf(
			"unexpected Update() return value: wantErr == %v, but err == %v",
			wantErr,
			err,
		)
	}
}

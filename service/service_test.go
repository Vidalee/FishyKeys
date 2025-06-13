package service

import (
	"fmt"
	"github.com/Vidalee/FishyKeys/internal/testutil"
	"testing"
)

func TestMain(m *testing.M) {
	var err error
	testDB, err = testutil.SetupTestDB()
	if err != nil {
		panic(fmt.Sprintf("failed to setup test db: %v", err))
	}
	defer func() {
		if err := testutil.TeardownTestDB(); err != nil {
			panic(fmt.Sprintf("failed to teardown test db: %v", err))
		}
	}()

	m.Run()
}

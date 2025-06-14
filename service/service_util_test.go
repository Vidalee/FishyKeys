package service

import (
	"fmt"
	"github.com/Vidalee/FishyKeys/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
)

var testDB *pgxpool.Pool

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

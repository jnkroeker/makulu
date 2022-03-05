// Package tests provides foundational support for running unit
// and integragtion tests.
package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"go.uber.org/zap"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// Configuration for running tests.
const (
	dbImage = "dgraph/standalone:master"
	dbPort  = "8080"
)

// NewUnit creates a test value with necessary application state to run
// database tests. It will return the host to use to connection to the database.
func NewUnit(t *testing.T) (*zap.SugaredLogger, string, func()) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	// Start a DB container instance with dgraph running.
	c := StartContainer(t, dbImage, dbPort)

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		t.Log("tearing down test ...")
		StopContainer(t, c.ID)

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = old
		fmt.Println("******************** LOGS ********************")
		fmt.Print(buf.String())
		fmt.Println("******************** LOGS ********************")
	}

	url := fmt.Sprintf("http://%s", c.Host)
	log, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println(err)
	}

	return log.Sugar(), url, teardown
}

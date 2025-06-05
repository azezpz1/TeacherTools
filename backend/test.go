// file: noop_test.go
package main

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func TestNoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// nothing else; just keeps the gomock import alive
}

package libmacaroon

import (
	"os"
	"runtime/pprof"
	"testing"
)

func writeHeapProfile(t *testing.T) {
	t.Helper()
	var err error
	filename := t.Name() + ".heap.out"
	var fd *os.File
	fd, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0o644))
	if err != nil {
		t.Errorf("writeHeapProfile: failed to craete memory profile file '%s': %v", filename, err)
		return
	}
	defer func() {
		if err == nil {
			return
		}
		// make sure its closed
		_ = fd.Close()
		// try to remove
		if rmErr := os.Remove(filename); rmErr != nil {
			t.Errorf("writeHeapProfile: failed to remove file '%s': %v", filename, err)
			return
		}
	}()
	if err = pprof.Lookup("heap").WriteTo(fd, 0); err != nil {
		t.Errorf("writeHeapProfile: failed to write heap profile '%s': %v", filename, err)
		return
	}
	if err = fd.Close(); err != nil {
		t.Errorf("writeHeapProfile: failed to close heap profile '%s': %v", filename, err)
		return
	}
	t.Logf("memory profile written to: %v", filename)
}

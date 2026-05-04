package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectOdooMajorVersions_Nested(t *testing.T) {
	tmp := t.TempDir()

	// Root manifest (depth 1)
	err := os.WriteFile(filepath.Join(tmp, "__manifest__.py"), []byte("'version': '18.0'"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Module manifest (depth 2)
	addonDir := filepath.Join(tmp, "addon")
	os.Mkdir(addonDir, 0755)
	os.WriteFile(filepath.Join(addonDir, "__manifest__.py"), []byte("'version': '19.0'"), 0644)

	// Nested module manifest (depth 3)
	nestedDir := filepath.Join(tmp, "modules", "nested")
	os.MkdirAll(nestedDir, 0755)
	os.WriteFile(filepath.Join(nestedDir, "__manifest__.py"), []byte("'version': '17.0'"), 0644)

	// Too deep manifest (depth 4) - should be ignored
	deepDir := filepath.Join(tmp, "a", "b", "c")
	os.MkdirAll(deepDir, 0755)
	os.WriteFile(filepath.Join(deepDir, "__manifest__.py"), []byte("'version': '14.0'"), 0644)

	versions, isOdoo, err := detectOdooMajorVersions(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if !isOdoo {
		t.Errorf("isOdoo = false, want true")
	}

	// We want it to find 17, 18, 19 but NOT 14.
	want := map[int]struct{}{17: {}, 18: {}, 19: {}}
	if !reflect.DeepEqual(versions, want) {
		t.Errorf("detectOdooMajorVersions() = %v, want %v", versions, want)
	}
}

func TestDetectOdooMajorVersions_Unparseable(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "__manifest__.py"), []byte("'version': 'invalid'"), 0644)

	versions, isOdoo, err := detectOdooMajorVersions(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if !isOdoo {
		t.Errorf("isOdoo = false, want true")
	}
	want := map[int]struct{}{-1: {}}
	if !reflect.DeepEqual(versions, want) {
		t.Errorf("detectOdooMajorVersions() = %v, want %v", versions, want)
	}
}

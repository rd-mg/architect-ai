package odoo

import (
    "os"
    "path/filepath"
    "testing"
)

func TestDetect_EmptyDir_NotOdoo(t *testing.T) {
    info, _ := Detect(t.TempDir())
    if info.IsOdoo {
        t.Error("empty dir must not be Odoo")
    }
}

func TestDetect_Manifest_V18(t *testing.T) {
    dir := t.TempDir()
    mod := filepath.Join(dir, "sale_custom")
    os.MkdirAll(mod, 0755)
    os.WriteFile(filepath.Join(mod, "__manifest__.py"),
        []byte(`{'name':'Sale Custom','version':'18.0.1.0.0','depends':['sale']}`), 0644)

    info, err := Detect(dir)
    if err != nil {
        t.Fatal(err)
    }
    if !info.IsOdoo || info.Version != "18" {
        t.Errorf("expected Odoo v18, got is_odoo=%v ver=%s", info.IsOdoo, info.Version)
    }
}

func TestDetect_Requirements(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("odoo>=18.0\nrequests\n"), 0644)
    info, _ := Detect(dir)
    if !info.IsOdoo {
        t.Error("must detect Odoo from requirements.txt")
    }
}

func TestExtractVersion_Variants(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "__manifest__.py")
    cases := []struct{ content, want string }{
        {`{'version': '19.0.1.0.0'}`, "19"},
        {`{"version": "17.0.2.1.0"}`, "17"},
        {`{'name': 'test'}`, "unknown"},
    }
    for _, tc := range cases {
        os.WriteFile(path, []byte(tc.content), 0644)
        got := extractVersion(path)
        if got != tc.want {
            t.Errorf("extractVersion content=%q got=%q want=%q", tc.content, got, tc.want)
        }
    }
}

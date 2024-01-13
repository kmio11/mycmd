package testutils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	root string
}

func NewTestData(t *testing.T, name string) *TestData {
	return &TestData{
		root: name,
	}
}

func (d *TestData) RootDir(t *testing.T) string {
	return filepath.Join("testdata", d.root)
}

func (d *TestData) FileName(t *testing.T, elm ...string) string {
	return filepath.Join(d.RootDir(t), filepath.Join(elm...))
}

func (d *TestData) WriteFile(t *testing.T, filename string, data []byte) {
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func (d *TestData) ReadFile(t *testing.T, filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func (d *TestData) CompareWithGolden(t *testing.T, update bool, goldenFileName string, actual []byte) {
	if update {
		d.WriteFile(t, goldenFileName, actual)
	}
	expected := d.ReadFile(t, goldenFileName)
	assert.Equal(t, string(expected), string(actual))
}

// TempDirInTestdata makes temp directory for test.
// If you need fixed name temp directory, use this method instead of testing.T.TempDir().
func (d *TestData) TempDirInTestdata(t *testing.T, suffix string) string {
	tmpDir := d.FileName(t, fmt.Sprintf(".tmp%s", suffix))
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Println("cleanup is faild : ", err)
		}
	})
	return tmpDir
}

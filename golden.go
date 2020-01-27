package golden

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// G is the object that assertions are made on.
type G struct {
	t  testing.TB
	fs afero.Fs

	ShouldUpdate bool

	FixtureDir    string
	FixturePrefix string
	FixtureSuffix string
}

// New creates a new, ready to use G.
func New(t testing.TB) G {
	return newG(t)
}

func newG(t testing.TB) G {
	return G{
		t:             t,
		fs:            afero.NewOsFs(),
		FixtureDir:    "testdata",
		FixturePrefix: "",
		FixtureSuffix: ".golden",
	}
}

// Assert asserts, that the given byte slice has the same content as a file,
// whose path is derived from the given name (usually
// "testdata/"+name+".golden").
func (g G) Assert(name string, got []byte) {
	if g.ShouldUpdate {
		err := g.write(name, got)
		assert.NoError(g.t, err)
	} else {
		err := g.readAndCompare(name, got)
		assert.NoError(g.t, err)
	}
}

// AssertStruct asserts, that the given struct has the same content as a file,
// whose path is derived from the given name (usually
// "testdata/"+name+".golden"). This is ensured by gob-encoding the given struct
// and comparing it against the file's contents.
func (g G) AssertStruct(name string, got interface{}) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(got); err != nil {
		g.t.Errorf("unable to encode instance of %T: %w", got, err)
	} else {
		g.Assert(name, buf.Bytes())
	}
}

func (g G) write(name string, data []byte) error {
	path := g.computeFilePath(name)
	return g.writeFile(path, data)
}

func (g G) readAndCompare(name string, got []byte) error {
	path := g.computeFilePath(name)
	want, err := g.readFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	assert.Equal(g.t, want, got)
	return nil
}

func (g G) readFile(path string) ([]byte, error) {
	file, err := g.fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = file.Close() }()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}

	return data, nil
}

func (g G) writeFile(path string, data []byte) error {
	err := g.fs.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return fmt.Errorf("mkdir all: %w", err)
	}

	file, err := g.fs.Create(path)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer func() { _ = file.Close() }()

	n, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("could only write %d of %d bytes", n, len(data))
	}

	return nil
}

func (g G) computeFilePath(name string) string {
	return filepath.Join(g.FixtureDir, g.FixturePrefix+name+g.FixtureSuffix)
}

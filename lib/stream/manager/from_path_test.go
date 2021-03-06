package manager

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/Jeffail/benthos/v3/lib/stream"
	yaml "gopkg.in/yaml.v3"
)

func TestFromPathHappy(t *testing.T) {
	testDir, err := ioutil.TempDir("", "streams_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	barDir := filepath.Join(testDir, "bar")
	if err = os.Mkdir(barDir, 0777); err != nil {
		t.Fatal(err)
	}

	fooPath := filepath.Join(testDir, "foo.yml")
	barPath := filepath.Join(barDir, "test.yaml")
	ignorePath := filepath.Join(barDir, "test_benthos_test.yaml")

	fooConf := stream.NewConfig()
	fooConf.Input.Type = "TEST_FOO"

	barConf := stream.NewConfig()
	barConf.Input.Type = "TEST_BAR"

	expConfs := map[string]stream.Config{
		"foo":      fooConf,
		"bar_test": barConf,
	}

	var fooBytes []byte
	if fooBytes, err = json.Marshal(fooConf); err != nil {
		t.Fatal(err)
	}
	var barBytes []byte
	if barBytes, err = yaml.Marshal(barConf); err != nil {
		t.Fatal(err)
	}

	if err = ioutil.WriteFile(fooPath, fooBytes, 0666); err != nil {
		t.Fatal(err)
	}
	if err = ioutil.WriteFile(barPath, barBytes, 0666); err != nil {
		t.Fatal(err)
	}
	if err = ioutil.WriteFile(ignorePath, []byte("ignore me please"), 0666); err != nil {
		t.Fatal(err)
	}

	actConfs := map[string]stream.Config{}
	var lints []string
	if lints, err = LoadStreamConfigsFromPath(testDir, "benthos_test", actConfs); err != nil {
		t.Fatal(err)
	}

	if len(lints) == 0 {
		t.Error("Expected lint errors")
	}

	var actKeys, expKeys []string
	for id := range actConfs {
		actKeys = append(actKeys, id)
	}
	sort.Strings(actKeys)
	for id := range expConfs {
		expKeys = append(expKeys, id)
	}
	sort.Strings(expKeys)

	if !reflect.DeepEqual(actKeys, expKeys) {
		t.Errorf("Wrong keys in loaded set: %v != %v", actKeys, expKeys)
	}

	if exp, act := "TEST_FOO", actConfs["foo"].Input.Type; exp != act {
		t.Errorf("Wrong value in loaded set: %v != %v", act, exp)
	}
	if exp, act := "TEST_BAR", actConfs["bar_test"].Input.Type; exp != act {
		t.Errorf("Wrong value in loaded set: %v != %v", act, exp)
	}
}

package sender

import (
	"io/ioutil"
	os "os"
	"path"
	"strconv"
	"testing"
	"time"
)

const (
	FixtureInterval = int64(30)
	FixtureFiles    = 10
	FixtureLines    = 55
	TestContent     = "THIS IS SOME DEMO CONTENT FOR A FILE"
)

func createFixture(t *testing.T) string {
	temp := os.TempDir()

	tempPath, err := ioutil.TempDir(temp, "filespooler")
	if err != nil {
		t.Fatal("could not create temp path: ", err)
	}

	ts := time.Now().Unix()

	for range make([]int, FixtureFiles) {
		ts -= FixtureInterval

		fileName := "spool-" + strconv.FormatInt(ts, 10)
		content := ""

		for range make([]int, FixtureLines) {
			content += TestContent + "\n"
		}

		writeFile(t, tempPath, fileName, content)
	}

	// extra files that should not be parsed
	_ = os.Mkdir(path.Join(tempPath, "directory"), 0755)
	writeFile(t, tempPath, ".gitignore", "# nothing")

	return tempPath
}

func writeFile(t *testing.T, directory string, name string, content string) {
	filePath := path.Join(directory, name)
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Could not write test file: %s - %s", filePath, err)
	}
}

func TestNewFileReader(t *testing.T) {
	spool := createFixture(t)
	defer func() {
		_ = os.RemoveAll(spool)
	}()

	r, err := NewFileReader(spool)
	if err != nil {
		t.Fatal(err)
	}

	files, err := r.ReadDir()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != FixtureFiles {
		var fileList []string
		for _, file := range files {
			fileList = append(fileList, file.RawName)
		}

		t.Fatalf("Found more files than expected in %s - %v", spool, fileList)
	}

	for _, file := range files {
		if len(file.Content()) == 0 {
			t.Fatalf("File RawContent is empty: %s", file.RawName)
		}

		err := r.Delete(file.RawName)
		if err != nil {
			t.Fatal(err)
		}
	}

	// check that directory is empty (expect for dir and dotfile)
	leftFiles, err := ioutil.ReadDir(spool)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range leftFiles {
		if file.IsDir() || file.Name()[0:1] == "." {
			continue
		}

		t.Fatal("Found left over file: ", file.Name())
	}
}

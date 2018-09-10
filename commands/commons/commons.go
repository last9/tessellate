package commons

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/h2non/filetype.v1"
)

func defaultBlackList() []string {
	return []string{"tfvars"}
}

func defaultManifest() []string {
	return []string{
		".tf.json",
	}
}

// Gets the actual file type, irrespective of the extension provided to the file.
func GetFileContentType(path string) (string, error) {
	// Open a file descriptor
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	// We only have to pass the file header = first 261 bytes
	head := make([]byte, 261)
	file.Read(head)

	fileType, err := filetype.Get(head)
	return fileType.Extension, nil
}

func ReadFileLines(file string) ([]string, error) {
	lines := []string{}

	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) > 0 {
					lines = append(lines, line)
				}
				break
			}

			return nil, err
		}

		lines = append(lines, line)
	}
	return lines, err
}

// CandidateFiles matches files that should be uploaed or not
func CandidateFiles(dirname string, manifest []string) ([]string, error) {
	if manifest == nil {
		manifest = defaultManifest()
	}
	// Always append defaultManifest() list to actual input, as we would like to parse all tf.json files as well.
	// Chances are, tf.json might not be written in the manifest file.
	manifest = append(manifest, defaultManifest()...)

	blacklist := defaultBlackList()
	var files []string
	if err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		for _, b := range blacklist {
			if strings.Contains(path, b) {
				log.Printf("skipping %s", path)
				return nil
			}
		}

		for _, m := range manifest {
			if strings.HasSuffix(path, m) {
				fType, err := GetFileContentType(path)
				if err != nil {
					log.Printf("cannot read file header, %v", err)
				}
				if fType == filetype.Unknown.Extension {
					files = append(files, path)
				}
				break
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

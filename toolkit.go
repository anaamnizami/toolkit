package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentracing/opentracing-go/log"
)

const SourceString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!_$123456789"

type Toolkit struct {
	MaxFileSize  int
	AllowedTypes []string
}

func (t *Toolkit) RandomString(num int) string {
	s, r := make([]rune, num), []rune(SourceString)
	for i := range s {
		n, _ := rand.Prime(rand.Reader, len(r))
		x, y := n.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

type UploadedFile struct {
	OrignalFileName  string
	UploadedFileName string
	FileSize         int64
}

func (t *Toolkit) UploadOneFile(r *http.Request, destFolder string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}
	files, err := t.UploadFiles(r, destFolder, renameFile)
	if err != nil {
		return nil, err
	}
	return files[0], nil
}

func (t *Toolkit) UploadFiles(r *http.Request, destFolder string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}
	var UploadedFiles []*UploadedFile
	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}
	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		log.Error(err)
		return nil, errors.New("uploaded file is too big")
	}
	for _, fheader := range r.MultipartForm.File {
		for _, hdr := range fheader {
			UploadedFiles, err = func(UploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var UploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()
				buff := make([]byte, 512)
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				//TODO check if the filetype is permitted
				allowed := true
				fileType := http.DetectContentType(buff)
				// allowedTypes := []string{"image/jpg, image/png, image/gif"}
				if len(t.AllowedTypes) > 0 {
					for _, x := range t.AllowedTypes {
						if strings.EqualFold(fileType, x) {
							allowed = true
						}
					}
				} else {
					allowed = false
				}

				if !allowed {
					return nil, errors.New("file type is not permitted")
				}
				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				if renameFile {
					UploadedFile.UploadedFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					UploadedFile.UploadedFileName = hdr.Filename
				}
				UploadedFile.OrignalFileName = hdr.Filename
				var outfile *os.File
				defer outfile.Close()
				if outfile, err := os.Create(filepath.Join(destFolder, UploadedFile.UploadedFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					UploadedFile.FileSize = fileSize
				}
				UploadedFiles = append(UploadedFiles, &UploadedFile)
				return UploadedFiles, nil
			}(UploadedFiles)
			if err != nil {
				return UploadedFiles, err
			}
		}
	}
	return UploadedFiles, err
}

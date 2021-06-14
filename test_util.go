package gobits

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDataFilePath = "testdata/Lenna.jpg"
)

func setupTestDataFile(t *testing.T) (io.ReadWriteSeeker, func()) {
	file, err := os.OpenFile(testDataFilePath, os.O_RDWR, 644)
	assert.Nil(t, err)
	assert.NotNil(t, file)
	return file, func() {
		file.Close()
	}
}

func backupTestDataFile(sourceFile io.Reader) {
	backupFIle, _ := os.Create(testDataFilePath + ".bak")
	defer backupFIle.Close()
	io.Copy(backupFIle, sourceFile)
}

func recoverDataFile() {
	os.Remove(testDataFilePath)
	os.Rename(testDataFilePath+".bak", testDataFilePath)
}

func rawAt(rwseeker io.ReadWriteSeeker, byteOffset int64) byte {
	buffer := make([]byte, 1)
	if _, err := rwseeker.Seek(byteOffset, 0); err != nil {
		return 0
	}
	_, err := rwseeker.Read(buffer)
	if err != nil {
		return 0
	}
	return buffer[0]
}

func rawSlice(rwseeker io.ReadWriteSeeker, byteOffset int64, lenght int64) []byte {
	buffer := make([]byte, lenght)
	if _, err := rwseeker.Seek(byteOffset, 0); err != nil {
		return nil
	}
	actualLengh, err := rwseeker.Read(buffer)
	if err != nil {
		return nil
	}
	return buffer[:actualLengh]
}

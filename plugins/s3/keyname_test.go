package s3

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUUIDName(t *testing.T) {
	fileTypes := ValidFileTypeExtensions()
	for _, ft := range fileTypes {
		fmt.Println(ft)
		name, err := UUIDName(ft)
		if err != nil {
			assert.Fail(t, "UUID Not generated")
		}
		extLength := len(ft)
		if ft == "" {
			// length of uuid 27e78fb6-8e6c-4fd9-b368-d7eed7705165
			assert.Len(t, name, 36+extLength)
		} else {
			// length of uuid 27e78fb6-8e6c-4fd9-b368-d7eed7705165 + "." + file extension
			assert.Len(t, name, 37+extLength)
		}
	}
}

func TestTimeStampName(t *testing.T) {
	fileTypes := ValidFileTypeExtensions()
	//2019-09-27 01:19:18 +0000 UTC
	tsTime := "1569547158"
	tsNow := time.Unix(int64(1569547158), 0)
	for _, ft := range fileTypes {
		name := TimeStampName(tsNow, ft)
		if ft == "" {
			// length of unix timestamp which is set in the TimeStampName function
			assert.Equal(t, name, tsTime)
		} else {
			//length of unix timestamp + "." + file extension
			fName := fmt.Sprintf("%s.%s", tsTime, ft)
			assert.Equal(t, name, fName)
		}
	}

}

func TestTSDBKeyNameUUID(t *testing.T) {
	encoder := TSDBEncoder{
		FileNameExtension: noExt,
		FileNameType:      fileNameUUID,
		FileNameStructure: noFolderStructure,
		Compress:          false,
	}
	fName, err := encoder.KeyName("localhost")
	if err != nil {
		assert.Fail(t, "Failed to generate keyname for uuid without extension")
	}
	assert.Len(t, fName, 36)
}

func TestTSDBKeyNameUUIDCompressed(t *testing.T) {
	encoder := TSDBEncoder{
		FileNameExtension: noExt,
		FileNameType:      fileNameUUID,
		FileNameStructure: noFolderStructure,
		Compress:          true,
	}
	fName, err := encoder.KeyName("localhost")
	if err != nil {
		assert.Fail(t, "Failed to generate keyname for uuid without extension")
	}
	assert.Len(t, fName, 36+3)
}

func TestTSDBKeyNameUUIDWithFolder(t *testing.T) {
	encoder := TSDBEncoder{
		FileNameExtension: noExt,
		FileNameType:      fileNameUUID,
		FileNameStructure: dirDateHostFolderStructure,
		Compress:          false,
	}
	var tsTime int64 = 1569547158
	tsNow := time.Unix(tsTime, 0)
	filename := RunTestTimeInFolders(t, encoder.FileNameStructure, encoder.FileNameType, encoder.FileNameExtension, encoder.Compress, tsNow)
	assert.Len(t, filename, 57)
}

func TestTSDBKeyNameUUIDWithFolderCompressed(t *testing.T) {
	encoder := TSDBEncoder{
		FileNameExtension: noExt,
		FileNameType:      fileNameUUID,
		FileNameStructure: dirDateHostFolderStructure,
		Compress:          true,
	}
	var tsTime int64 = 1569547158
	tsNow := time.Unix(tsTime, 0)
	filename := RunTestTimeInFolders(t, encoder.FileNameStructure, encoder.FileNameType, encoder.FileNameExtension, encoder.Compress, tsNow)
	assert.Len(t, filename, 60)
}

func TestCSVKeyNameUUID(t *testing.T) {
	encoder := CSVEncoder{
		FileNameExtension: tsvExt,
		FileNameType:      fileNameUUID,
		FileNameStructure: noFolderStructure,
		Compress:          false,
	}
	fName, err := encoder.KeyName("localhost")
	if err != nil {
		assert.Fail(t, "Failed to generate keyname for uuid without extension")
	}
	assert.Len(t, fName, 36+4)
}

func TestCSVKeyNameUUIDCompressed(t *testing.T) {
	encoder := CSVEncoder{
		FileNameExtension: tsvExt,
		FileNameType:      fileNameUUID,
		FileNameStructure: noFolderStructure,
		Compress:          true,
	}
	fName, err := encoder.KeyName("localhost")
	if err != nil {
		assert.Fail(t, "Failed to generate keyname for uuid without extension")
	}
	assert.Len(t, fName, 36+7)
}

func TestCSVKeyNameUUIDWithFolder(t *testing.T) {
	encoder := CSVEncoder{
		FileNameExtension: tsvExt,
		FileNameType:      fileNameUUID,
		FileNameStructure: dirDateHostFolderStructure,
		Compress:          false,
	}
	var tsTime int64 = 1569547158
	tsNow := time.Unix(tsTime, 0)
	filename := RunTestTimeInFolders(t, encoder.FileNameStructure, encoder.FileNameType, encoder.FileNameExtension, encoder.Compress, tsNow)
	assert.Len(t, filename, 61)
}

func TestCSVKeyNameUUIDWithFolderCompressed(t *testing.T) {
	encoder := CSVEncoder{
		FileNameExtension: tsvExt,
		FileNameType:      fileNameUUID,
		FileNameStructure: dirDateHostFolderStructure,
		Compress:          true,
	}
	var tsTime int64 = 1569547158
	tsNow := time.Unix(tsTime, 0)
	filename := RunTestTimeInFolders(t, encoder.FileNameStructure, encoder.FileNameType, encoder.FileNameExtension, encoder.Compress, tsNow)
	assert.Len(t, filename, 64)
}

func TestCSVKeyNameTimeStamp(t *testing.T) {
	encoder := CSVEncoder{
		FileNameExtension: tsvExt,
		FileNameType:      fileNameTimeStamp,
		FileNameStructure: noFolderStructure,
		Compress:          false,
	}
	fName, err := encoder.KeyName("localhost")
	if err != nil {
		assert.Fail(t, "Failed to generate keyname for uuid without extension")
	}
	assert.Len(t, fName, 14)
}

func TestCSVKeyNameTimeStampCompressed(t *testing.T) {
	encoder := CSVEncoder{
		FileNameExtension: tsvExt,
		FileNameType:      fileNameTimeStamp,
		FileNameStructure: noFolderStructure,
		Compress:          true,
	}
	fName, err := encoder.KeyName("localhost")
	if err != nil {
		assert.Fail(t, "Failed to generate keyname for uuid without extension")
	}
	assert.Len(t, fName, 17)
}

func TestCSVKeyNameTimeStampWithFolder(t *testing.T) {
	encoder := CSVEncoder{
		FileNameExtension: tsvExt,
		FileNameType:      fileNameTimeStamp,
		FileNameStructure: dirDateHostFolderStructure,
		Compress:          false,
	}
	var tsTime int64 = 1569547158
	tsNow := time.Unix(tsTime, 0)
	filename := RunTestTimeInFolders(t, encoder.FileNameStructure, encoder.FileNameType, encoder.FileNameExtension, encoder.Compress, tsNow)
	assert.Len(t, filename, 35)
}

func TestCSVKeyNameTimeStampWithFolderCompressed(t *testing.T) {
	encoder := CSVEncoder{
		FileNameExtension: tsvExt,
		FileNameType:      fileNameTimeStamp,
		FileNameStructure: dirDateHostFolderStructure,
		Compress:          true,
	}
	var tsTime int64 = 1569547158
	tsNow := time.Unix(tsTime, 0)
	filename := RunTestTimeInFolders(t, encoder.FileNameStructure, encoder.FileNameType, encoder.FileNameExtension, encoder.Compress, tsNow)
	assert.Len(t, filename, 38)
}

func RunTestTimeInFolders(t *testing.T, fileNameStructure, fileNameType, fileNameExtension string, compress bool, tsNow time.Time) string {
	// This should generate this "2019/09/26/localhost/random uuid goes here"
	fName, err := KeyName("localhost", fileNameStructure, fileNameType, fileNameExtension, compress, tsNow)
	if err != nil {
		assert.Failf(t, "Failed to generate keyname %s", err.Error())
	}
	results := strings.Split(fName, "/")
	assert.Equal(t, 5, len(results), "Expected %#v to contain 5 parts", results)

	year, err := strconv.Atoi(results[0])
	assert.NoError(t, err)
	month, err := strconv.Atoi(results[1])
	assert.NoError(t, err)
	day, err := strconv.Atoi(results[2])
	assert.NoError(t, err)

	hostname2 := results[3]

	sameYear := year == int(time.Now().Year()) ||
		year == int(tsNow.Year())
	sameMonth := month == int(time.Now().Month()) ||
		month == int(tsNow.Month())
	sameDay := day == int(time.Now().Day()) ||
		day == int(tsNow.Day())

	assert.True(t, sameYear, "Expected year %s and received %s", tsNow.Year(), year)
	assert.True(t, sameMonth, "Expected month %s and received %s", tsNow.Month(), month)
	assert.True(t, sameDay, "Expected day %d and received %s", tsNow.Day(), day)

	assert.Equal(t, "localhost", hostname2)
	return fName

}

package s3

import (
	"fmt"
	"path"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	jsonExt  = "json"
	csvExt   = "csv"
	tsvExt   = "tsv"
	gzExt    = "gz"
	tsvGzExt = "tsv.gz"
	noExt    = ""
)
const (
	dirDateHostFolderStructure = "date_host"
	noFolderStructure          = ""
	fileNameUUID               = "uuid"
	fileNameTimeStamp          = "timestamp"
)

func ValidFileTypeExtensions() [3]string {
	return [3]string{"tsv", "csv", ""}
}

func ValidFileNameTypes() [2]string {
	return [2]string{"uuid", "timestamp"}
}

func ValidFolderStructures() [2]string {
	return [2]string{"date_host", "none"}
}

func TimeStampName(t time.Time, ft string) string {
	if ft == "" {
		return strconv.FormatInt(t.Unix(), 10)
	} else {
		return strconv.FormatInt(t.Unix(), 10) + "." + string(ft)
	}
}

func UUIDName(ft string) (string, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("Something went wrong: %s", err)
	}
	if ft == "" {
		return uuid.String(), err
	} else {
		return fmt.Sprintf("%s.%s", uuid.String(), string(ft)), err
	}
}

func KeyName(hostname, fileNameStructure, fileNameType, fileNameExtension string, compress bool, tNow time.Time) (string, error) {
	var fullPath string
	var keyName string
	var err error
	if fileNameType == fileNameUUID {
		keyName, err = UUIDName(fileNameExtension)
		if err != nil {
			return fullPath, err
		}
	} else if fileNameType == fileNameTimeStamp {
		keyName = TimeStampName(tNow, fileNameExtension)
	} else {
		return fullPath, fmt.Errorf("Unsupported Filetype %s", fileNameType)
	}
	if compress == true {
		keyName = fmt.Sprintf("%s.%s", keyName, gzExt)
	}
	if fileNameStructure == dirDateHostFolderStructure {
		fullPath = path.Join(tNow.Format("2006/01/02"), hostname, keyName)
	} else {
		fullPath = keyName
	}
	return fullPath, err
}

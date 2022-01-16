package domain

import (
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	partFilenamePiece = "_part_"
	metaFilenamePiece = "_meta"
)

func init() {

}

var uuidRegexp = regexp.MustCompile("^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$")

func IsCorrectUuid(str string) bool {
	return uuidRegexp.MatchString(str)
}

func ChunkFileName(uid string, num int) string {
	return uid + partFilenamePiece + strconv.Itoa(num)
}

func MetaFileName(uid string) string {
	return uid + metaFilenamePiece
}

func ExtractUuidFromPartName(fn string) (string, error) {
	pos := strings.Index(fn, partFilenamePiece)
	if pos < 32 {
		return "", exceptions.NewApiError(http.StatusBadRequest, "incorrect part name")
	}
	return fn[:pos], nil
}

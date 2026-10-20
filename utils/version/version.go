package version

import (
	"cmp"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

// CheckUpdate 检查更新
func CheckUpdate() (bool, string) {
	response, err := http.Get(types.AppCheckUpdateUrl)
	if err != nil {
		return false, ""
	}
	defer response.Body.Close()

	jsonBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false, ""
	}

	tag, err := jsonparser.GetString(jsonBytes, "tag_name")
	if err != nil {
		return false, ""
	}

	return CompareVersion(tag, types.AppVersion, false), tag
}

func CompareVersion(v1, v2 string, equal bool) bool {
	var (
		v1IsDev = strings.HasSuffix(v1, "-dev")
		v2IsDev = strings.HasSuffix(v2, "-dev")
	)
	if v1IsDev && !v2IsDev {
		return true
	}
	if !v1IsDev && v2IsDev {
		return false
	}

	v1 = strings.Trim(v1, "v")
	v2 = strings.Trim(v2, "v")
	if equal && v1 == v2 {
		return true
	}
	if v1 != "" && v2 == "" {
		return true
	}

	v1Arr := strings.Split(v1, ".")
	v2Arr := strings.Split(v2, ".")
	for i := 0; i < 3 && i < len(v1Arr) && i < len(v2Arr); i++ {
		if c := comparePart(v1Arr[i], v2Arr[i]); c != 0 {
			return c > 0
		}
	}
	return false
}

// comparePart compares one part of two versions by its leading number, so
// that 10 comes after 9, and then by whatever follows the number.
func comparePart(a, b string) int {
	na, ra := leadingNumber(a)
	nb, rb := leadingNumber(b)
	if na != nb {
		return cmp.Compare(na, nb)
	}
	return strings.Compare(ra, rb)
}

func leadingNumber(s string) (int, string) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	n, _ := strconv.Atoi(s[:i])
	return n, s[i:]
}

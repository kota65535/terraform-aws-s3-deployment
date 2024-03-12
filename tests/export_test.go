package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func readJson(t *testing.T, path string) map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	bytes, err := os.ReadFile(path)
	require.NoError(t, err)
	err = json.Unmarshal(bytes, &ret)
	require.NoError(t, err)
	return ret
}

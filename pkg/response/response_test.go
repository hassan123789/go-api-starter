package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestError(t *testing.T) {
	c, rec := setupTestContext()

	err := Error(c, http.StatusBadRequest, "something went wrong")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "something went wrong", resp.Error)
}

func TestBadRequest(t *testing.T) {
	c, rec := setupTestContext()

	err := BadRequest(c, "invalid input")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "invalid input", resp.Error)
}

func TestUnauthorized(t *testing.T) {
	c, rec := setupTestContext()

	err := Unauthorized(c, "not authenticated")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "not authenticated", resp.Error)
}

func TestNotFound(t *testing.T) {
	c, rec := setupTestContext()

	err := NotFound(c, "resource not found")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "resource not found", resp.Error)
}

func TestInternalServerError(t *testing.T) {
	c, rec := setupTestContext()

	err := InternalServerError(c, "server error")
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "server error", resp.Error)
}

func TestSuccess(t *testing.T) {
	c, rec := setupTestContext()

	data := map[string]string{"name": "test"}
	err := Success(c, http.StatusOK, data)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "test", resp["name"])
}

func TestCreated(t *testing.T) {
	c, rec := setupTestContext()

	data := map[string]int{"id": 1}
	err := Created(c, data)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]int
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 1, resp["id"])
}

func TestOK(t *testing.T) {
	c, rec := setupTestContext()

	data := []string{"item1", "item2"}
	err := OK(c, data)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []string
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Len(t, resp, 2)
	assert.Equal(t, "item1", resp[0])
}

func TestNoContent(t *testing.T) {
	c, rec := setupTestContext()

	err := NoContent(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}

// Test struct types
func TestErrorResponseStruct(t *testing.T) {
	resp := ErrorResponse{
		Error:   "error message",
		Message: "additional details",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled ErrorResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, resp.Error, unmarshaled.Error)
	assert.Equal(t, resp.Message, unmarshaled.Message)
}

func TestSuccessResponseStruct(t *testing.T) {
	resp := SuccessResponse{
		Data:    map[string]int{"count": 5},
		Message: "success",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled SuccessResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, resp.Message, unmarshaled.Message)
	assert.NotNil(t, unmarshaled.Data)
}

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCategory(t *testing.T) {
	admin := adminClient(t)

	t.Run("happy path", func(t *testing.T) {
		code := uniqueCode()
		status, raw, err := admin.do(http.MethodPost, "/admin/categories", createCategoryBody{Code: code, Name: "Shoes"})
		require.NoError(t, err)
		requireStatus(t, http.StatusCreated, status, raw)
		cat := decode[categoryView](t, raw)
		trackCategory(cat.ID)
		assert.Equal(t, code, cat.Code)
		assert.Equal(t, "Shoes", cat.Name)
		assert.True(t, cat.Active)
	})

	t.Run("duplicate code → 409", func(t *testing.T) {
		code := uniqueCode()
		status, raw, err := admin.do(http.MethodPost, "/admin/categories", createCategoryBody{Code: code, Name: "First"})
		require.NoError(t, err)
		requireStatus(t, http.StatusCreated, status, raw)
		trackCategory(decode[categoryView](t, raw).ID)

		status, raw, err = admin.do(http.MethodPost, "/admin/categories", createCategoryBody{Code: code, Name: "Second"})
		require.NoError(t, err)
		requireStatus(t, http.StatusConflict, status, raw)
	})

	t.Run("invalid code → 400", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/categories", createCategoryBody{Code: "INVALID CODE", Name: "x"})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})
}

func TestListCategories(t *testing.T) {
	admin := adminClient(t)
	created := createCategory(t, admin)

	t.Run("includes seeded and created categories", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/categories", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		list := decode[listCategoriesView](t, raw)

		codes := make(map[string]bool, len(list.Categories))
		ids := make(map[string]bool, len(list.Categories))
		for _, c := range list.Categories {
			codes[c.Code] = true
			ids[c.ID] = true
		}
		assert.True(t, codes["clothing"], "seed category clothing missing")
		assert.True(t, codes["accessories"], "seed category accessories missing")
		assert.True(t, ids[created.ID], "created category missing")
	})

	t.Run("any authenticated user can list", func(t *testing.T) {
		user := newUserClient(t)
		status, raw, err := user.do(http.MethodGet, "/categories", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
	})
}

func TestUpdateCategory(t *testing.T) {
	admin := adminClient(t)
	cat := createCategory(t, admin)

	status, raw, err := admin.do(http.MethodPut, "/admin/categories/"+cat.ID, updateCategoryBody{Name: "Renamed", Active: false})
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	updated := decode[categoryView](t, raw)
	assert.Equal(t, "Renamed", updated.Name)
	assert.False(t, updated.Active)
}

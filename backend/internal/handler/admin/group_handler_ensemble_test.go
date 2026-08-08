//go:build unit

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type ensembleCreateGroupProbe struct {
	service.AdminService
	input          *service.CreateGroupInput
	updated        *service.UpdateGroupInput
	updatedMember  *service.EnsembleProposerInput
	existingMember []service.EnsembleProposer
	ensembleConfig service.EnsembleConfig
	updatedConfig  *service.EnsembleConfig
}

func (s *ensembleCreateGroupProbe) CreateGroup(_ context.Context, input *service.CreateGroupInput) (*service.Group, error) {
	s.input = input
	return &service.Group{ID: 42, Name: input.Name, Platform: input.Platform, Status: service.StatusActive}, nil
}

func (s *ensembleCreateGroupProbe) UpdateGroup(_ context.Context, _ int64, input *service.UpdateGroupInput) (*service.Group, error) {
	s.updated = input
	return &service.Group{ID: 42, Name: input.Name, Platform: input.Platform, Status: service.StatusActive}, nil
}

func (s *ensembleCreateGroupProbe) ListEnsembleProposers(context.Context, int64) ([]service.EnsembleProposer, error) {
	return s.existingMember, nil
}

func (s *ensembleCreateGroupProbe) UpdateEnsembleProposer(_ context.Context, _, _ int64, input service.EnsembleProposerInput) (*service.EnsembleProposer, error) {
	s.updatedMember = &input
	return &service.EnsembleProposer{ID: 8, GroupID: 42, Role: input.Role, Model: input.Model, Enabled: input.Enabled}, nil
}

func (s *ensembleCreateGroupProbe) GetEnsembleConfig(context.Context, int64) (service.EnsembleConfig, error) {
	return s.ensembleConfig, nil
}

func (s *ensembleCreateGroupProbe) UpdateEnsembleConfig(_ context.Context, _ int64, config service.EnsembleConfig) (service.EnsembleConfig, error) {
	s.updatedConfig = &config
	s.ensembleConfig = config
	return config, nil
}

func TestCreateGroupAcceptsEnsemblePlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	probe := &ensembleCreateGroupProbe{}
	router := gin.New()
	router.POST("/api/v1/admin/groups", NewGroupHandler(probe, nil, nil).Create)

	body, err := json.Marshal(map[string]any{
		"name":     "ensemble-test",
		"platform": service.PlatformEnsemble,
	})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/groups", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, probe.input)
	require.Equal(t, service.PlatformEnsemble, probe.input.Platform)
}

func TestUpdateGroupAcceptsEnsemblePlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	probe := &ensembleCreateGroupProbe{}
	router := gin.New()
	router.PUT("/api/v1/admin/groups/:id", NewGroupHandler(probe, nil, nil).Update)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/admin/groups/42", strings.NewReader(`{"platform":"ensemble"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, probe.updated)
	require.Equal(t, service.PlatformEnsemble, probe.updated.Platform)
}

func TestUpdateEnsembleProposerPreservesOmittedRoleAndEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	probe := &ensembleCreateGroupProbe{existingMember: []service.EnsembleProposer{{
		ID: 8, GroupID: 42, Role: service.EnsembleRoleAggregator, Model: "gpt-5", Enabled: false,
	}}}
	router := gin.New()
	router.PUT("/api/v1/admin/groups/:id/ensemble-proposers/:proposer_id", NewGroupHandler(probe, nil, nil).UpdateEnsembleProposer)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/admin/groups/42/ensemble-proposers/8", strings.NewReader(`{"model":"gpt-5.1"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, probe.updatedMember)
	require.Equal(t, service.EnsembleRoleAggregator, probe.updatedMember.Role)
	require.False(t, probe.updatedMember.Enabled)
}

func TestEnsembleConfigHandlerRoundTripsSourceGroupIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	probe := &ensembleCreateGroupProbe{ensembleConfig: service.EnsembleConfig{SourceGroupIDs: []int64{10, 20}}}
	router := gin.New()
	router.GET("/api/v1/admin/groups/:id/ensemble-config", NewGroupHandler(probe, nil, nil).GetEnsembleConfig)
	router.PUT("/api/v1/admin/groups/:id/ensemble-config", NewGroupHandler(probe, nil, nil).UpdateEnsembleConfig)

	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/admin/groups/42/ensemble-config", nil))
	require.Equal(t, http.StatusOK, getRecorder.Code)
	require.Contains(t, getRecorder.Body.String(), `"source_group_ids":[10,20]`)

	putRecorder := httptest.NewRecorder()
	putRequest := httptest.NewRequest(http.MethodPut, "/api/v1/admin/groups/42/ensemble-config", strings.NewReader(`{"source_group_ids":[30,40],"min_proposers":1}`))
	putRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(putRecorder, putRequest)
	require.Equal(t, http.StatusOK, putRecorder.Code)
	require.NotNil(t, probe.updatedConfig)
	require.Equal(t, []int64{30, 40}, probe.updatedConfig.SourceGroupIDs)
	require.Contains(t, putRecorder.Body.String(), `"source_group_ids":[30,40]`)
}

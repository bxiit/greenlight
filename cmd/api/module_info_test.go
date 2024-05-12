package main

import (
	"encoding/json"
	"fmt"
	"github.com/bxiit/greenlight/internal/data"
	"github.com/bxiit/greenlight/internal/jsonlog"
	"github.com/bxiit/greenlight/internal/mailer"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type InfosTestSuite struct {
	suite.Suite
	Server *http.Server
	Client *http.Client
	App    *application
}

func TestInfosTestSuite(t *testing.T) {
	suite.Run(t, &InfosTestSuite{})
}

func (s *InfosTestSuite) SetupSuite() {
	var cfg Config
	viper.SetConfigFile("./config.json")
	err := viper.ReadInConfig()
	s.Nil(err)

	err = viper.Unmarshal(&cfg)
	s.Nil(err)

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db), // data.NewModels() function to initialize App Models struct
		mailer: mailer.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Sender),
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: app.routes(),
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			_ = fmt.Errorf("failed to listen")
		}
	}()
	s.Server = server
	s.Client = &http.Client{}
	s.App = app
}

func (s *InfosTestSuite) TestModuleInfoModel_Insert() {
	// Create a new ModuleInfo instance with the desired values
	moduleInfo := &data.ModuleInfo{
		ModuleName:     "Test Module",
		ModuleDuration: 60000000000,
		ExamType:       "Final",
	}

	moduleInfoJson, err := json.Marshal(moduleInfo)
	s.Nil(err)
	request, err := http.NewRequest(http.MethodPost, "http://localhost:4002/v1/module-infos", strings.NewReader(string(moduleInfoJson)))
	s.Nil(err)

	request.Header.Add("Content-Type", "application/json")
	response, err := s.Client.Do(request)
	s.Nil(err)
	s.Equal(403, response.StatusCode)

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	s.Nil(err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	s.Nil(err)

	message, exists := result["error"]
	s.True(exists, "Expected 'error' key in response")
	s.Equal("your user account doesn't have the necessary permissions to access this resource", message)
}

func (s *InfosTestSuite) TestModuleInfoModel_Get() {
	request, err := http.NewRequest(http.MethodGet, "http://localhost:4002/v1/module-infos/6", nil)
	s.Nil(err)
	request.Header.Add("Content-Type", "application/json")

	response, err := s.Client.Do(request)
	s.Nil(err)

	body, err := io.ReadAll(response.Body)
	s.Nil(err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	s.Nil(err)

	message, exists := result["error"]
	s.True(exists, "Expected 'error' key in response")
	s.Equal("you must be authenticated to access this resource", message)
}

func (s *InfosTestSuite) TestModuleInfoModel_Get_Authenticated() {
	token := AuthenticateUserInfo("admin@example.com", "adminpassword")
	s.NotEmpty(token, "Token is empty")

	request, err := http.NewRequest(http.MethodGet, "http://localhost:4002/v1/module-infos/6", nil)
	s.Nil(err)
	request.Header.Add("Authorization", "Bearer "+token)

	response, err := s.Client.Do(request)
	s.Nil(err)

	body, err := io.ReadAll(response.Body)
	s.Nil(err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	s.Nil(err)

	message, exists := result["error"]
	s.False(exists, "'error' message not expected")
	s.Nil(message, "message not expected")

	moduleInfo, moduleInfoExists := result["module_info"]

	miMap, ok := moduleInfo.(map[string]interface{})
	s.True(ok, "Mapping to map from map[string]interface{} failed")
	mi := &data.ModuleInfo{
		ID:         int(miMap["id"].(float64)),
		ModuleName: miMap["moduleName"].(string),
		ExamType:   miMap["examType"].(string),
		Version:    miMap["version"].(string),
	}

	s.True(moduleInfoExists, "Module info expected")
	s.Equal(6, mi.ID, "Id of module are not equal")
	s.Equal("Test1", mi.ModuleName)
}

func (s *InfosTestSuite) TestModuleInfoModel_GetLatestFifty() {
	token := AuthenticateUserInfo("admin@example.com", "adminpassword")
	s.NotEmpty(token, "Token is empty")

	request, err := http.NewRequest(http.MethodGet, "http://localhost:4002/v1/module-infos", nil)
	s.Nil(err)
	request.Header.Add("Authorization", "Bearer "+token)
	response, err := s.Client.Do(request)
	s.Nil(err)
	defer response.Body.Close()
	s.Equal(200, response.StatusCode)

	// {"error":"you must be authenticated to access this resource"}
	body, err := io.ReadAll(response.Body)
	s.Nil(err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	s.Nil(err)
}

func (s *InfosTestSuite) TestModuleInfoModel_GetLatestFifty_Authenticated() {
	request, err := http.NewRequest(http.MethodGet, "http://localhost:4002/v1/module-infos", nil)
	s.Nil(err)
	response, err := s.Client.Do(request)
	s.Nil(err)
	defer response.Body.Close()
	s.Equal(401, response.StatusCode)

	body, err := io.ReadAll(response.Body)
	s.Nil(err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	s.Nil(err)

	message, exists := result["error"]
	s.True(exists, "Expected 'error' key in response")
	s.Equal("you must be authenticated to access this resource", message)
}

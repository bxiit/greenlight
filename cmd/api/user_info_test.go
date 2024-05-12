package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bxiit/greenlight/internal/data"
	"github.com/bxiit/greenlight/internal/jsonlog"
	"github.com/bxiit/greenlight/internal/mailer"
	"github.com/bxiit/greenlight/internal/validator"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func AuthenticateUserInfo(email, password string) string {
	var userData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	userData.Email = email
	userData.Password = password

	userDataJson, err := json.Marshal(userData)
	if err != nil {
		fmt.Println("Error marshalling user data:", err)
		return ""
	}

	request, err := http.NewRequest(http.MethodPost,
		"http://localhost:4002/v1/tokens/authentication",
		strings.NewReader(string(userDataJson)))
	if err != nil {
		return ""
	}

	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Minute,
	}

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return ""
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return ""
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Error unmarshalling response body: %v\n", err)
		return ""
	}

	authToken, exists := result["authentication_token"]
	if !exists {
		fmt.Println("Authentication authToken not found in the response")
		return ""
	}

	var authTokenMap, ok = authToken.(map[string]interface{})
	if !ok {
		fmt.Println("Authentication token is not a map")
		return ""
	}

	token, tokenExists := authTokenMap["token"]
	if !tokenExists {
		fmt.Println("Authentication token is not a map")
		return ""
	}

	tokenValue, ok := token.(string)
	if !ok {
		fmt.Println("Interface value is not a string")
		return ""
	}
	return tokenValue
}

type Suite struct {
	suite.Suite
	Server *http.Server
	Client *http.Client
	App    *application
}

func TestSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}

func (s *Suite) SetupSuite() {
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

func (s *Suite) TestUserInfoModel_GetUserInfo() {
	userInfo, err := s.App.models.UserInfos.Get(12)
	s.Nil(err)

	s.Equal(12, int(userInfo.ID))
}

func (s *Suite) TestUserInfoRepo_GetAllUserInfo() {
	userInfos, err := s.App.models.UserInfos.GetAll()
	s.Nil(err)
	s.Equal(13, len(userInfos))
}

func (s *Suite) TestUserInfo_DeleteUserInfo_InvalidId() {
	err := s.App.models.UserInfos.Delete(-1)
	s.Equal(errors.New("record (row, entry) not found"), err)
}

func (s *Suite) TestUserInfo_DeleteUserInfo_NotExistId() {
	err := s.App.models.UserInfos.Delete(2000)
	s.Equal(errors.New("record (row, entry) not found"), err)
}

func (s *Suite) TestUserInfo_ValidateUserInfo() {
	v := validator.New()
	userInfo := &data.UserInfo{
		Name: "",
	}
	err := userInfo.PasswordHashed.Set("password")
	s.Nil(err)

	data.ValidateUserInfo(v, userInfo)

	s.False(v.Valid())
}

func (s *Suite) TestUserInfo_UpdateUserInfo() {
	ui, err := s.App.models.UserInfos.GetByEmail("test2ass3@example.com")
	s.Nil(err)

	ui.Name = "test updated name"

	err = s.App.models.UserInfos.Update(ui)
	s.Nil(err)

	ui, err = s.App.models.UserInfos.GetByEmail("test2ass3@example.com")
	s.Nil(err)

	s.Equal("test2ass3@example.com", ui.Email)
	s.Equal("test updated name", ui.Name)
}

package appconfig

import (
	"encoding/json"
	"github.com/ksensehq/eventnative/authorization"
	"github.com/ksensehq/eventnative/geo"
	"github.com/ksensehq/eventnative/logging"
	"github.com/ksensehq/eventnative/useragent"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
)

type serverName struct {
	ServerName string `json:"server_name"`
}

type AppConfig struct {
	ServerName string
	Authority  string

	GeoResolver geo.Resolver
	UaResolver  useragent.Resolver

	AuthorizationService *authorization.Service

	closeMe []io.Closer
}

var Instance *AppConfig

const fileDestination = "/resources/server.name"

func setDefaultParams() {
	viper.SetDefault("server.port", "8001")
	viper.SetDefault("server.static_files_dir", "./web")
	viper.SetDefault("server.auth_reload_sec", 30)
	viper.SetDefault("server.destinations_reload_sec", 40)
	viper.SetDefault("geo.maxmind_path", "/home/eventnative/app/res/")
	viper.SetDefault("log.path", "/home/eventnative/logs/events")
	viper.SetDefault("log.show_in_server", false)
	viper.SetDefault("log.rotation_min", "5")
	viper.SetDefault("synchronization_service.connection_timeout_seconds", 20)
}

func Init() error {
	setDefaultParams()

	serverName := viper.GetString("server.name")
	if serverName == "" {
		serverName = readServerName(fileDestination)
		//serverName = "unnamed-server"
	}

	err := logging.InitGlobalLogger(logging.Config{
		LoggerName:  "main",
		ServerName:  serverName,
		FileDir:     viper.GetString("server.log.path"),
		RotationMin: viper.GetInt64("server.log.rotation_min"),
		MaxBackups:  viper.GetInt("server.log.max_backups")})
	if err != nil {
		return err
	}

	logging.Info("*** Creating new AppConfig ***")
	logging.Info("Server Name:", serverName)
	publicUrl := viper.GetString("server.public_url")
	if publicUrl == "" {
		logging.Warn("Server public url: will be taken from Host header")
	} else {
		logging.Info("Server public url:", publicUrl)
	}

	var appConfig AppConfig
	appConfig.ServerName = serverName

	port := viper.GetString("port")
	if port == "" {
		port = viper.GetString("server.port")
	}
	appConfig.Authority = "0.0.0.0:" + port

	geoResolver, err := geo.CreateResolver(viper.GetString("geo.maxmind_path"))
	if err != nil {
		logging.Warn("Run without geo resolver:", err)
	}

	authService, err := authorization.NewService()
	if err != nil {
		return err
	}

	appConfig.AuthorizationService = authService
	appConfig.GeoResolver = geoResolver
	appConfig.UaResolver = useragent.NewResolver()

	Instance = &appConfig
	return nil
}

func (a *AppConfig) ScheduleClosing(c io.Closer) {
	a.closeMe = append(a.closeMe, c)
}

func (a *AppConfig) Close() {
	for _, cl := range a.closeMe {
		if err := cl.Close(); err != nil {
			logging.Error(err)
		}
	}
}

func readServerName(fileStr string) string {

	file, err := os.Open(fileStr)
	if err != nil {
		return createFile(fileStr)
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil || len(byteValue) == 0 {
		serverNameFileStruct := createFileStruct()
		b, _ := json.Marshal(serverNameFileStruct)
		file.Write(b)
	}

	var serverNameFromFile serverName
	json.Unmarshal(byteValue, &serverNameFromFile)

	return serverNameFromFile.ServerName
}

func createFile(fileStr string) string {
	serverNameFile := createFileStruct()

	file, _ := json.MarshalIndent(serverNameFile, "", " ")

	_ = ioutil.WriteFile(fileStr, file, 0644)

	return serverNameFile.ServerName
}

func createFileStruct() serverName {
	return serverName{
		ServerName: "unnamed-server",
	}
}

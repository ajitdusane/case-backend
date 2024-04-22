package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	"github.com/case-framework/case-backend/pkg/utils"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"gopkg.in/yaml.v2"
)

// Environment variables
const (
	ENV_CONFIG_FILE_PATH = "CONFIG_FILE_PATH"
)

type config struct {
	// Logging configs
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// DB configs
	DBConfigs struct {
		ParticipantUserDB db.DBConfigYaml `json:"participant_user_db" yaml:"participant_user_db"`
		GlobalInfosDB     db.DBConfigYaml `json:"global_infos_db" yaml:"global_infos_db"`
		MessagingDB       db.DBConfigYaml `json:"messaging_db" yaml:"messaging_db"`
	} `json:"db_configs" yaml:"db_configs"`

	InstanceIDs []string `json:"instance_ids" yaml:"instance_ids"`

	// user management configs
	UserManagementConfig struct {
		DeleteUnverifiedUsersAfter        time.Duration `json:"delete_unverified_users_after" yaml:"delete_unverified_users_after"`
		SendReminderToConfirmAccountAfter time.Duration `json:"send_reminder_to_confirm_account_after" yaml:"send_reminder_to_confirm_account_after"`
		EmailContactVerificationTokenTTL  time.Duration `json:"email_contact_verification_token_ttl" yaml:"email_contact_verification_token_ttl"`
	} `json:"user_management_config" yaml:"user_management_config"`

	MessagingConfigs messagingTypes.MessagingConfigs `json:"messaging_configs" yaml:"messaging_configs"`
}

var conf config

var (
	participantUserDBService *userDB.ParticipantUserDBService
	globalInfosDBService     *globalinfosDB.GlobalInfosDBService
	messagingDBService       *messagingDB.MessagingDBService
)

func init() {
	// Read config from file
	yamlFile, err := os.ReadFile(os.Getenv(ENV_CONFIG_FILE_PATH))
	if err != nil {
		panic(err)
	}

	err = yaml.UnmarshalStrict(yamlFile, &conf)
	if err != nil {
		panic(err)
	}

	// Init logger:
	utils.InitLogger(
		conf.Logging.LogLevel,
		conf.Logging.IncludeSrc,
		conf.Logging.LogToFile,
		conf.Logging.Filename,
		conf.Logging.MaxSize,
		conf.Logging.MaxAge,
		conf.Logging.MaxBackups,
	)

	// check config values:
	if conf.UserManagementConfig.DeleteUnverifiedUsersAfter == 0 {
		slog.Error("DeleteUnverifiedUsersAfter is not set")
		panic("DeleteUnverifiedUsersAfter is not set")
	}

	if conf.UserManagementConfig.SendReminderToConfirmAccountAfter == 0 {
		slog.Error("SendReminderToConfirmAccountAfter is not set")
		panic("SendReminderToConfirmAccountAfter is not set")
	}

	// init db
	participantUserDBService, err = userDB.NewParticipantUserDBService(db.DBConfigFromYamlObj(conf.DBConfigs.ParticipantUserDB, conf.InstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Participant User DB", slog.String("error", err.Error()))
		panic(err)
	}

	globalInfosDBService, err = globalinfosDB.NewGlobalInfosDBService(db.DBConfigFromYamlObj(conf.DBConfigs.GlobalInfosDB, conf.InstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Global Infos DB", slog.String("error", err.Error()))
		panic(err)
	}

	messagingDBService, err = messagingDB.NewMessagingDBService(db.DBConfigFromYamlObj(conf.DBConfigs.MessagingDB, conf.InstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Messaging DB", slog.String("error", err.Error()))
		panic(err)
	}

	// init message sending
	initMessageSendingConfig()
}

func initMessageSendingConfig() {
	emailsending.InitMessageSendingVariables(
		loadEmailClientHTTPConfig(),
		conf.MessagingConfigs.GlobalEmailTemplateConstants,
	)
}

func loadEmailClientHTTPConfig() httpclient.ClientConfig {
	return httpclient.ClientConfig{
		RootURL: conf.MessagingConfigs.SmtpBridgeConfig.URL,
		APIKey:  conf.MessagingConfigs.SmtpBridgeConfig.APIKey,
		Timeout: conf.MessagingConfigs.SmtpBridgeConfig.RequestTimeout,
	}
}

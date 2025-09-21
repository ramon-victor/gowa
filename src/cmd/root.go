package cmd

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"os"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	domainChat "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chat"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	domainGroup "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/group"
	domainMessage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/message"
	domainNewsletter "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/newsletter"
	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	domainUser "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/user"
	domainWebhook "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/webhook"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mau.fi/whatsmeow"
)

var (
	EmbedIndex embed.FS
	EmbedViews embed.FS

	// Whatsapp
	whatsappCli *whatsmeow.Client

	// Chat Storage
	chatStorageDB   *sql.DB
	chatStorageRepo domainChatStorage.IChatStorageRepository

	// Usecase
	appUsecase        domainApp.IAppUsecase
	chatUsecase       domainChat.IChatUsecase
	sendUsecase       domainSend.ISendUsecase
	userUsecase       domainUser.IUserUsecase
	messageUsecase    domainMessage.IMessageUsecase
	groupUsecase      domainGroup.IGroupUsecase
	newsletterUsecase domainNewsletter.INewsletterUsecase
	webhookUsecase    domainWebhook.IWebhookUsecase
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "Send free whatsapp API",
	Long: `This application is from clone https://github.com/aldinokemal/go-whatsapp-web-multidevice, 
you can send whatsapp over http api but your whatsapp account have to be multi device version`,
}

func init() {
	// Load environment variables first
	utils.LoadConfig(".")

	time.Local = time.UTC

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Initialize flags first, before any subcommands are added
	initCommandFlags()

	// Then initialize other components
	cobra.OnInitialize(initConfigFromEnv, initApp)
}

// initEnvConfig loads configuration from environment variables
func initConfigFromEnv() {
	fmt.Println(viper.AllSettings())
	// Application settings
	if envPort := viper.GetString("app_port"); envPort != "" {
		config.AppPort = envPort
	}
	if envDebug := viper.GetBool("app_debug"); envDebug {
		config.AppDebug = envDebug
	}
	if envOs := viper.GetString("app_os"); envOs != "" {
		config.AppOs = envOs
	}
	if envBasicAuth := viper.GetString("app_basic_auth"); envBasicAuth != "" {
		credential := strings.Split(envBasicAuth, ",")
		config.AppBasicAuthCredential = credential
	}
	if envBasePath := viper.GetString("app_base_path"); envBasePath != "" {
		config.AppBasePath = envBasePath
	}

	// Database settings
	if envDBURI := viper.GetString("db_uri"); envDBURI != "" {
		config.DBURI = envDBURI
	}
	if envDBKEYSURI := viper.GetString("db_keys_uri"); envDBKEYSURI != "" {
		config.DBKeysURI = envDBKEYSURI
	}

	// WhatsApp settings
	if envAutoReply := viper.GetString("whatsapp_auto_reply"); envAutoReply != "" {
		config.WhatsappAutoReplyMessage = envAutoReply
	}
	if viper.IsSet("whatsapp_auto_mark_read") {
		config.WhatsappAutoMarkRead = viper.GetBool("whatsapp_auto_mark_read")
	}
	if viper.IsSet("whatsapp_account_validation") {
		config.WhatsappAccountValidation = viper.GetBool("whatsapp_account_validation")
	}
}

func initCommandFlags() {
	// Application flags
	rootCmd.PersistentFlags().StringVarP(
		&config.AppPort,
		"port", "p",
		config.AppPort,
		"change port number with --port <number> | example: --port=8080",
	)

	rootCmd.PersistentFlags().BoolVarP(
		&config.AppDebug,
		"debug", "d",
		config.AppDebug,
		"hide or displaying log with --debug <true/false> | example: --debug=true",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppOs,
		"os", "",
		config.AppOs,
		`os name --os <string> | example: --os="Chrome"`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.AppBasicAuthCredential,
		"basic-auth", "b",
		config.AppBasicAuthCredential,
		"basic auth credential | -b=yourUsername:yourPassword",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppBasePath,
		"base-path", "",
		config.AppBasePath,
		`base path for subpath deployment --base-path <string> | example: --base-path="/gowa"`,
	)

	// Database flags
	rootCmd.PersistentFlags().StringVarP(
		&config.DBURI,
		"db-uri", "",
		config.DBURI,
		`the database uri to store the connection data database uri (by default, we'll use sqlite3 under storages/whatsapp.db). database uri --db-uri <string> | example: --db-uri="file:storages/whatsapp.db?_foreign_keys=on or postgres://user:password@localhost:5432/whatsapp"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.DBKeysURI,
		"db-keys-uri", "",
		config.DBKeysURI,
		`the database uri to store the keys database uri (by default, we'll use the same database uri). database uri --db-keys-uri <string> | example: --db-keys-uri="file::memory:?cache=shared&_foreign_keys=on"`,
	)

	// WhatsApp flags
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappAutoReplyMessage,
		"autoreply", "",
		config.WhatsappAutoReplyMessage,
		`auto reply when received message --autoreply <string> | example: --autoreply="Don't reply this message"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAutoMarkRead,
		"auto-mark-read", "",
		config.WhatsappAutoMarkRead,
		`auto mark incoming messages as read --auto-mark-read <true/false> | example: --auto-mark-read=true`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAccountValidation,
		"account-validation", "",
		config.WhatsappAccountValidation,
		`enable or disable account validation --account-validation <true/false> | example: --account-validation=true`,
	)
}

func initChatStorageDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("%s?_journal_mode=WAL", config.ChatStorageURI)
	if config.ChatStorageEnableForeignKeys {
		connStr += "&_foreign_keys=on"
	}

	logrus.Debugf("Opening chat storage database with connection string: %s", connStr)
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open chat storage database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	logrus.Debugf("Chat storage connection pool configured: maxOpen=%d, maxIdle=%d", 25, 5)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping chat storage database: %w", err)
	}

	logrus.Debugln("Chat storage database connection established successfully")
	return db, nil
}

func initWebhookDB() (*sql.DB, bool, error) {
	isPostgres := strings.Contains(config.DBURI, "postgres:")
	var db *sql.DB
	var err error

	logrus.Debugf("Initializing webhook database connection (PostgreSQL: %v)", isPostgres)

	if isPostgres {
		logrus.Debugf("Opening PostgreSQL connection with URI: %s", config.DBURI)
		db, err = sql.Open("postgres", config.DBURI)
	} else {
		connStr := config.DBURI
		if config.WebhookEnableForeignKeys {
			connStr += "&_foreign_keys=on"
			logrus.Debug("Foreign keys enabled for webhook database")
		}
		if config.WebhookEnableWAL {
			connStr += "&_journal_mode=WAL"
			logrus.Debug("WAL journal mode enabled for webhook database")
		}
		logrus.Debugf("Opening SQLite connection with string: %s", connStr)
		db, err = sql.Open("sqlite3", connStr)
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to create webhook database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	logrus.Debugf("Webhook database connection pool configured: maxOpen=%d, maxIdle=%d", 25, 5)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, false, fmt.Errorf("failed to ping webhook database: %w", err)
	}

	logrus.Debugln("Webhook database connection established successfully")
	return db, isPostgres, nil
}

// setupLogging configures logging based on debug mode
func initLogging() {
	if config.AppDebug {
		config.WhatsappLogLevel = "DEBUG"
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugln("Debug logging enabled")
	} else {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Infoln("Info level logging enabled")
	}
}

// createRequiredFolders creates all necessary folders for the application
func initDirectories() error {
	folders := []string{config.PathQrCode, config.PathSendItems, config.PathStorages, config.PathMedia}
	logrus.Debugf("Creating required folders: %v", folders)
	
	err := utils.CreateFolder(folders...)
	if err != nil {
		return fmt.Errorf("failed to create required folders: %w", err)
	}
	
	logrus.Infof("Successfully created required folders: %v", folders)
	return nil
}

// initializeChatStorage initializes the chat storage database and repository
func initChatStorage() error {
	var err error
	logrus.Info("Initializing chat storage...")
	
	chatStorageDB, err = initChatStorageDB()
	if err != nil {
		return fmt.Errorf("failed to initialize chat storage: %w", err)
	}

	chatStorageRepo = chatstorage.NewStorageRepository(chatStorageDB)
	err = chatStorageRepo.InitializeSchema()
	if err != nil {
		return fmt.Errorf("failed to initialize chat storage schema: %w", err)
	}

	logrus.Info("Chat storage initialized successfully")
	return nil
}

// initializeWhatsAppDatabases initializes WhatsApp databases (main and keys)
func initWhatsAppDBs(ctx context.Context) (*sqlstore.Container, *sqlstore.Container, error) {
	logrus.Info("Initializing WhatsApp databases...")
	
	whatsappDB := whatsapp.InitWaDB(ctx, config.DBURI)
	logrus.Debugf("Main WhatsApp database initialized with URI: %s", config.DBURI)
	
	var keysDB *sqlstore.Container
	if config.DBKeysURI != "" {
		keysDB = whatsapp.InitWaDB(ctx, config.DBKeysURI)
		logrus.Infof("WhatsApp keys database initialized with URI: %s", config.DBKeysURI)
	} else {
		logrus.Info("Using main database for WhatsApp keys (no separate keys URI specified)")
	}

	logrus.Info("WhatsApp databases initialized successfully")
	return whatsappDB, keysDB, nil
}

// initializeWhatsAppClient initializes the WhatsApp client
func initWhatsAppClient(ctx context.Context, whatsappDB, keysDB *sqlstore.Container) error {
	logrus.Info("Initializing WhatsApp client...")
	
	whatsapp.InitWaCLI(ctx, whatsappDB, keysDB, chatStorageRepo)
	logrus.Info("WhatsApp client initialized successfully")
	return nil
}

// initializeUsecases initializes all application usecases
func initUsecases() {
	logrus.Info("Initializing application usecases...")
	
	appUsecase = usecase.NewAppService(chatStorageRepo)
	chatUsecase = usecase.NewChatService(chatStorageRepo)
	sendUsecase = usecase.NewSendService(appUsecase, chatStorageRepo)
	userUsecase = usecase.NewUserService()
	messageUsecase = usecase.NewMessageService(chatStorageRepo)
	groupUsecase = usecase.NewGroupService()
	newsletterUsecase = usecase.NewNewsletterService()
	
	logrus.Info("All usecases initialized successfully")
	logrus.Debugf("Usecases initialized: app, chat, send, user, message, group, newsletter")
}

// initializeWebhookSystem initializes the webhook database, repository, and service
func initWebhook() error {
	logrus.Info("Initializing webhook system...")
	
	webhookDB, isPostgres, err := initWebhookDB()
	if err != nil {
		return fmt.Errorf("failed to initialize webhook database: %w", err)
	}

	webhookRepo := webhook.NewRepository(webhookDB, isPostgres)
	webhookUsecase = usecase.NewWebhookService(webhookRepo)

	err = webhookRepo.InitializeSchema()
	if err != nil {
		return fmt.Errorf("failed to initialize webhook schema: %w", err)
	}

	whatsapp.InitWebhookService(webhookRepo)
	logrus.Info("Webhook system initialized successfully")
	logrus.Debugf("Webhook database type: %s", map[bool]string{true: "PostgreSQL", false: "SQLite"}[isPostgres])
	return nil
}

func initApp() {
	logrus.Info("Starting application initialization...")

	ctx := context.Background()
	
	initLogging()

	if err := initDirectories(); err != nil {
		logrus.Warnf("Folder creation warning: %v (application may continue)", err)
	}

	if err := initChatStorage(); err != nil {
		logrus.Fatalf("Fatal error: failed to initialize chat storage: %v", err)
	}

	whatsappDB, keysDB, err := initWhatsAppDBs(ctx)
	if err != nil {
		logrus.Fatalf("Fatal error: failed to initialize WhatsApp databases: %v", err)
	}

	if err := initWhatsAppClient(ctx, whatsappDB, keysDB); err != nil {
		logrus.Fatalf("Fatal error: failed to initialize WhatsApp client: %v", err)
	}

	if err := initWebhook(); err != nil {
		logrus.Fatalf("Fatal error: failed to initialize webhook system: %v", err)
	}

	initUsecases()

	logrus.Infoln("✅ Application initialized successfully")
	logrus.Infof("📊 Configuration: Debug=%v, Port=%s, BasePath=%s",
		config.AppDebug, config.AppPort, config.AppBasePath)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(embedIndex embed.FS, embedViews embed.FS) {
	EmbedIndex = embedIndex
	EmbedViews = embedViews
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

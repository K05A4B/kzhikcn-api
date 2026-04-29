package data

import (
	"database/sql"
	_ "kzhikcn/pkg/data/cache"
	"kzhikcn/pkg/log"
	"kzhikcn/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	driver string

	ErrNoConnectionToDatabase = errors.New("no connection to the database")

	onConnectedHooks = utils.SyncSlice[func()]{}
)

func OnConnected(fn func()) {
	onConnectedHooks.Append(fn)
}

func ConnectDatabase(driverName, dsn string) (err error) {
	if db != nil {
		return nil
	}

	var dri gorm.Dialector
	switch driverName {
	case "sqlite3", "sqlite":
		dri, err = connectSQLite(dsn)

	case "mysql":
		dri = mysql.Open(dsn)
	default:
		return errors.Errorf("unsupported driver `%s`", driverName)
	}

	if err != nil {
		return errors.Errorf("failed to connect database. %e", err)
	}

	driver = driverName
	db, err = gorm.Open(dri)
	if err != nil {
		return errors.Errorf("failed to connect database. %e", err)
	}

	if utils.IsDevelopment() {
		db.Logger = &gormLogger{logger: *log.GetLogger()}
	} else {
		db.Logger = (&gormLogger{logger: *log.GetLogger()}).LogMode(logger.Warn)
	}

	for _, fn := range onConnectedHooks.Raw() {
		fn()
	}

	return
}

func connectSQLite(dsn string) (driver gorm.Dialector, err error) {
	dri := sqlite.Open(dsn)
	path, _ := parseSQLiteDSN(dsn)
	_, err = os.Stat(path)
	if !strings.HasPrefix(path, ":memory:") && os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(path), 0640)

		f, err := os.OpenFile(path, os.O_CREATE, 0700)
		if err != nil {
			return nil, err
		}

		f.Close()
	}
	return dri, nil
}

func parseSQLiteDSN(dsn string) (path string, params map[string]string) {
	params = make(map[string]string)

	parts := strings.SplitN(dsn, "?", 2)

	// path
	path = strings.TrimPrefix(parts[0], "file:")

	// query
	if len(parts) == 2 {
		for _, kv := range strings.Split(parts[1], "&") {
			pair := strings.SplitN(kv, "=", 2)
			if len(pair) == 2 {
				params[pair[0]] = pair[1]
			}
		}
	}

	return
}

func AutoMigrates() error {
	state, err := GetSchemaState()
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if state != nil && state.Version == schemaVersion {
		return nil
	}

	err = db.AutoMigrate(
		&Admin{},
		&Article{},
		&Category{},
		&Tag{},
		&ArticleTag{},
		&FriendLink{},
	)
	if err != nil {
		return err
	}

	return db.Model(&SchemaState{}).Create(&SchemaState{Version: schemaVersion, LastMigration: time.Now()}).Error
}

func InitDatabase() (*Admin, error) {
	err := AutoMigrates()
	if err != nil {
		return nil, err
	}

	admin := &Admin{
		Password: []byte("admin"),
		Username: "admin",
	}

	err = AddAdmin(admin)

	admin.Password = []byte("admin")
	return admin, err
}

func DB() *gorm.DB {
	return db
}

func GetDriverName() string {
	return driver
}

func Model(a any) (tx *gorm.DB) {
	return db.Model(a)
}

func Create(a any) (tx *gorm.DB) {
	return db.Create(a)
}

func Delete(a any, cond ...any) (tx *gorm.DB) {
	return db.Delete(a, cond...)
}

func FirstOrCreate(dest any, cond ...any) (tx *gorm.DB) {
	return db.FirstOrCreate(dest, cond...)
}

func Save(a any) (tx *gorm.DB) {
	return db.Save(a)
}

func Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return db.Transaction(fc, opts...)
}

func Preload(query string, args ...any) (tx *gorm.DB) {
	return db.Preload(query, args...)
}

func Where(query any, args ...any) (tx *gorm.DB) {
	return db.Where(query, args...)
}

func First(dest interface{}, cond ...any) (tx *gorm.DB) {
	return db.First(dest, cond...)
}

package config

import (
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
	"gopkg.in/yaml.v3"
)

var (
	config    *Config
	configMux = sync.RWMutex{}

	loadedHooks = []func(*Config){}
)

type LogConf struct {
	Enable     bool               `yaml:"enable"`
	Lumberjack *lumberjack.Logger `yaml:"lumberjack"`
	LogLevel   string             `yaml:"log_level"`
}

type Database struct {
	Driver string         `yaml:"driver"`
	Dsn    EnvVarResolver `yaml:"dsn"`
}

type HttpRate struct {
	LimitPerIP    RateLimit        `yaml:"limit_per_ip"`
	BlackList     []NetAddr        `yaml:"black_list"`
	HighQuotaKeys []EnvVarResolver `yaml:"high_quota_keys"`
}

type Auth struct {
	TOTPMasterKey EnvVarResolver `yaml:"totp_master_key"`
	JWT           struct {
		Secret EnvVarResolver `yaml:"secret"`
		Expiry Duration       `yaml:"expiry"`
	} `yaml:"jwt"`
}

type SitemapExtend struct {
	Location        string `yaml:"loc"`
	LastModify      string `yaml:"last_mod"`
	ChangeFrequency string `yaml:"change_freq"`
	Priority        int8   `yaml:"priority"`
}

type MachineReadableResources struct {
	BaseUrl      EnvVarResolver `yaml:"base_url"`
	UrlTemplates struct {
		Article  TemplateString `yaml:"article"`
		Category TemplateString `yaml:"category"`
		Tag      TemplateString `yaml:"tag"`
		Friends  TemplateString `yaml:"friends"`
	} `yaml:"url_templates"`

	Rss struct {
		Enable      bool   `yaml:"enable"`
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
		MaxArticles int    `yaml:"max_articles"`
	} `yaml:"rss"`

	Sitemap struct {
		Enable              bool            `yaml:"enable"`
		Extends             []SitemapExtend `yaml:"extends"`
		TagMinArticles      int             `yaml:"tag_min_articles"`
		CategoryMinArticles int             `yaml:"category_min_articles"`
	} `yaml:"sitemap"`
}

type StorageConf struct {
	Provider string `yaml:"provider"`
	Articles struct {
		BasePath string `yaml:"base_path"`
	} `yaml:"articles"`
}

type CacheRedisConf struct {
	Addr     string         `yaml:"addr"`
	Prefix   string         `yaml:"prefix"`
	Username string         `yaml:"username"`
	Password EnvVarResolver `yaml:"password"`
	DB       int            `yaml:"db"`
	Timeout  Duration       `yaml:"timeout"`
}

type CacheLocalConf struct {
	Dir string `yaml:"dir"`
}

type CacheConf struct {
	Provider string         `yaml:"provider"`
	Local    CacheLocalConf `yaml:"local"`
	Redis    CacheRedisConf `yaml:"redis"`
}

type Config struct {
	CertFile string      `yaml:"cert_file"`
	KeyFile  string      `yaml:"key_file"`
	Storage  StorageConf `yaml:"storage"`

	MRR MachineReadableResources `yaml:"machine_readable_resources"`

	HttpRate HttpRate  `yaml:"http_rate"`
	Auth     Auth      `yaml:"auth"`
	Database Database  `yaml:"db"`
	Log      LogConf   `yaml:"log"`
	Cache    CacheConf `yaml:"cache"`
}

func LoadConfig(file string) (*Config, error) {
	configMux.Lock()
	defer configMux.Unlock()

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	config = new(Config)
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	for _, fn := range loadedHooks {
		fn(config)
	}

	return config, nil
}

func GetConf() *Config {
	configMux.RLock()
	defer configMux.RUnlock()
	copied := *config
	return &copied
}

func Conf() *Config {
	return GetConf()
}

func HookLoaded(fns ...func(*Config)) {
	loadedHooks = append(loadedHooks, fns...)
}

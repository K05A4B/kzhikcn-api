package config

import (
	"io"
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
	Driver string `yaml:"driver"`
	Dsn    string `yaml:"dsn"`
}

type HttpRate struct {
	LimitPerIP    RateLimit `yaml:"limit_per_ip"`
	BlackList     []NetAddr `yaml:"black_list"`
	HighQuotaKeys []string  `yaml:"high_quota_keys"`
}

type Auth struct {
	JWT struct {
		Secret string   `yaml:"secret"`
		Expiry Duration `yaml:"expiry"`
	} `yaml:"jwt"`
}

type SitemapExtend struct {
	Location        string `yaml:"loc"`
	LastModify      string `yaml:"last_mod"`
	ChangeFrequency string `yaml:"change_freq"`
	Priority        int8   `yaml:"priority"`
}

type MachineReadableResources struct {
	BaseUrl      string `yaml:"base_url"`
	UrlTemplates struct {
		Article  TemplateString `yaml:"article"`
		Category TemplateString `yaml:"category"`
		Tag      TemplateString `yaml:"tag"`
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
	Addr     string   `yaml:"addr"`
	Prefix   string   `yaml:"prefix"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db"`
	Timeout  Duration `yaml:"timeout"`
}

type CacheLocalConf struct {
	Dir              string `yaml:"dir"`
	ValueLogFileSize Size   `yaml:"value_log_file_size"`
	MemTableSize     Size   `yaml:"mem_table_size"`
}

type CacheConf struct {
	Provider string         `yaml:"provider"`
	Local    CacheLocalConf `yaml:"local"`
	Redis    CacheRedisConf `yaml:"redis"`
}

type CORSConf struct {
	Enable           bool     `yaml:"enable"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           Duration `yaml:"max_age"`
}

type Event struct {
	// 事件名
	Name string `yaml:"name"`
	// 事件触发时机
	On string `yaml:"on"`
	// 事件执行类型 webhook | command
	Type string `yaml:"type"`
	// 事件执行入口
	Entry string `yaml:"entry"`
	// 是否异步执行。nil 表示默认异步，false 表示强制同步。
	// app.* 生命周期事件始终同步，此配置对其无效。
	Async *bool `yaml:"async"`
	// 单次分发超时，覆盖 event_dispatcher.timeout。未配置或非正值时使用分发器默认超时。
	Timeout Duration `yaml:"timeout"`
	// webhook 请求附带的 HTTP 头，可覆盖默认的 Content-Type 与 User-Agent。
	// 仅对 type=webhook 生效。值支持 ${环境变量} 引用。
	Headers map[string]string `yaml:"headers"`
}

// EventDispatcherConf 配置事件分发器的超时与异步执行池。
type EventDispatcherConf struct {
	// 单次事件分发的超时时间
	Timeout Duration `yaml:"timeout"`
	// 异步执行池的 worker 数量（仅启动时生效）
	Workers int `yaml:"workers"`
	// 异步执行池的队列长度（仅启动时生效）
	QueueSize int `yaml:"queue_size"`
}

type Config rawConfig

type rawConfig struct {
	CertFile string      `yaml:"cert_file"`
	KeyFile  string      `yaml:"key_file"`
	Storage  StorageConf `yaml:"storage"`

	MRR MachineReadableResources `yaml:"machine_readable_resources"`

	HttpRate HttpRate  `yaml:"http_rate"`
	Auth     Auth      `yaml:"auth"`
	Database Database  `yaml:"db"`
	Log      LogConf   `yaml:"log"`
	Cache    CacheConf `yaml:"cache"`
	CORS     CORSConf  `yaml:"cors"`

	EventDispatcher EventDispatcherConf `yaml:"event_dispatcher"`
	Events          []Event             `yaml:"events"`
}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	nodes := []*yaml.Node{}

	// 去除顶层的ScalarNode
	for _, item := range value.Content {
		if item.Kind == yaml.ScalarNode {
			continue
		}

		nodes = append(nodes, item)
	}

	err := resolveEnv(value.Content)
	if err != nil {
		return err
	}

	raw := rawConfig{}
	err = value.Decode(&raw)
	if err != nil {
		return err
	}

	*c = Config(raw)

	return nil
}

func LoadConfig(reader io.Reader) (*Config, error) {
	configMux.Lock()
	defer configMux.Unlock()

	config = new(Config)
	dec := yaml.NewDecoder(reader)
	err := dec.Decode(config)
	if err != nil {
		return nil, err
	}

	for _, fn := range loadedHooks {
		fn(config)
	}

	return config, nil
}

func LoadConfigFromFile(file string) (*Config, error) {
	data, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	defer data.Close()
	config, err := LoadConfig(data)
	return config, err
}

// Bind 用 c 的内容替换全局配置，使未持有 App 的包（如 authtoken）与 App 保持一致。
// 复制的是 c 的深拷贝，避免调用方与全局单例共享切片/映射。
func Bind(c *Config) {
	if c == nil {
		return
	}

	cloned := c.Clone()

	configMux.Lock()
	defer configMux.Unlock()

	if config == nil {
		config = new(Config)
	}
	*config = *cloned
}

// Clone 返回配置的深拷贝。
// 仅复制会被读取的切片与映射；Log.Lumberjack 指向同一个日志写入器，属于有意共享。
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}

	cloned := *c

	cloned.HttpRate.BlackList = append([]NetAddr(nil), c.HttpRate.BlackList...)
	cloned.HttpRate.HighQuotaKeys = append([]string(nil), c.HttpRate.HighQuotaKeys...)

	cloned.MRR.Sitemap.Extends = append([]SitemapExtend(nil), c.MRR.Sitemap.Extends...)

	cloned.CORS.AllowedOrigins = append([]string(nil), c.CORS.AllowedOrigins...)
	cloned.CORS.AllowedMethods = append([]string(nil), c.CORS.AllowedMethods...)
	cloned.CORS.AllowedHeaders = append([]string(nil), c.CORS.AllowedHeaders...)
	cloned.CORS.ExposedHeaders = append([]string(nil), c.CORS.ExposedHeaders...)

	if c.Events != nil {
		events := make([]Event, len(c.Events))
		for i, event := range c.Events {
			events[i] = event
			if event.Headers != nil {
				headers := make(map[string]string, len(event.Headers))
				for k, v := range event.Headers {
					headers[k] = v
				}
				events[i].Headers = headers
			}
		}
		cloned.Events = events
	}

	return &cloned
}

func GetConf() *Config {
	configMux.RLock()
	defer configMux.RUnlock()
	return config.Clone()
}

func Conf() *Config {
	return GetConf()
}

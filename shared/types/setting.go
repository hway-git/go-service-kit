package types

var Settings *Setting

type Setting struct {
	App    APP    `json:"app"`    // 应用配置
	Extend Extend `json:"extend"` // 扩展配置
	MySQL  MySQL  `json:"mysql"`  // 数据库配置
	Redis  Redis  `json:"redis"`  // Redis配置
}

// BaseSetting 返回应用嵌入的基础配置，供共享加载入口组装。
func (s *Setting) BaseSetting() *Setting {
	return s
}

// ========

// ========

// ========

// =====================================    APP    ==================================

type APP struct {
	Addr     string `json:"addr"`      // 监听地址
	Port     int    `json:"port"`      // 监听端口
	Env      string `json:"env"`       // 环境
	Debug    bool   `json:"debug"`     // 是否开启debug模式
	Timezone string `json:"timezone"`  // 时区
	NeedCros bool   `json:"need_cros"` // 是否需要设置跨域
	TZ       string `json:"tz"`        // 时区
	VER      string `json:"ver"`       // 静态资源版本号
}

// =====================================    MySQL    ==================================

type MySQL struct {
	// 主库
	MySQLConfig
	// 副本
	Replica []MySQLConfig `json:"replica"`
	// 其他 名称-配置
	Other map[string]MySQLConfig `json:"other"`
}

// MySQLConfig 数据库连接配置（通用）
type MySQLConfig struct {
	Charset  string `json:"charset"`       // 字符集（可选）
	Dsn      string `json:"dsn,omitempty"` // 数据库连接串
	Idle     int    `json:"idle"`          // 连接池中最大空闲连接数
	Open     int    `json:"open"`          // 数据库的最大打开连接数
	Lifetime int64  `json:"lifetime"`      // 连接的最大生命周期（秒）
}

// ==================================        Redis        ======================

type Redis struct {
	// 主配置（单点/cluster 都用这个结构承载）
	RedisConfig
	// 其他 名称-配置（通常是额外的单点/独立实例）
	Other map[string]RedisConfig `json:"other"`
}

// RedisConfig redis连接配置（通用）
const (
	RedisSingle  = "single"
	RedisCluster = "cluster"
)

type RedisConfig struct {
	Mode     string   `json:"mode"`               // 连接模式
	Address  string   `json:"address"`            // 单点地址（host:port）
	Addrs    []string `json:"addrs"`              // 集群节点地址列表（cluster 模式）
	Password string   `json:"password,omitempty"` // 密码
	DB       int      `json:"db"`                 // 数据库
	PoolSize int      `json:"poolsize"`           // 连接池数量
	Tls      bool     `json:"tls"`                // 是否启用TLS
}

// ==================================        Extend        ======================

type Extend struct {
	CommonCfgFile string `json:"common_cfg"` // 通用配置文件路径
}

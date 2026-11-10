package enum

// #####################    语言    ##########################
const (
	LANGUAGE_CN string = "zh"
	LANGUAGE_EN string = "en"
	LANGUAGE_PT string = "pt"
	LANGUAGE_VI string = "vi"
)

// #####################    时区    ##########################
const TIMEZONE_DEFAULT string = TIMEZONE_CHINA // 默认时区
const (
	TIMEZONE_GMT     string = "UTC"               // UTC
	TIMEZONE_CHINA   string = "Asia/Shanghai"     // 中国时区
	TIMEZONE_BRAZIL  string = "America/Sao_Paulo" // 巴西时区
	TIMEZONE_VIETNAM string = "Asia/Ho_Chi_Minh"  // 越南时区
	TIMEZONE_USA     string = "America/New_York"  // 美国时区
)

var TIMEZONE_LIST = []string{
	TIMEZONE_GMT,
	TIMEZONE_CHINA,
	TIMEZONE_BRAZIL,
	TIMEZONE_VIETNAM,
	TIMEZONE_USA,
}

var TIMEZONE_MAP = map[string]string{
	TIMEZONE_GMT:     "UTC",
	TIMEZONE_CHINA:   "中国",
	TIMEZONE_BRAZIL:  "巴西",
	TIMEZONE_VIETNAM: "越南",
	TIMEZONE_USA:     "美国",
}

// ######################   分页   #########################
const (
	DEFAULT_PAGE_SIZE int = 15
)

// ###################     上传文件用途类型   ##################

const (
	UPLOAD_SOURCE_TYPE_AVATAR int = 1
)

var (
	UPLOAD_SOURCE_TYPE_MAP = map[int]string{
		UPLOAD_SOURCE_TYPE_AVATAR: "avatar", //头像
	}
)

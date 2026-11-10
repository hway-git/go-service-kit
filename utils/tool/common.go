package tool

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"example.com/go-service-kit/utils/enum"
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 分页相关
func GetPageInfo(page int, size int) (offset int, pageSize int) {
	if page == 0 {
		page = 1
	}
	pageSize = size
	if pageSize == 0 {
		pageSize = enum.DEFAULT_PAGE_SIZE
	}
	offset = (page - 1) * pageSize
	return
}

// ########################   通用   ######################

// 拼接 Url 参数
func SplicingUrlParams(url string, params string) string {
	if params == "" {
		return url
	}
	// 去除参数左边的 & 或者 ?, 后面视情况而定
	params = strings.TrimLeft(params, "&")
	params = strings.TrimLeft(params, "?")
	if strings.Contains(url, "?") {
		url += "&" + params // 如果已经有参数了，就用 "&" 拼接
	} else {
		url += "?" + params // 如果还没有参数，就用 "?" 开头
	}
	return url
}

// 获取 uuid
func GetUUID() string {
	id := uuid.New()
	noHyphenID := strings.ReplaceAll(id.String(), "-", "")
	return noHyphenID
}

// 生成MD5
func Md5(Str string) string {
	data := []byte(Str)
	md5Hash := md5.Sum(data)
	md5Value := hex.EncodeToString(md5Hash[:])
	return md5Value
}

// 生成一个随机的盐值
func GenerateSalt(length int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes)
}

// hash md5 加密
func HashMd5(hashInput string) string {
	hash := md5.Sum([]byte(hashInput))
	hashValue := hex.EncodeToString(hash[:])
	return hashValue
}

// hash sha256 加密
func HashSha256(hashInput string) string {
	h := sha256.New()
	h.Write([]byte(hashInput))
	hashInBytes := h.Sum(nil)
	hashInString := hex.EncodeToString(hashInBytes)
	return hashInString
}

// 计算 HMAC-SHA256 签名
func HMACSHA256(message, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// 获取客户端真实IP
func GetIP(ctx *gin.Context) string {
	return ctx.ClientIP()
}

// DD 函数用于打印变量而不终止程序
func DD(data ...any) {
	// 使用 JSON 格式化输出，便于阅读复杂结构
	for _, v := range data {
		// 转换为 JSON 格式的字符串
		jsonData, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting data: %v\n", err)
			continue
		}
		fmt.Println(string(jsonData))
	}
}

// 生成count个 [start,end) 结束的不重复的随机数
func GenerateRandomNum(start int, end int, count int) []int {
	nums := make([]int, 0)
	if end < start || (end-start) < count {
		return nums
	}
	for len(nums) < count {
		num := rand.Intn((end - start)) + start
		if InArray(num, nums) {
			continue
		}
		nums = append(nums, num)
	}
	return nums
}

// 邮箱脱敏
func MaskEmail(email string) string {
	if len(email) < 5 {
		return email
	}
	return email[:3] + "***" + email[len(email)-3:]
}

// 银行卡脱敏
func MaskBankCard(cardNo string) string {
	visible := 5
	length := len(cardNo)
	if length <= visible {
		return cardNo // 卡号长度小于等于5，直接返回
	}
	maskedPart := strings.Repeat("*", length-visible)
	visiblePart := cardNo[length-visible:]
	return maskedPart + visiblePart
}

// 数字前面补0
func FillZero[T comparable](num T, length int) string {
	numStr := fmt.Sprintf("%v", num)
	if len(numStr) >= length {
		return numStr
	}
	return strings.Repeat("0", length-len(numStr)) + numStr
}

func RandomNumber(length int) int32 {
	min := math.Pow(10, float64(length-1))
	max := math.Pow(10, float64(length)) - 1
	num := rand.Intn(int(max-min)) + int(min)
	return int32(num)
}

// 账号脱敏
func MaskAccount(account string) string {
	runes := []rune(account)
	n := len(runes)
	if n <= 2 {
		return account
	}
	if n <= 6 {
		return string(runes[:1]) + "***" + string(runes[n-1:])
	}
	return string(runes[:2]) + "***" + string(runes[n-2:])
}

// 去掉回放地址的 http:// 或 https://
func UrlTrimHttpPrefix(url string) string {
	if strings.HasPrefix(url, "https://") {
		return strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		return strings.TrimPrefix(url, "http://")
	}
	return url
}

// 生成 [min, max) 范围内的随机 int
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

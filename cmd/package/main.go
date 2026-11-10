// 打包工具检查并构建两个后端，使用白名单生成 Linux 增量包。
package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"debug/buildinfo"
	"debug/elf"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var buildFlags = []string{"-trimpath", "-buildvcs=false", "-tags=timetzdata", "-ldflags=-s -w"}

type binaryInfo struct {
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
	Static bool   `json:"static"`
}

type releaseInfo struct {
	ReleaseID        string                `json:"release_id"`
	Target           string                `json:"target"`
	BuiltAt          string                `json:"built_at"`
	GoVersion        string                `json:"go_version"`
	SourceCommit     string                `json:"source_commit,omitempty"`
	WorktreeModified *bool                 `json:"worktree_modified,omitempty"`
	BuildFlags       []string              `json:"build_flags"`
	IncludesConfig   bool                  `json:"includes_configuration"`
	Programs         map[string]binaryInfo `json:"programs"`
}

func main() {
	arch := flag.String("arch", "amd64", "Linux 目标架构：amd64 或 arm64")
	goCommand := flag.String("go", "go", "Go 可执行文件路径")
	resultFile := flag.String("result-file", "", "可选的打包结果 JSON 路径")
	flag.Parse()
	if err := packageRelease(*goCommand, *arch, *resultFile); err != nil {
		fmt.Fprintln(os.Stderr, "打包失败：", err)
		os.Exit(1)
	}
}

func packageRelease(goCommand, arch, resultFile string) error {
	if flag.NArg() != 0 || (arch != "amd64" && arch != "arm64") {
		return fmt.Errorf("只支持 amd64、arm64，且不接受位置参数")
	}
	root, err := projectRoot()
	if err != nil {
		return err
	}
	goPath, err := exec.LookPath(goCommand)
	if err != nil {
		return fmt.Errorf("找不到 Go 可执行文件：%w", err)
	}
	goPath, err = filepath.Abs(goPath)
	if err != nil {
		return err
	}
	target := commandEnvironment(goPath, "linux", arch)
	fmt.Println("静态检查 linux/" + arch)
	if err := runGo(root, goPath, target, "vet", "-tags=timetzdata", "./..."); err != nil {
		return fmt.Errorf("静态检查未完成：%w", err)
	}
	output := filepath.Join(root, "dist")
	if err := os.MkdirAll(output, 0o755); err != nil {
		return err
	}
	stamp := time.Now().UTC()
	releaseDir, err := os.MkdirTemp(output, "release-"+stamp.Format("20060102-150405")+"-")
	if err != nil {
		return err
	}
	releaseID := strings.TrimPrefix(filepath.Base(releaseDir), "release-")
	payload := releaseDir
	if err := os.MkdirAll(filepath.Join(payload, "bin"), 0o755); err != nil {
		return err
	}
	programs := make(map[string]binaryInfo)
	for _, name := range []string{"api", "manage"} {
		fmt.Println("构建 linux/" + arch + " " + name)
		path := filepath.Join(payload, "bin", name)
		args := append([]string{"build"}, buildFlags...)
		args = append(args, "-o", path, "./cmd/"+name)
		if err := runGo(root, goPath, target, args...); err != nil {
			return fmt.Errorf("构建 %s 失败：%w", name, err)
		}
		info, err := validateBinary(path, arch)
		if err != nil {
			return fmt.Errorf("校验 %s 失败：%w", name, err)
		}
		programs[name] = info
	}
	version, err := commandOutput(root, goPath, "version")
	if err != nil {
		return err
	}
	commit, modified := sourceInfo(root)
	manifest := &releaseInfo{ReleaseID: releaseID, Target: "linux/" + arch, BuiltAt: stamp.Format(time.RFC3339),
		GoVersion: version, SourceCommit: commit, WorktreeModified: modified, BuildFlags: buildFlags, Programs: programs}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	archiveName := "go-service-kit-linux-" + arch + "-" + releaseID + ".tar.gz"
	if err := os.WriteFile(filepath.Join(payload, "RELEASE.json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	files := []string{"bin/api", "bin/manage"}
	archive := filepath.Join(releaseDir, archiveName)
	if err := createArchive(archive, payload, files, stamp); err != nil {
		return err
	}
	info, err := fileDigest(archive)
	if err != nil {
		return err
	}
	var sums strings.Builder
	for _, name := range append(files, "RELEASE.json") {
		fileInfo, err := fileDigest(filepath.Join(releaseDir, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		fmt.Fprintf(&sums, "%s  %s\n", fileInfo.SHA256, name)
	}
	fmt.Fprintf(&sums, "%s  %s\n", info.SHA256, archiveName)
	if err := os.WriteFile(filepath.Join(releaseDir, "SHA256SUMS"), []byte(sums.String()), 0o644); err != nil {
		return err
	}
	if resultFile != "" {
		result, err := json.Marshal(map[string]string{"archive": archive, "sha256": info.SHA256})
		if err != nil {
			return err
		}
		if err := os.WriteFile(resultFile, append(result, '\n'), 0o600); err != nil {
			return err
		}
	}
	fmt.Printf("\n打包完成：%s\n文件大小：%.1f MiB\nSHA256：%s\n", archive, float64(info.Bytes)/(1024*1024), info.SHA256)
	return nil
}

// projectRoot 通过 Go 模块和两个入口定位脚手架根目录。
func projectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			for _, entry := range []string{"cmd/api/main.go", "cmd/manage/main.go", "cmd/package/main.go"} {
				if info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(entry))); err != nil || !info.Mode().IsRegular() {
					return "", fmt.Errorf("当前 Go 模块缺少脚手架入口：%s", entry)
				}
			}
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("请从项目目录执行打包工具")
		}
		dir = parent
	}
}

// commandEnvironment 为交叉编译设置平台，不加载业务环境文件。
func commandEnvironment(goPath, goos, arch string) []string {
	values := map[string]string{"GOOS": goos, "GOARCH": arch, "CGO_ENABLED": "0"}
	if goos == "linux" && arch == "amd64" {
		values["GOAMD64"] = "v1"
	} else if goos == "linux" && arch == "arm64" {
		values["GOARM64"] = "v8.0"
	}
	var result []string
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if _, exists := values[strings.ToUpper(name)]; !exists && !strings.EqualFold(name, "PATH") {
			result = append(result, entry)
		}
	}
	for name, value := range values {
		result = append(result, name+"="+value)
	}
	return append(result, "PATH="+filepath.Dir(goPath)+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func runGo(root, goPath string, environment []string, args ...string) error {
	cmd := exec.Command(goPath, args...)
	cmd.Dir, cmd.Env = root, environment
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func commandOutput(root, program string, args ...string) (string, error) {
	cmd := exec.Command(program, args...)
	cmd.Dir = root
	raw, err := cmd.Output()
	return strings.TrimSpace(string(raw)), err
}

func sourceInfo(root string) (string, *bool) {
	gitRoot, err := commandOutput(root, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", nil
	}
	if relative, err := filepath.Rel(root, filepath.FromSlash(gitRoot)); err != nil || relative != "." {
		return "", nil
	}
	commit, err := commandOutput(root, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", nil
	}
	status, err := commandOutput(root, "git", "status", "--porcelain", "--untracked-files=normal")
	if err != nil {
		return commit, nil
	}
	modified := status != ""
	return commit, &modified
}

func validateBinary(path, arch string) (binaryInfo, error) {
	file, err := elf.Open(path)
	if err != nil {
		return binaryInfo{}, err
	}
	defer file.Close()
	expected := elf.EM_X86_64
	if arch == "arm64" {
		expected = elf.EM_AARCH64
	}
	if file.Class != elf.ELFCLASS64 || file.Data != elf.ELFDATA2LSB || file.Machine != expected {
		return binaryInfo{}, fmt.Errorf("ELF 架构不匹配")
	}
	for _, program := range file.Progs {
		if program.Type == elf.PT_INTERP {
			return binaryInfo{}, fmt.Errorf("二进制依赖动态解释器")
		}
	}
	for _, section := range file.Sections {
		if section.Name == ".symtab" || strings.HasPrefix(section.Name, ".debug_") {
			return binaryInfo{}, fmt.Errorf("二进制仍包含调试符号")
		}
	}
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return binaryInfo{}, err
	}
	settings := make(map[string]string)
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	required := map[string]string{"GOOS": "linux", "GOARCH": arch, "CGO_ENABLED": "0", "-trimpath": "true", "-tags": "timetzdata"}
	if arch == "amd64" {
		required["GOAMD64"] = "v1"
	} else {
		required["GOARM64"] = "v8.0"
	}
	for key, value := range required {
		if settings[key] != value {
			return binaryInfo{}, fmt.Errorf("构建参数 %s 不匹配", key)
		}
	}
	result, err := fileDigest(path)
	result.Static = true
	return result, err
}

func fileDigest(path string) (binaryInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return binaryInfo{}, err
	}
	defer file.Close()
	hash := sha256.New()
	bytes, err := io.Copy(hash, file)
	return binaryInfo{Bytes: bytes, SHA256: hex.EncodeToString(hash.Sum(nil))}, err
}

// createArchive 只归档指定文件，完成内容和权限校验后才发布正式压缩包。
func createArchive(path, payload string, files []string, stamp time.Time) error {
	output, err := os.CreateTemp(filepath.Dir(path), ".package-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(output.Name())
	compressed := gzip.NewWriter(output)
	archive := tar.NewWriter(compressed)
	writeErr := writeArchive(archive, payload, files, stamp)
	for _, closeErr := range []error{archive.Close(), compressed.Close(), output.Close()} {
		if writeErr == nil {
			writeErr = closeErr
		}
	}
	if writeErr != nil {
		return writeErr
	}
	if err := verifyArchive(output.Name(), payload, files); err != nil {
		return err
	}
	return os.Rename(output.Name(), path)
}

func writeArchive(archive *tar.Writer, payload string, files []string, stamp time.Time) error {
	if err := archive.WriteHeader(&tar.Header{Name: "bin/", Typeflag: tar.TypeDir, Mode: 0o755, ModTime: stamp, Uname: "root", Gname: "root"}); err != nil {
		return err
	}
	for _, name := range files {
		path := filepath.Join(payload, filepath.FromSlash(name))
		stat, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("归档只接受普通文件：%s", name)
		}
		mode := int64(0o644)
		if strings.HasPrefix(name, "bin/") {
			mode = 0o755
		}
		if err := archive.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: mode, Size: stat.Size(), ModTime: stamp, Uname: "root", Gname: "root"}); err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(archive, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func verifyArchive(path, payload string, files []string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer compressed.Close()
	archive := tar.NewReader(compressed)
	expected := make(map[string]binaryInfo)
	for _, name := range files {
		info, err := fileDigest(filepath.Join(payload, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		expected[name] = info
	}
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag == tar.TypeDir && header.Name == "bin/" {
			continue
		}
		info, ok := expected[header.Name]
		if !ok || header.Typeflag != tar.TypeReg || header.Size != info.Bytes ||
			(strings.HasPrefix(header.Name, "bin/") && header.Mode != 0o755) {
			return fmt.Errorf("压缩包存在非预期文件或权限")
		}
		hash := sha256.New()
		if _, err := io.Copy(hash, archive); err != nil {
			return err
		}
		if hex.EncodeToString(hash.Sum(nil)) != info.SHA256 {
			return fmt.Errorf("压缩包文件摘要不匹配")
		}
		delete(expected, header.Name)
	}
	if len(expected) != 0 {
		return fmt.Errorf("压缩包缺少发布文件")
	}
	_, err = io.Copy(io.Discard, compressed)
	return err
}

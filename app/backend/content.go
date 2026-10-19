package backend

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

/*
内容热更新：课程正文放在 content 频道（GitHub Release 的 content tag，同样镜像到 Gitee），
里面只有 content.json 与 content.zip 两个文件。程序里内置一份课程，磁盘上的内容包按
文件名覆盖同名课程，所以改讲义、改练习不用重新发版。
*/
const (
	contentDirName      = "content"
	updatesDirName      = "updates"
	maxContentFileBytes = 256 << 10
	maxContentPackBytes = 4 << 20
	maxContentFiles     = 32
	contentVersionFile  = "version"
)

var contentFileName = regexp.MustCompile(`^[a-z][a-z0-9-]{1,40}\.json$`)

func (e *Engine) updatesDir() string { return filepath.Join(e.dataDir, updatesDirName) }
func (e *Engine) contentDir() string { return filepath.Join(e.dataDir, contentDirName) }

/*
contentVersion 是当前生效的课程内容版本：应用过内容包就是包里记的版本，
没应用过则退回"内置课程的哈希"。后者让全新安装不会一上来就提示"内容有更新"——
内置内容与已发布内容一致时，两个版本号本来就相等。

算法必须与 app/scripts/make-update-manifest.sh 保持一致：
按文件名排序，逐个写入 文件名 + \0 + 内容 + \0，取 sha256 的前 10 位十六进制。
*/
func (e *Engine) contentVersion() string {
	data, err := os.ReadFile(filepath.Join(e.contentDir(), contentVersionFile))
	if err == nil {
		if version := strings.TrimSpace(string(data)); version != "" {
			return version
		}
	}
	files, err := readEmbeddedCourses()
	if err != nil {
		return ""
	}
	bundled := make(map[string][]byte, len(files))
	for _, file := range files {
		data, err := curriculumFS.ReadFile("courses/" + file.source)
		if err != nil {
			continue
		}
		bundled[file.source] = data
	}
	return contentSignature(bundled)
}

// contentSignature 给一组课程文件算版本号；与发布脚本用同一套算法。
func contentSignature(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	digest := sha256.New()
	for _, name := range names {
		digest.Write([]byte(name))
		digest.Write([]byte{0})
		digest.Write(files[name])
		digest.Write([]byte{0})
	}
	return hex.EncodeToString(digest.Sum(nil))[:10]
}

// applyContent 下载内容包、校验、落盘，然后把课程重新载入内存。
func (e *Engine) applyContent(prefer string) (string, error) {
	var manifest contentInfo
	source, err := fetchManifest(e.updateClient, preferred(contentBases(), prefer), contentManifest, &manifest)
	if err != nil {
		return "", errors.New("暂时拿不到课程内容清单，请稍后重试")
	}
	if manifest.File == "" {
		return "", errors.New("课程内容清单缺少文件名")
	}
	if manifest.Version == "" {
		return "", errors.New("课程内容清单缺少版本号")
	}
	if manifest.MinAppVersion != "" && versionNewer(manifest.MinAppVersion, Version) {
		return "", errors.New("课程内容需要更新的程序版本，请先更新 GoStudy")
	}
	if manifest.Version == e.contentVersion() {
		return manifest.Version, nil
	}
	target := filepath.Join(e.updatesDir(), contentPack)
	e.downloads.set(func(s *downloadState) {
		*s = downloadState{State: "downloading", Total: manifest.Pack.Size, Asset: contentPack, Kind: kindContent}
	})
	err = fetchTo(e.updateClient, source+"/"+manifest.File, target, manifest.Pack, func(received int64) {
		e.downloads.set(func(s *downloadState) { s.Received = received })
	})
	if err != nil {
		e.downloads.set(func(s *downloadState) { s.State = "failed"; s.Error = err.Error() })
		return "", err
	}
	e.downloads.set(func(s *downloadState) {
		s.State = "verifying"
		s.Received = manifest.Pack.Size
		s.Total = manifest.Pack.Size
	})
	if err := e.installContentPack(target, manifest); err != nil {
		e.downloads.set(func(s *downloadState) { s.State = "failed"; s.Error = err.Error() })
		return "", err
	}
	e.downloads.set(func(s *downloadState) { s.State = "ready"; s.Mode = "content"; s.Kind = kindContent })
	return manifest.Version, nil
}

// installContentPack 把内容包解到 content 目录，校验通过才替换旧的一份。
func (e *Engine) installContentPack(packPath string, manifest contentInfo) error {
	files, err := readContentPack(packPath)
	if err != nil {
		return err
	}
	// 先按"内置课程 + 这批文件"组装一遍：坏内容包不能把应用的学习内容弄没了。
	merged, err := mergedCourseFiles(files)
	if err != nil {
		return fmt.Errorf("课程内容校验失败：%w", err)
	}
	if _, _, err := assembleCurriculum(merged); err != nil {
		return fmt.Errorf("课程内容校验失败：%w", err)
	}

	staging := e.contentDir() + ".new"
	if err := os.RemoveAll(staging); err != nil {
		return errors.New("无法写入课程内容目录")
	}
	if err := os.MkdirAll(staging, 0700); err != nil {
		return errors.New("无法写入课程内容目录")
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(staging, name), data, 0600); err != nil {
			_ = os.RemoveAll(staging)
			return errors.New("无法保存课程内容")
		}
	}
	if err := os.WriteFile(filepath.Join(staging, contentVersionFile), []byte(manifest.Version), 0600); err != nil {
		_ = os.RemoveAll(staging)
		return errors.New("无法保存课程内容版本")
	}
	// 替换：先把旧目录挪开，再把新目录放上去，失败时还原。
	dir := e.contentDir()
	backup := dir + ".old"
	_ = os.RemoveAll(backup)
	if _, err := os.Stat(dir); err == nil {
		if err := os.Rename(dir, backup); err != nil {
			_ = os.RemoveAll(staging)
			return errors.New("无法替换课程内容目录")
		}
	}
	if err := os.Rename(staging, dir); err != nil {
		_ = os.Rename(backup, dir)
		_ = os.RemoveAll(staging)
		return errors.New("无法启用新的课程内容")
	}
	_ = os.RemoveAll(backup)
	if err := e.reloadCurriculum(); err != nil {
		return err
	}
	return nil
}

// mergedCourseFiles 把内容包里的文件盖到内置课程上，返回合并后的课程文件列表。
func mergedCourseFiles(files map[string][]byte) ([]courseFile, error) {
	merged, err := readEmbeddedCourses()
	if err != nil {
		return nil, err
	}
	index := make(map[string]int, len(merged))
	for i, file := range merged {
		index[file.source] = i
	}
	for name, data := range files {
		file, err := parseCourseFile(data)
		if err != nil {
			return nil, fmt.Errorf("%s：%w", name, err)
		}
		file.source = name
		if i, ok := index[name]; ok {
			merged[i] = file
			continue
		}
		merged = append(merged, file)
	}
	return merged, nil
}

// readContentPack 读 zip 并做结构校验：只允许根目录下的课程 JSON。
func readContentPack(path string) (map[string][]byte, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, errors.New("课程内容包无法读取")
	}
	defer archive.Close()
	files := map[string][]byte{}
	total := 0
	for _, entry := range archive.File {
		name := entry.Name
		// 只认根目录下形如 grammar.json 的文件：目录、绝对路径、多余文件一律拒绝。
		if entry.FileInfo().IsDir() || strings.ContainsAny(name, `/\`) || !contentFileName.MatchString(name) {
			return nil, fmt.Errorf("课程内容包里有多余的内容：%s", name)
		}
		if entry.UncompressedSize64 > maxContentFileBytes {
			return nil, fmt.Errorf("%s 超过单文件上限", name)
		}
		if len(files) >= maxContentFiles {
			return nil, errors.New("课程内容包文件数量超过上限")
		}
		data, err := readZipEntry(entry)
		if err != nil {
			return nil, err
		}
		total += len(data)
		if total > maxContentPackBytes {
			return nil, errors.New("课程内容包解压后超过上限")
		}
		files[name] = data
	}
	if len(files) == 0 {
		return nil, errors.New("课程内容包是空的")
	}
	return files, nil
}

func readZipEntry(entry *zip.File) ([]byte, error) {
	reader, err := entry.Open()
	if err != nil {
		return nil, fmt.Errorf("%s 无法读取", entry.Name)
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, maxContentFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%s 读取失败", entry.Name)
	}
	if len(data) > maxContentFileBytes {
		return nil, fmt.Errorf("%s 超过单文件上限", entry.Name)
	}
	return data, nil
}

// contentPackFromDir 把课程目录打成一个内容包，供发布脚本与测试使用。
func contentPackFromDir(dir string) ([]byte, map[string][]byte, error) {
	names, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, nil, err
	}
	if len(names) == 0 {
		return nil, nil, fmt.Errorf("目录里没有课程文件：%s", dir)
	}
	files := map[string][]byte{}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, full := range names {
		data, err := os.ReadFile(full)
		if err != nil {
			return nil, nil, err
		}
		name := filepath.Base(full)
		files[name] = data
		entry, err := writer.Create(name)
		if err != nil {
			return nil, nil, err
		}
		if _, err := entry.Write(data); err != nil {
			return nil, nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, nil, err
	}
	return buffer.Bytes(), files, nil
}

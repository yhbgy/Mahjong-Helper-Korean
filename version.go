package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"net/http"
	"time"
)

const versionDev = "dev"

// 설명
// go build -ldflags "-X main.version=$(git describe --abbrev=0 --tags)" -o mahjong-helper
var version = versionDev

func fetchLatestVersionTag() (latestVersionTag string, err error) {
	const apiGetLatestRelease = "https://api.github.com/repos/EndlessCheng/mahjong-helper/releases/latest"
	const timeout = 10 * time.Second

	c := &http.Client{Timeout: timeout}
	resp, err := c.Get(apiGetLatestRelease)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("[fetchLatestVersionTag] 반환 %s", resp.Status)
	}

	d := struct {
		TagName string `json:"tag_name"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return
	}

	return d.TagName, nil
}

func checkNewVersion(currentVersionTag string) {
	const latestReleasePage = "https://github.com/EndlessCheng/mahjong-helper/releases/latest"

	latestVersionTag, err := fetchLatestVersionTag()
	if err != nil {
		// 설명
		return
	}

	if latestVersionTag > currentVersionTag {
		color.HiGreen("새 버전 발견: %s! %s 에서 다운로드하세요", latestVersionTag, latestReleasePage)
	}
}

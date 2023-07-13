/*
Copyright 2021 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"regexp"
	"strings"
	"unicode"

	ref "github.com/containers/image/docker/reference"
	"github.com/mozillazg/go-pinyin"
)

func GetJiraKeys(title string) (keys []string) {

	re := regexp.MustCompile("[a-zA-Z0-9]+-[0-9]+")
	keys = re.FindAllString(title, -1)
	return
}

func ReplaceWrapLine(script string) string {
	return strings.Replace(strings.Replace(
		script,
		"\r\n",
		"\n",
		-1,
	), "\r", "\n", -1)
}

// Test case reference https://github.com/containers/image/blob/main/docker/reference/reference_test.go
func ExtractImageName(image string) string {
	imageNameStr := ""

	reference, err := ref.Parse(image)
	if err != nil {
		return imageNameStr
	}
	if named, ok := reference.(ref.Named); ok {
		imageNameArr := strings.Split(named.Name(), "/")
		imageNameStr = imageNameArr[len(imageNameArr)-1]
	}

	return imageNameStr
}

func GetImageNameFromContainerInfo(imageName, containerName string) string {
	if imageName == "" {
		return containerName
	}
	return imageName
}

func ContainsChinese(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func GetPinyinFromChinese(han string) (string, string) {
	firstLetter := ""
	fullLetter := ""
	a := pinyin.NewArgs()
	pinyinArr := pinyin.Pinyin(han, a)
	for _, pinyin := range pinyinArr {
		for _, l := range pinyin {
			fullLetter += l
			firstLetter += string(l[0])
		}
	}
	return fullLetter, firstLetter
}

func RemoveExtraSpaces(input string) string {
	// Remove spaces before and after strings
	trimmed := strings.TrimSpace(input)

	// Replace multiple consecutive spaces in the middle of a string with a single space
	regex := regexp.MustCompile(`\s+`)
	normalized := regex.ReplaceAllString(trimmed, " ")

	return normalized
}

package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"github.com/disintegration/imaging"
	"github.com/mozillazg/go-pinyin"
)

func handle9Scale(file string, left int, top int, right int, bottom int) {
	src, err := imaging.Open(file)
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
		return
	}

	src_tl := imaging.CropAnchor(src, left, top, imaging.TopLeft)
	src_tr := imaging.CropAnchor(src, right, top, imaging.TopRight)
	src_bl := imaging.CropAnchor(src, left, bottom, imaging.BottomLeft)
	src_br := imaging.CropAnchor(src, right, bottom, imaging.BottomRight)

	dst := imaging.New(left+right, top+bottom, color.NRGBA{0, 0, 0, 0})
	dst = imaging.Paste(dst, src_tl, image.Pt(0, 0))
	dst = imaging.Paste(dst, src_tr, image.Pt(left, 0))
	dst = imaging.Paste(dst, src_bl, image.Pt(0, top))
	dst = imaging.Paste(dst, src_br, image.Pt(left, top))

	err = imaging.Save(dst, file)
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
}

func gitUpload(cfg ChopperCfg) {
	if cfg.Git.Password == "" || cfg.Git.UserName == "" || cfg.Git.URL == "" {
		return
	}
	dir := path.Join(cfg.DirPath, ".remote")
	d, err := os.Stat(dir)
	if err != nil {
		return
	}
	if !d.IsDir() {
		return
	}
}

func export(cfg ChopperCfg, win fyne.Window) {
	f, err := os.Stat(cfg.DirPath)
	if err != nil {
		dialog.NewError(err, win)
		return
	}
	if !f.IsDir() {
		dialog.NewError(errors.New("目标位置不是一个文件夹"), win)
		return
	}
	prog := dialog.NewProgressInfinite("导出", "正在导出", win)
	prog.Show()

	files, err := ioutil.ReadDir(cfg.DirPath)
	if err != nil {
		dialog.NewError(err, win)
		return
	}
	pyArgs := pinyin.NewArgs()
	pyArgs.Style = pinyin.Tone3
	pyArgs.Fallback = func(r rune, a pinyin.Args) []string {
		// 去掉空格
		if r == 32 {
			return []string{}
		} else {
			return []string{
				string(r),
			}
		}
	}

	regType := regexp.MustCompile(`^@.+_`)
	reg9Scale := regexp.MustCompile(`#\([\d|,]+\)`)
	for _, file := range files {
		fileName := file.Name()
		targetName := fileName

		// 替换前缀类型: 按钮 -> btn; 背景 -> bg; 图标 -> icon; 预览 -> preview
		loc := regType.FindStringIndex(targetName)
		if len(loc) > 0 {
			typeName := targetName[:loc[1]]
			typeTag := ""
			switch typeName {
			case "@按钮_":
				typeTag = "btn_"
			case "@背景_":
				typeTag = "bg_"
			case "@图标_":
				typeTag = "icon_"
			case "@预览_":
				typeTag = "preview_"
			}
			targetName = typeTag + targetName[loc[1]:]
		}

		// 处理九宫格图片
		loc = reg9Scale.FindStringIndex(targetName)
		if len(loc) > 0 {
			// 去掉 @( )
			scaleTag := targetName[loc[0]+2 : loc[1]-1]
			scaleStrs := strings.Split(scaleTag, ",")
			var scaleNums []int
			for _, s := range scaleStrs {
				num, err := strconv.Atoi(s)
				if err == nil {
					scaleNums = append(scaleNums, num)
				}
			}
			var left, top, right, bottom int
			if len(scaleNums) == 1 {
				left, top, right, bottom = scaleNums[0], scaleNums[0], scaleNums[0], scaleNums[0]
			} else if len(scaleNums) == 2 {
				left, right = scaleNums[0], scaleNums[0]
				top, bottom = scaleNums[1], scaleNums[1]
			} else if len(scaleNums) == 3 {
				left = scaleNums[0]
				top, bottom = scaleNums[1], scaleNums[1]
				right = scaleNums[2]
			} else {
				left = scaleNums[0]
				top = scaleNums[1]
				right = scaleNums[2]
				bottom = scaleNums[3]
			}
			handle9Scale(path.Join(cfg.DirPath, fileName), left, top, right, bottom)
			fileBytes := []byte(fileName)
			targetName = string(append(fileBytes[:loc[0]], fileBytes[loc[1]:]...))
		}
		if strings.HasSuffix(fileName, ".png") || strings.HasSuffix(fileName, ".jpg") {
			newName := strings.Join(pinyin.LazyPinyin(targetName, pyArgs), "")
			if targetName == newName {
				continue
			}
			err = os.Rename(path.Join(cfg.DirPath, fileName), path.Join(cfg.DirPath, newName))
			if err != nil {
				fmt.Printf("rename error :%s\n", newName)
			}
		}
	}

	gitUpload(cfg)

	prog.Hide()
}

package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"github.com/disintegration/imaging"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/mozillazg/go-pinyin"
)

func hashFile(file string) string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

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

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func gitUpload(cfg ChopperCfg, files []string) error {
	if len(files) == 0 {
		return nil
	}
	if cfg.Git.Password == "" || cfg.Git.UserName == "" || cfg.Git.URL == "" {
		return nil
	}
	dir := path.Join(cfg.DirPath, ".remote")
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		_, err := git.PlainClone(dir, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: cfg.Git.UserName,
				Password: cfg.Git.Password,
			},
			URL:      cfg.Git.URL,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}
	}
	d, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !d.IsDir() {
		return err
	}
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	ref, err := r.Head()
	if err != nil {
		return err
	}
	err = w.Reset(&git.ResetOptions{
		Commit: ref.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		return err
	}
	_ = w.Pull(&git.PullOptions{RemoteName: "origin"})

	for _, f := range files {
		_ = copyFile(path.Join(cfg.DirPath, f), path.Join(dir, f))
	}

	s, err := w.Status()
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return nil
	}

	_, err = w.Add(".")
	if err != nil {
		return err
	}

	_, err = w.Commit("update res", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "chopper",
			Email: "chopper@didiapp.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: cfg.Git.UserName,
			Password: cfg.Git.Password,
		},
	})

	return err
}

func export(cfg ChopperCfg, win fyne.Window) {
	if cfg.DirPath == "" {
		dialog.NewError(errors.New("目标文件夹没有配置"), win)
		return
	}
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

	var dstFiles []string
	regType := regexp.MustCompile(`^@.+?-`)
	reg9Scale := regexp.MustCompile(`#\([\d|,]+\)`)
	for _, file := range files {
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".png") || strings.HasSuffix(fileName, ".jpg") {
			targetName := fileName

			// 替换前缀类型: 按钮 -> btn; 背景 -> bg; 图标 -> icon; 预览 -> preview
			loc := regType.FindStringIndex(targetName)
			if len(loc) > 0 {
				typeName := targetName[:loc[1]]
				typeTag := ""
				switch typeName {
				case "@按钮-":
					typeTag = "btn_"
				case "@背景-":
					typeTag = "bg_"
				case "@图标-":
					typeTag = "icon_"
				case "@预览-":
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
			newName := strings.Join(pinyin.LazyPinyin(targetName, pyArgs), "")
			if targetName == newName {
				continue
			}
			dstFile := path.Join(cfg.DirPath, newName)
			err = os.Rename(path.Join(cfg.DirPath, fileName), dstFile)
			fmt.Println("hash ", hashFile(dstFile))
			if err != nil {
				fmt.Printf("rename error :%s\n", newName)
			}
			dstFiles = append(dstFiles, newName)
		}
	}

	err = gitUpload(cfg, dstFiles)
	if err != nil {
		prog.Hide()
		dialog.NewError(err, win)
	} else {
		if err != nil {
			log.Fatalln(err)
		}
		if len(dstFiles) > 0 {
			err = robot(cfg)
			prog.Hide()
			if err != nil {
				dialog.NewError(err, win)
			} else {
				dialog.NewInformation("Info", "文件已重新命名:\n"+strings.Join(dstFiles, "\n"), win)
			}
		} else {
			prog.Hide()
			dialog.NewInformation("Info", "没有可命名的文件:\n", win)
		}
	}

}

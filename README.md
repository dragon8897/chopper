## 目的

### 美术使用

辅助美术将 `sketch` 软件切好的图片:

1. 改名: 中文 -> 拼音
2. 上传: 通过监视文件夹变动, 使用 git 上传图片

### 程序员使用(待定)

维护一个 map 表, 当图片资源 git 更新后, 自动按文件名或文件 md5 值同步更新图片

## lib

### 拼音库

[github.com/mozillazg/go-pinyin](https://github.com/mozillazg/go-pinyin)

### git 库

[github.com/go-git/go-git](https://github.com/go-git/go-git)

### gui 库

[github.com/fyne-io/fyne](https://github.com/fyne-io/fyne)

问题:

- 中文乱码:
  - 使用 `os.Setenv()` 设置 `FYNE_FONT` 环境变量, 指向一个中文的 **TTF** 字体, 然后 `defer os.Unsetenv()`

## to-do list

- [ ] 创建配置界面
  - [ ] 目标文件夹
  - [ ] git 配置项
- [ ] 启动子进程
  - [ ] 更改文件名
  - [ ] 上传 git (使用隐藏文件夹)
    - [ ] git pull
    - [ ] 拷贝并替换图片资源
    - [ ] git push

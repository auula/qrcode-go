// Since the release version of the original author's warehouse has not been updated,
// It’s just that the warehouse code has been updated, and it doesn’t modify the ‘main’ in the release.
// All my forks come and repost.
// 由于原作者的那个仓库release版本没有更新，
// 只是仓库代码更新了，并没有修改发行版本里面的‘main’
// 所有本人fork过来重新发布
// https://github.com/lihaotian0607/qrcode/issues/1

package qrcode

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"

	_ "image/jpeg"
)

type Avatar struct {
	Src    string // 头像地址
	Width  int    // 头像宽度
	Height int    // 头像高度
}

type BackgroundImage struct {
	Src    string
	X      int
	Y      int
	Width  int
	Height int
}

type ForegroundImage struct {
	Src string
}

type QrCode struct {
	qr                    *qrcode.QRCode
	Avatar                *Avatar
	ForegroundImage       *ForegroundImage
	BackgroundImage       *BackgroundImage
	CreateBackgroundImage Generate
	CreateForegroundImage Generate
	CreateAvatar          Generate
}

type Generate func(image.Image) (image.Image, error)

const (
	// Level L: 7% error recovery.
	Low = qrcode.Low

	// Level M: 15% error recovery. Good default choice.
	Medium = qrcode.Medium

	// Level Q: 25% error recovery.
	High = qrcode.High

	// Level H: 30% error recovery.
	Highest = qrcode.Highest
)

type IQrCode interface {
	// 设置头像
	SetAvatar(*Avatar)

	// 设置背景图
	SetBackgroundImage(*BackgroundImage)

	// 设置背景颜色
	SetBackgroundColor(color.Color)

	// 设置前景图
	SetForegroundImage(string)

	// 设置前景颜色
	SetForegroundColor(color.Color)

	DisableBorder(bool)

	// 返回生成的二维码图片
	Image(size int) (image.Image, error)

	// 返回 png 二维码图片
	PNG(size int) ([]byte, error)

	// 将二维码以PNG写入io.Writer
	Write(size int, out io.Writer) error

	// 将二维码以PNG写入指定的文件
	WriteFile(size int, filename string) error
}

func New(content string, level qrcode.RecoveryLevel) (*QrCode, error) {

	qr, err := qrcode.New(content, level)
	if err != nil {
		return nil, err
	}

	qrCode := &QrCode{}
	qrCode.qr = qr

	qrCode.SetCreateAvatar(qrCode.DefaultCreateAvatar)

	qrCode.SetCreateBackgroundImage(qrCode.DefaultCreateBackgroundImage)

	qrCode.SetCreateForegroundImage(qrCode.DefaultCreateForegroundImage)

	return qrCode, nil

}

// 设置头像
func (q *QrCode) SetAvatar(avatar *Avatar) {
	q.Avatar = avatar
}

func (q *QrCode) SetCreateAvatar(create Generate) {
	q.CreateAvatar = create
}

func (q *QrCode) SetCreateBackgroundImage(create Generate) {
	q.CreateBackgroundImage = create
}

func (q *QrCode) SetCreateForegroundImage(create Generate) {
	q.CreateForegroundImage = create
}

// 设置背景图
func (q *QrCode) SetBackgroundImage(img *BackgroundImage) {
	q.BackgroundImage = img
}

// 设置背景颜色
func (q *QrCode) SetBackgroundColor(color color.Color) {
	q.qr.BackgroundColor = color
}

// 设置前景图
func (q *QrCode) SetForegroundImage(src string) {
	q.ForegroundImage = &ForegroundImage{Src: src}
}

// 设置前景颜色
func (q *QrCode) SetForegroundColor(color color.Color) {
	q.qr.ForegroundColor = color
}

func (q *QrCode) DisableBorder(disable bool) {
	q.qr.DisableBorder = disable
}

// 返回生成的二维码图片
func (q *QrCode) Image(size int) (image.Image, error) {
	img := q.qr.Image(size)
	var err error

	if q.ForegroundImage != nil {
		if img, err = q.CreateForegroundImage(img); err != nil {
			return nil, err
		}
	}

	if q.Avatar != nil {
		if img, err = q.CreateAvatar(img); err != nil {
			return nil, err
		}
	}

	if q.BackgroundImage != nil {
		if img, err = q.CreateBackgroundImage(img); err != nil {
			return nil, err
		}
	}

	return img, nil
}

// 返回 png 二维码图片
func (q *QrCode) PNG(size int) ([]byte, error) {
	img, err := q.Image(size)
	if err != nil {
		return nil, err
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}

	var b bytes.Buffer
	err = encoder.Encode(&b, img)

	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// 将二维码以PNG写入io.Writer
func (q *QrCode) Write(size int, out io.Writer) error {
	var p []byte

	p, err := q.PNG(size)

	if err != nil {
		return err
	}
	_, err = out.Write(p)
	return err
}

// 将二维码以PNG写入指定的文件
func (q *QrCode) WriteFile(size int, filename string) error {
	var p []byte

	p, err := q.PNG(size)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, p, os.FileMode(0644))
}

// 图片转png
func imageToPng(img image.Image) (image.Image, error) {

	var reader = bytes.NewBuffer(nil)
	err := png.Encode(reader, img)
	if err != nil {
		return nil, err
	}

	img, err = png.Decode(reader)
	if err != nil {
		log.Fatalf("err %s", err.Error())
		return nil, err
	}

	return img, nil
}

// 默认创建头像的方法
func (q *QrCode) DefaultCreateAvatar(img image.Image) (image.Image, error) {
	avatar, err := os.Open(q.Avatar.Src)
	if err != nil {
		return nil, fmt.Errorf("open avatar file error: %s", err.Error())
	}

	defer avatar.Close()

	decode, _, err := image.Decode(avatar)
	if err != nil {
		return nil, err
	}

	decode = resize.Resize(uint(q.Avatar.Width), uint(q.Avatar.Height), decode, resize.Lanczos3)

	b := img.Bounds()

	// 设置为居中
	offset := image.Pt((b.Max.X-decode.Bounds().Max.X)/2, (b.Max.Y-decode.Bounds().Max.Y)/2)

	m := image.NewRGBA(b)

	draw.Draw(m, b, img, image.Point{X: 0, Y: 0}, draw.Src)

	draw.Draw(m, decode.Bounds().Add(offset), decode, image.Point{X: 0, Y: 0}, draw.Over)

	return m, err
}

// 默认创建背景图的方法
func (q *QrCode) DefaultCreateBackgroundImage(img image.Image) (image.Image, error) {
	file, err := os.Open(q.BackgroundImage.Src)
	if err != nil {
		return nil, fmt.Errorf("打开背景图文件失败 %s", err.Error())
	}

	img = resize.Resize(uint(q.BackgroundImage.Width), uint(q.BackgroundImage.Height), img, resize.Lanczos3)

	defer file.Close()

	bg, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	offset := image.Pt(q.BackgroundImage.X, q.BackgroundImage.Y)

	b := bg.Bounds()

	m := image.NewRGBA(b)

	draw.Draw(m, b, bg, image.Point{X: 0, Y: 0}, draw.Src)

	draw.Draw(m, img.Bounds().Add(offset), img, image.Point{X: 0, Y: 0}, draw.Over)

	return m, nil
}

// 默认创建前景图的方法
func (q *QrCode) DefaultCreateForegroundImage(img image.Image) (image.Image, error) {

	file, err := os.Open(q.ForegroundImage.Src)
	if err != nil {
		return nil, fmt.Errorf("打开前景图文件失败 %s", err.Error())
	}

	defer file.Close()

	decode, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	// 获取二维码的宽高
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y

	// 获取要填充的图片宽高
	foregroundW, foregroundH := decode.Bounds().Max.X, decode.Bounds().Max.Y

	if width != foregroundW || height != foregroundH {
		// 如果不一致将填充图剪裁
		decode = resize.Resize(uint(width), uint(height), decode, resize.Lanczos3)
	}

	m := image.NewRGBA(img.Bounds())
	d := image.NewRGBA(decode.Bounds())

	draw.Draw(m, m.Bounds(), img, image.Point{X: 0, Y: 0}, draw.Src)
	draw.Draw(d, d.Bounds(), decode, image.Point{X: 0, Y: 0}, draw.Src)

	for y := 0; y < img.Bounds().Max.X; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {

			// 检测像素是否为白色或者透明色
			if m.At(x, y).(color.RGBA).R == 255 && m.At(x, y).(color.RGBA).G == 255 && m.At(x, y).(color.RGBA).B == 255 && m.At(x, y).(color.RGBA).A == 255 {
				continue
			}

			if m.At(x, y).(color.RGBA) == q.qr.BackgroundColor {
				continue
			}

			// 填充颜色
			m.Set(x, y, color.RGBA{R: d.At(x, y).(color.RGBA).R, G: d.At(x, y).(color.RGBA).G, B: d.At(x, y).(color.RGBA).B, A: d.At(x, y).(color.RGBA).A})
		}
	}

	return m, nil
}

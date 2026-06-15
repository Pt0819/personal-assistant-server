package avatar

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var palette = []color.RGBA{
	{0xE5, 0x39, 0x35, 0xFF},
	{0xD8, 0x1B, 0x60, 0xFF},
	{0x8E, 0x24, 0xAA, 0xFF},
	{0x5E, 0x35, 0xB1, 0xFF},
	{0x39, 0x49, 0xAB, 0xFF},
	{0x1E, 0x88, 0xE5, 0xFF},
	{0x03, 0x9B, 0xE5, 0xFF},
	{0x00, 0xAC, 0xC1, 0xFF},
	{0x00, 0x89, 0x7B, 0xFF},
	{0x43, 0xA0, 0x47, 0xFF},
	{0x7C, 0xB3, 0x42, 0xFF},
	{0xC0, 0xCA, 0x33, 0xFF},
	{0xF4, 0x51, 0x1E, 0xFF},
	{0x6D, 0x4C, 0x41, 0xFF},
	{0x55, 0x6B, 0x2F, 0xFF},
	{0x20, 0x82, 0xAE, 0xFF},
}

const (
	avatarSize = 256
	fontSize   = 128
)

// Generate 生成首字符头像 PNG，返回 PNG 字节数据
// userID 用于确定性选色
// nickname 用于提取首字符（中文转拼音首字母）
func Generate(userID uint, nickname string) ([]byte, error) {
	initial := getInitial(nickname)
	bgColor := palette[userID%uint(len(palette))]

	img := image.NewRGBA(image.Rect(0, 0, avatarSize, avatarSize))

	// 填充背景
	for y := 0; y < avatarSize; y++ {
		for x := 0; x < avatarSize; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// 渲染白色首字母
	if err := drawText(img, initial); err != nil {
		// 渲染失败退回纯色背景头像
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// getInitial 提取昵称首字母
func getInitial(nickname string) string {
	if nickname == "" {
		return "?"
	}

	runes := []rune(nickname)
	first := runes[0]

	// ASCII 字母直接返回大写
	if (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') {
		return strings.ToUpper(string(first))
	}

	// ASCII 数字直接返回
	if first >= '0' && first <= '9' {
		return string(first)
	}

	// 中日韩文字 → 转拼音，取首字母
	if unicode.Is(unicode.Han, first) {
		py := pinyin.Pinyin(string(first), pinyin.NewArgs())
		if len(py) > 0 && len(py[0]) > 0 {
			return strings.ToUpper(string(py[0][0][0]))
		}
	}

	return "?"
}

// drawText 在图片上居中绘制白色文字
func drawText(img *image.RGBA, text string) error {
	tt, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return err
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size: fontSize,
		DPI:  72,
	})
	if err != nil {
		return err
	}
	defer face.Close()

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: face,
	}

	// 测量文字宽度
	advance := d.MeasureString(text)
	x := (avatarSize*64 - advance.Ceil()) / 2

	// 垂直居中
	ascent := face.Metrics().Ascent.Ceil()
	descent := face.Metrics().Descent.Ceil()
	textHeight := ascent + descent
	y := (avatarSize + textHeight) / 2 - descent

	d.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	d.DrawString(text)
	return nil
}

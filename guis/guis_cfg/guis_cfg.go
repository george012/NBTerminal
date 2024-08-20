package guis_cfg

import (
	"github.com/george012/fltk_go"
	"github.com/kbinani/screenshot"
	"sync"
)

const (
	GUIHexColorWithGRAY       = 0x75757500 // en: Gray, zh-cn:灰色
	GUIHexColorWithLightGray  = 0xeeeeee00 // en: LightGray, zh-cn:亮灰色
	GUIHexColorWithBlue       = 0x42A5F500 // en: Blue, zh-cn:蓝色
	GUIHexColorWithSelBlue    = 0x2196F300 // en: sell blue, zh-cn:塞尔蓝
	GUIHexColorWithCustomCyan = 0x008B8B00
)

type ScreenSize struct {
	Width  int
	Height int
}

type Frame struct {
	X      int
	Y      int
	Width  int
	Height int
}

var (
	DefaultFont       = fltk_go.HELVETICA
	DefaultFontSize   = 16
	DefaultWindowSize = ScreenSize{
		Width:  1440,
		Height: 900,
	}
	screenSizeInstance *ScreenSize
	once               sync.Once
)

func GetScreenSize() *ScreenSize {
	once.Do(func() {
		bounds := screenshot.GetDisplayBounds(0)
		screenSizeInstance = &ScreenSize{
			Width:  bounds.Dx(),
			Height: bounds.Dy(),
		}
	})
	return screenSizeInstance
}

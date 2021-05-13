/**
 * @author:  ccforevercn<1253705861@qq.com>
 * @link     http://ccforever.cn
 * @license  https://github.com/ccforevercn
 * @date:    2021/5/12
 */
package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"
)


var (
	// 扫描的根目录
	address string
	// 当前文件目录
	currentPath string
	// 模板图片
	templateImage string
	// 背景图片根目录
	writeBackgroundImagePath string
	// 背景图片访问域名
	backgroundUrl string
	// 核心图片根目录
	writeCoreImagePath string
	// 核心图片访问域名
	coreUrl string
	// 保存的sql文件
	sqlFilePath string
	// 核心图片距离左侧距离
	left int
	// 核心图片距离顶部距离
	top int
	// 核心图片名称
	coreName string
	// 背景图片名称
	backgroundName string
	// 模板图片坐标轴结构体
	templateGridMap map[int]TemplateGrid
	// 模板图片宽度
	templateWidth int
	// 模板图片高度
	templateHeight int
	// sql文件内容
	sqlContent string
)

func init()  {
	// 设置当前文件目录
	currentPath, _ = os.Getwd()
	templateGridMap = make(map[int]TemplateGrid) // 创建内存
	sqlContent = "INSERT INTO table(top, left, background_url, core_url) VALUES"
}
// 保存模板图片坐标的结构体
type TemplateGrid struct {
	X int
	Y int
}

// 设置坐标轴
func (templateGrid *TemplateGrid) Set(x, y int)  {
	templateGrid.X = x
	templateGrid.Y = y
}

func main () {
	InputParam()
	Start()
}

/*
 * 获取输入参数
 */
func InputParam () {
	// 设置SQL文件保存目录
	fmt.Printf("请输入SQL文件保存目录，不输入为当前工作目录：")
	fmt.Scanln(&sqlFilePath)
	sqlFilePathLen := len(sqlFilePath) > 0 // 验证SQL文件保存目录是否为空，如果为空则重置SQL文件保存目录为当前路径
	if !sqlFilePathLen {
		sqlFilePath = currentPath
	}
	os.MkdirAll(sqlFilePath, os.ModeDir) // 创建保存SQL文件的根目录
	sqlFilePath += string(os.PathSeparator) // 加入间隔符

	// 获取背景图片访问域名
	fmt.Printf("请输入背景图片访问的域名(结尾+/)，不输入默认为空：")
	fmt.Scanln(&backgroundUrl)

	// 获取核心图片访问域名
	fmt.Printf("请输入核心图片访问的域名(结尾+/)，不输入默认为空：")
	fmt.Scanln(&coreUrl)

	// 获取保存背景图片的根目录
	fmt.Printf("请输入保存背景图片的根目录，不输入为当前工作目录：")
	fmt.Scanln(&writeBackgroundImagePath)
	writeBackgroundImagePathLen := len(writeBackgroundImagePath) > 0 // 验证背景图片的根目录是否为空，如果为空则重置背景图片的根目录为当前路径
	if !writeBackgroundImagePathLen {
		writeBackgroundImagePath = currentPath
	}
	os.MkdirAll(writeBackgroundImagePath, os.ModeDir) // 创建保存背景图片的根目录
	writeBackgroundImagePath += string(os.PathSeparator) // 加入间隔符

	// 获取保存核心图片的根目录
	fmt.Printf("请输入保存核心图片的根目录，不输入为当前工作目录：")
	fmt.Scanln(&writeCoreImagePath)
	writeCoreImagePathPathLen := len(writeCoreImagePath) > 0 // 验证核心图片的根目录是否为空，如果为空则重置核心图片的根目录为当前路径
	if !writeCoreImagePathPathLen {
		writeCoreImagePath = currentPath
	}
	os.MkdirAll(writeCoreImagePath, os.ModeDir) // 创建保存核心图片的根目录
	writeCoreImagePath += string(os.PathSeparator) // 加入间隔符

	// 获取模板图片地址
	for  {
		fmt.Printf("请输入模板图片地址(必须是png格式的图片)：")
		fmt.Scanln(&templateImage)
		templateImageLen := len(templateImage) > 0
		if templateImageLen {
			templateImageExt := path.Ext(templateImage) // 文件扩展名
			if templateImageExt == ".png" {
				break
			}
		}
	}

	// 获取原始图片目录
	fmt.Printf("请输入原始图片目录，不输入为当前工作目录：")
	fmt.Scanln(&address)
	addressLen := len(address) > 0 // 验证原始图片目录是否为空，如果为空则重置原始图片目录为当前路径
	if !addressLen {
		address = currentPath
	}
	address += string(os.PathSeparator) // 加入间隔符
	if address == writeBackgroundImagePath || address == writeCoreImagePath {
		fmt.Println("原始图片目录不能和背景图片目录、核心图片目录一样")
		os.Exit(0)
	}
}

/*
 * 开始处理图片
 */
func Start () {
	fileInfoList, _ := ioutil.ReadDir(address)
	for i := range fileInfoList {
		filePath :=  address + string(os.PathSeparator) // 打开的文件或者文件夹追加系统路径分隔符
		fileName := filePath + fileInfoList[i].Name() // 设置文件夹或者文件名称绝对路径
		fileExt := path.Ext(fileName) // 文件扩展名
		if fileExt == ".jpg" || fileExt == ".png" || fileExt == ".jpeg" {
			GetTemplateGrid()
			CoreImage(fileName)
			backgroundImage(fileName)
			SqlFile()
		}
	}
}

/*
 * 生成图片文件名称
 */
func SetMd5(str string) string  {
	byteStr := []byte(str)
	md5Str := md5.New()
	md5Str.Write(byteStr)
	return hex.EncodeToString(md5Str.Sum(nil))
}

/*
 * 获取模板图片坐标轴
 */
func GetTemplateGrid()  {
	template, _ := os.Open(templateImage) // 读取模板图片内容
	defer template.Close() // 关闭模板
	templateDecode,_ , _ := image.Decode(template) // 获取模板图像
	templateWidth = templateDecode.Bounds().Dx() // 获取模板图片的宽度
	templateHeight = templateDecode.Bounds().Dy() // 获取模板图片的高度
	loop := 0 // 设置主键
	for xi := 0; xi <= templateWidth; xi++ {
		for xj := 0; xj <= templateHeight; xj++ {
			r, g, b, _ := templateDecode.At(xi, xj).RGBA()  // 获取模板图片颜色值
			if r != 0 && g != 0 && b != 0 { // 颜色不为透明时保存
				templateGridMap[loop] = TemplateGrid{xi, xj} // 保存到模板图片坐标结构体中
				loop++ // 主键累加
			}
		}
	}
}

/*
 * 创建核心图片
 */
func CoreImage(imageName string) {
	source, _ := os.Open(imageName) // 获取原始图片
	defer source.Close() // 关闭原始图片
	sourceDecode,_ ,_ := image.Decode(source) // 获取原始图片图像
	rand.Seed(time.Now().UnixNano()) // 随机值设置初始状态
	leftRandMax :=  sourceDecode.Bounds().Dx() - templateWidth // 获取左边距的最大距离
	topRandMax :=  sourceDecode.Bounds().Dy() - templateHeight // 获取顶边距的最大距离
	left = rand.Intn(leftRandMax) // 随机获取左边距
	top = rand.Intn(topRandMax) // 随机获取高边距
	var tempImage image.Image // 创建临时图片
	// 原始图片解码
	if tempRgbImage, ok := sourceDecode.(*image.YCbCr); ok {
		tempImage = tempRgbImage.SubImage(image.Rect(left, top, templateWidth + left, templateHeight + top)).(*image.YCbCr)
	} else if tempRgbImage, ok := sourceDecode.(*image.RGBA); ok {
		tempImage = tempRgbImage.SubImage(image.Rect(left, top, templateWidth + left, templateHeight + top)).(*image.RGBA)
	} else if tempRgbImage, ok := sourceDecode.(*image.NRGBA); ok {
		tempImage = tempRgbImage.SubImage(image.Rect(left, top, templateWidth + left, templateHeight + top)).(*image.NRGBA)
	} else {
		log.Println("原始图片目录不能和背景图片目录、核心图片目录一样")
		os.Exit(0)
	}
	tempName := writeCoreImagePath + strconv.Itoa(rand.Intn(999999)) + "_temp.png" // 临时图片名称
	tempFile, _ := os.Create(tempName)  // 创建临时图片
	defer os.Remove(tempName) // 删除临时图片
	defer tempFile.Close() // 关闭临时图片
	png.Encode(tempFile, tempImage) // 写入临时图片为PNG格式
	tempCore, _ := os.Open(tempName) // 读取临时图片内容
	defer tempCore.Close() // 关闭临时文件
	tempCoreDecode, _, _ := image.Decode(tempCore) // 获取临时图片内容
	coreImage := image.Rect(0, 0, templateWidth, templateHeight) // 创建核心图片
	coreRgba := image.NewRGBA(coreImage) // 创建核心图片RGBA
	for _, tplValue := range templateGridMap {
		coreRgba.Set(tplValue.X, tplValue.Y, tempCoreDecode.At(tplValue.X, tplValue.Y)) // 设置模板坐标轴对应的临时图片颜色到核心图片
	}
	coreName = SetMd5(strconv.Itoa(rand.Intn(999999)) + "_core") + ".png"  // 核心图片名称
	coreFile, _ := os.Create(writeCoreImagePath + coreName)  // 创建核心图片
	defer coreFile.Close() // 关闭核心图片
	png.Encode(coreFile, coreRgba) // 写入核心图片为PNG格式
}

/*
 * 创建背景图片
 */
func backgroundImage(imageName string)  {
	source, _ := os.Open(imageName) // 获取原始图片
	defer source.Close() // 关闭原始图片
	sourceDecode,_ ,_ := image.Decode(source) // 获取原始图片图像
	backgroundRgba := image.NewRGBA64(sourceDecode.Bounds()) // 创建背景图片
	cx := sourceDecode.Bounds().Dx() // 获取原始图片的宽度
	cy := sourceDecode.Bounds().Dy() // 获取原始图片的高度
	for ci := 0; ci <= cx; ci++ {
		for cj := 0; cj <= cy; cj++ {
			r, g, b, a := sourceDecode.At(ci, cj).RGBA()  // 获取原始图片坐标对应的色值
			opacity := uint16(float64(a)*1) // 设置透明度
			v := backgroundRgba.ColorModel().Convert(color.NRGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: opacity}) // rgba色值转换
			rr, gg, bb, aa := v.RGBA() // 获取转换后的rgba色值
			backgroundRgba.SetRGBA64(ci, cj, color.RGBA64{R: uint16(rr), G: uint16(gg), B: uint16(bb), A: uint16(aa)}) // 设置色值到图片上
		}
	}
	percentage := 0.4 // 透明度
	for _, tplValue := range templateGridMap {
		r, g, b, a := sourceDecode.At(tplValue.X + left, tplValue.Y + top).RGBA() // 获取原始图片坐标对应的色值
		opacity := uint16(float64(a)*percentage) // 设置透明度
		v := backgroundRgba.ColorModel().Convert(color.NRGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: opacity}) // rgba色值转换
		rr, gg, bb, aa := v.RGBA() // 获取转换后的rgba色值
		backgroundRgba.SetRGBA64(tplValue.X + left, tplValue.Y + top, color.RGBA64{R: uint16(rr), G: uint16(gg), B: uint16(bb), A: uint16(aa)}) // 设置色值到图片上
	}
	backgroundName = SetMd5(strconv.Itoa(rand.Intn(999999)) + "_image") + ".jpg"  // 新的背景图片名称
	backgroundFile, _ := os.Create(writeBackgroundImagePath + backgroundName) // 创建文件
	defer backgroundFile.Close()  // 关闭文件
	jpeg.Encode(backgroundFile, backgroundRgba, nil) // 将图像以JPEG格式写入到图片内
}

/*
 * 创建sql文件
 */
func SqlFile()  {
	sqlContent += "(" + strconv.Itoa(top) + "," + strconv.Itoa(left) + "," + backgroundUrl + backgroundName + "," + coreUrl + coreName + "),"
	sqlFileName := sqlFilePath + "images.sql"
	sqlFile, _ := os.Create(sqlFileName)
	defer sqlFile.Close()
	sqlFile.WriteString(sqlContent)
	sqlFile.Sync()
}
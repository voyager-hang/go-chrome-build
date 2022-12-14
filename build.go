package go_chrome_build

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func DoBuild(sysType string) {
	conf := getConfig()
	if conf.IntegratedBrowser {
		// 获取浏览器位置
		browserPath, browserName, packageMsg := getBrowserPath(sysType)
		// 获取打包浏览器位置
		runPath := GetWorkingDirPath()
		browserDir := runPath + "/resources/browser"
		err := createDir(browserDir)
		if err != nil {
			EchoError("create dir browser error")
			return
		}
		newBrowserPath := browserDir + "/" + browserName
		defer os.Remove(newBrowserPath)
		err = createDir(newBrowserPath)
		if err != nil {
			EchoError("copy browser dir err: " + err.Error())
		}
		if !IsExist(browserPath) {
			DownBrowser(packageMsg.sysStruct, packageMsg.version, packageMsg.osName, browserPath)
		}
		_, err = copyFile(newBrowserPath, browserPath)
		if err != nil {
			EchoError("copy browser err: " + err.Error())
			return
		}
	}
	buildDir := []string{
		"./resources/...",
	}
	err := Pack(sysType+"_build.go", "main", buildDir)
	if err != nil {
		EchoError(err.Error())
	} else {
		EchoSuccess("Build Compilation complete.")
	}
}

func PackWindows() {
	DoBuild("windows")
	err := BuildSysO()
	if err != nil {
		EchoError(err.Error())
		os.Exit(0)
		return
	}
	sysType := runtime.GOOS
	switch sysType {
	case "darwin":
		// macos
		darwinPackWindows()
	case "linux":
		linuxPackWindows()
	case "windows":
		windowsPackWindows()
	}
}

func PackLinux() {
	DoBuild("linux")
	sysType := runtime.GOOS
	switch sysType {
	case "darwin":
		// macos
		darwinPackLinux()
	case "linux":
		linuxPackLinux()
	case "windows":
		windowsPackLinux()
	}
}

func PackMacOs() {
	DoBuild("darwin")
	sysType := runtime.GOOS
	switch sysType {
	case "darwin":
		// macos
		darwinPackDarwin()
	case "linux":
		linuxPackDarwin()
	case "windows":
		windowsPackDarwin()
	}
}

func PackNowSys() {
	sysType := runtime.GOOS
	switch sysType {
	case "darwin":
		// macos
		PackMacOs()
	case "linux":
		PackLinux()
	case "windows":
		PackWindows()
	}
}

func getBrowserPath(sysType string) (string, string, packageMsg) {
	// 浏览器不存在 下载到打包目录
	packConfig := getConfig()
	browserPath := ""
	browserName := ""
	packageData := packageMsg{
		sysStruct: "",
		version:   "",
		osName:    "",
	}
	switch sysType {
	case "darwin":
		browserName = "chrome-mac.zip"
		packageData.version = packConfig.ChromeVersion.Darwin
		packageData.osName = "darwin"
		if packConfig.DarwinAppleChip {
			packageData.sysStruct = "Mac_Arm"
			if packConfig.ChromePackPath.DarwinArm != "" && IsExist(packConfig.ChromePackPath.DarwinArm) {
				if !strings.HasSuffix(packConfig.ChromePackPath.DarwinArm, "chrome-mac.zip") {
					EchoError("filename must be chrome-mac.zip")
					os.Exit(1)
				}
				browserPath = packConfig.ChromePackPath.DarwinArm
			}
		} else {
			packageData.sysStruct = "Mac"
			if packConfig.ChromePackPath.Darwin != "" && IsExist(packConfig.ChromePackPath.Darwin) {
				if !strings.HasSuffix(packConfig.ChromePackPath.Darwin, "chrome-mac.zip") {
					EchoError("filename must be chrome-mac.zip")
					os.Exit(1)
				}
				browserPath = packConfig.ChromePackPath.Darwin
			}
		}
	case "linux":
		browserName = "chrome-linux.zip"
		if packConfig.ChromePackPath.Linux != "" && IsExist(packConfig.ChromePackPath.Linux) {
			if !strings.HasSuffix(packConfig.ChromePackPath.Linux, "chrome-linux.zip") {
				EchoError("filename must be chrome-linux.zip")
				os.Exit(1)
			}
			browserPath = packConfig.ChromePackPath.Linux
		}

		packageData.version = packConfig.ChromeVersion.Linux
		packageData.osName = "linux"
		packageData.sysStruct = "Linux_x64"
	case "windows":
		browserName = "chrome-win.zip"

		packageData.version = packConfig.ChromeVersion.Windows
		packageData.osName = "win"
		if packConfig.WindowsArch == "386" {
			packageData.sysStruct = "Win"
			if packConfig.ChromePackPath.Windows != "" && IsExist(packConfig.ChromePackPath.Windows) {
				if !strings.HasSuffix(packConfig.ChromePackPath.Windows, "chrome-win.zip") {
					EchoError("filename must be chrome-win.zip")
					os.Exit(1)
				}
				browserPath = packConfig.ChromePackPath.Windows
			}
		} else {
			packageData.sysStruct = "Win_x64"
			if packConfig.ChromePackPath.Windows64 != "" && IsExist(packConfig.ChromePackPath.Windows64) {
				if !strings.HasSuffix(packConfig.ChromePackPath.Windows64, "chrome-win.zip") {
					EchoError("filename must be chrome-win.zip")
					os.Exit(1)
				}
				browserPath = packConfig.ChromePackPath.Windows64
			}
		}
	}
	if browserPath == "" {
		browserPath = GetWorkingDirPath() + "/browser/" + packageData.sysStruct + "/" + browserName
	}
	return browserPath, browserName, packageData
}

func DownBrowser(sysStruct, version, osName, downPath string) string {
	sysStructArr := map[string]string{
		"Linux_x64": "",
		"Mac":       "",
		"Mac_Arm":   "",
		"Win":       "",
		"Win_x64":   "",
	}
	if _, ok := sysStructArr[sysStruct]; !ok {
		panic("System Architecture is not in the list;{Linux_x64,Mac,Mac_Arm,Win,Win_x64}")
	}
	downUrl := fmt.Sprintf("https://registry.npmmirror.com/-/binary/chromium-browser-snapshots/%s/%s/chrome-%s.zip",
		sysStruct, version, osName)
	if downPath == "" {
		panic("downPath is empty")
	}
	err := createDir(downPath)
	if err != nil {
		panic(err)
	}
	err = DownloadFile(downUrl, downPath+".tmp")
	if err != nil {
		_ = os.Remove(downPath + ".tmp")
		panic(err)
	} else {
		err = os.Rename(downPath+".tmp", downPath)
		if err != nil {
			panic(err)
		}
	}
	return downPath
}

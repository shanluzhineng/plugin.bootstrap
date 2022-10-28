package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/abmpio/upack/cmd"
	upack "github.com/abmpio/upack/pkg"
)

func init() {
	fmt.Println("pluginbootstrap init function called")
}

type Bootstrap struct {
	installedPluginList []*upack.InstalledPackage
}

type IBootstrap interface {
	BootstrapPlugin() error
}

func newBootstrap() *Bootstrap {
	b := &Bootstrap{
		installedPluginList: make([]*upack.InstalledPackage, 0),
	}
	return b
}

func (b *Bootstrap) BootstrapPlugin() (err error) {
	r := upack.PlugIns
	packages, err := r.ListInstalledPackages()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	for _, eachPackage := range packages {
		assetPlugInInstalled(eachPackage)
	}
	packages, err = r.ListInstalledPackages()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	for _, eachPackage := range packages {
		soFilePath := findPluginSoFileName(*eachPackage.Path)
		if len(soFilePath) <= 0 {
			fmt.Printf("插件目录 %s中没有找到插件对应的so文件,将忽略此插件\r\n", *eachPackage.Path)
			continue
		}
		fmt.Printf("准备加载插件 %s文件:%s\r\n", eachPackage.PackageName(), soFilePath)
		//打开目录
		currentPlug, err := plugin.Open(soFilePath)
		if err != nil {
			fmt.Printf("加载插件so文件时出现异常,异常信息:%s\r\n", err)
			continue
		}
		if currentPlug == nil {
			fmt.Printf("无法加载插件so文件\r\n")
			continue
		}
		bootstrapSymbol, err := currentPlug.Lookup("PluginBootstrap")
		if err != nil {
			fmt.Printf("尝试在插件中搜索PluginBootstrap符号时出现异常,异常信息:%s\r\n", err.Error())
			continue
		}
		bootstarp, ok := bootstrapSymbol.(IBootstrap)
		if !ok {
			fmt.Println("插件导出的PluginBootstrap符号必须实现IBootstrap接口")
			continue
		}
		err = bootstarp.BootstrapPlugin()
		if err != nil {
			fmt.Printf("调用插件的BootstrapPlugin方法时出现异常,异常信息:%s\r\n", err.Error())
			continue
		}
		fmt.Printf("插件加载成功,插件名:%s,目录:%s\r\n", eachPackage.PackageName(), soFilePath)
		b.installedPluginList = append(b.installedPluginList, eachPackage)
	}
	return nil
}

func (b *Bootstrap) InstalledPluginList() []*upack.InstalledPackage {
	return b.installedPluginList
}

func findPluginSoFileName(path string) string {
	files, err := os.ReadDir(path)
	if err != nil {
		return ""
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".so" {
			return filepath.Join(path, file.Name())
		}
	}
	return ""
}

func assetPlugInInstalled(pluginPackage *upack.InstalledPackage) {
	if pluginPackage.Path != nil && folderIsExist(*pluginPackage.Path) {
		//已经安装
		return
	}
	installedPackage := installedPackageJsonFileContains(pluginPackage.Group, pluginPackage.Name, pluginPackage.Version.String())
	if installedPackage != nil {
		//已经安装，检测文件夹是否已经下载
		if folderIsExist(*installedPackage.Path) {
			return
		}
	}
	fmt.Printf("缺少插件文件,%s,准备安装...\r\n", pluginPackage.PackageName())
	cmdInstall := &cmd.Install{
		PackageName: pluginPackage.PackageName(),
	}
	installResult := cmdInstall.Run()
	if installResult == 0 {
		fmt.Printf("插件 %s 安装成功,安装目录:%s\r\n", pluginPackage.PackageName(), cmdInstall.InstalledPath())
	} else {
		fmt.Printf("插件 %s 安装失败\r\n", pluginPackage.PackageName())
	}
}

func installedPackageJsonFileContains(groupName string, name string, version string) *upack.InstalledPackage {
	packages, err := upack.PlugIns.ListInstalledPackages()
	if err != nil || len(packages) <= 0 {
		return nil
	}
	for _, eachPackage := range packages {
		//比较组名与模块名称，忽略大小写
		result := strings.EqualFold(eachPackage.Group, groupName) && strings.EqualFold(eachPackage.Name, name)
		if !result || len(version) <= 0 {
			//不匹配，如没有版本则不比较版本
			continue
		}
		newVersion, err := upack.ParseUniversalPackageVersion(version)
		if err != nil {
			return nil
		}
		versionCompared := newVersion.Compare(eachPackage.Version)
		if versionCompared <= 0 {
			//已安装了最新的版本
			return eachPackage
		}
		//安装的是旧版本
		return nil
	}
	return nil
}

func folderIsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

var PluginBootstrap = newBootstrap()

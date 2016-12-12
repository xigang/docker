package builtins

import (
	"runtime"

	"github.com/docker/docker/api"
	apiserver "github.com/docker/docker/api/server"
	"github.com/docker/docker/daemon/networkdriver/bridge"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/engine"
	"github.com/docker/docker/events"
	"github.com/docker/docker/pkg/parsers/kernel"
	"github.com/docker/docker/registry"
)

func Register(eng *engine.Engine) error {
	//网络初始化，TODO后期阅读
	if err := daemon(eng); err != nil {
		return err
	}

	//web api初始化
	if err := remote(eng); err != nil {
		return err
	}

	//注册events事件的handler,通过这些api查看docker内部的事件信息，log信息
	if err := events.New().Install(eng); err != nil {
		return err
	}

	//版本信息
	if err := eng.Register("version", dockerVersion); err != nil {
		return err
	}

	//注册register handler ”auth”，向公有registry进行认证；”search”，在公有registry上搜索指定的镜像
	return registry.NewService().Install(eng)
}

// remote: a RESTful api for cross-docker communication
func remote(eng *engine.Engine) error {
	//serveapi 循环多种协议创建http.Server,服务client端请求
	if err := eng.Register("serveapi", apiserver.ServeApi); err != nil {
		return err
	}
	//acceptconnections 通知init守护进程，当Docker daemon启动时，可以让Docker daemon进程接受请求
	return eng.Register("acceptconnections", apiserver.AcceptConnections)
}

// daemon: a default execution and storage backend for Docker on Linux,
// with the following underlying components:
//
// * Pluggable storage drivers including aufs, vfs, lvm and btrfs.
// * Pluggable execution drivers including lxc and chroot.
//
// In practice `daemon` still includes most core Docker components, including:
//
// * The reference registry client implementation
// * Image management
// * The build facility
// * Logging
//
// These components should be broken off into plugins of their own.
//
func daemon(eng *engine.Engine) error {
	return eng.Register("init_networkdriver", bridge.InitDriver)
}

// builtins jobs independent of any subsystem
func dockerVersion(job *engine.Job) engine.Status {
	v := &engine.Env{}
	v.SetJson("Version", dockerversion.VERSION)
	v.SetJson("ApiVersion", api.APIVERSION)
	v.Set("GitCommit", dockerversion.GITCOMMIT)
	v.Set("GoVersion", runtime.Version())
	v.Set("Os", runtime.GOOS)
	v.Set("Arch", runtime.GOARCH)
	if kernelVersion, err := kernel.GetKernelVersion(); err == nil {
		v.Set("KernelVersion", kernelVersion.String())
	}
	if _, err := v.WriteTo(job.Stdout); err != nil {
		return job.Error(err)
	}
	return engine.StatusOK
}

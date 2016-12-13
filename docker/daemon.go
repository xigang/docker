// +build daemon

package main

import (
	"log"

	"github.com/docker/docker/builtins"
	"github.com/docker/docker/daemon"
	_ "github.com/docker/docker/daemon/execdriver/lxc"
	_ "github.com/docker/docker/daemon/execdriver/native"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/engine"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/signal"
)

const CanDaemon = true

var (
	daemonCfg = &daemon.Config{}
)

func init() {
	//InstallFlags对daemoncfg变量的各个属性进行赋值
	daemonCfg.InstallFlags()
}

func mainDaemon() {
	//判断剩余的参数是否为0，如果是0则正常启动daemon，否则标准输出help信息，退出
	if flag.NArg() != 0 {
		flag.Usage()
		return
	}
	//初始化一个docker engine对象,Engine是docker运行的核心模块，负责docker任务的调度管理
	//Engine扮演docker container存储仓库的角色，并通过job的形式管理这些容器。
	eng := engine.New()

	//engine信号捕获,保证docker daemo程序正常退出,
	//设置Trap特定信号的处理方法，特定信号有SIGINT，SIGTERM以及SIGQUIT
	signal.Trap(eng.Shutdown)

	// Load builtins
	// 为engine注册多个handler,便于后续执行相应的任务时，运行指定的handler
	// Handler包括:网络初始化，web API服务，事件查询，版本查看，Docker register验证及搜索
	if err := builtins.Register(eng); err != nil {
		log.Fatal(err)
	}

	// load the daemon in the background so we can immediately start
	// the http api so that connections don't fail while the daemon
	// is booting
	go func() {
		//1.创建一个daemon对象,初始化Docker daemon的基本环境，如处理config参数，验证系统的支持度，
		//配置Docker的工作目录，设置加载driver,创建graph环境。
		d, err := daemon.NewDaemon(daemonCfg, eng)
		if err != nil {
			log.Fatal(err)
		}

		//2.向eng对象中注册众多的handler
		if err := d.Install(eng); err != nil {
			log.Fatal(err)
		}
		// after the daemon is done setting up we can tell the api to start
		// accepting connections
		//3.当Docker daemon启动完毕时，运行acceptconnections job，
		//目的:向init守护进行发送信号，使Docker daemon能正常接受请求
		if err := eng.Job("acceptconnections").Run(); err != nil {
			log.Fatal(err)
		}
	}()
	// TODO actually have a resolved graphdriver to show?
	// 打印Docker版本及驱动信息
	log.Printf("docker daemon: %s %s; execdriver: %s; graphdriver: %s",
		dockerversion.VERSION,
		dockerversion.GITCOMMIT,
		daemonCfg.ExecDriver,
		daemonCfg.GraphDriver,
	)

	// Serve api
	// Serve api的创建及运行
	job := eng.Job("serveapi", flHosts...)       //让Docker Daemon提供API访问服务，flHosts提供使用的协议及监听地址
	job.SetenvBool("Logging", true)              //设置日志
	job.SetenvBool("EnableCors", *flEnableCors)  //设置跨域
	job.Setenv("Version", dockerversion.VERSION) //设置docker版本
	job.Setenv("SocketGroup", *flSocketGroup)    //设置socket组

	job.SetenvBool("Tls", *flTls) //设置服务端证书
	job.SetenvBool("TlsVerify", *flTlsVerify)
	job.Setenv("TlsCa", *flCa)
	job.Setenv("TlsCert", *flCert)
	job.Setenv("TlsKey", *flKey)
	job.SetenvBool("BufferRequests", true)
	if err := job.Run(); err != nil { //设置完毕相关的环境变量,运行serverapi服务
		log.Fatal(err)
	}
}

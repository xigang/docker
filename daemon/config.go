package daemon

import (
	"net"

	"github.com/docker/docker/daemon/networkdriver"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
)

const (
	defaultNetworkMtu    = 1500
	DisableNetworkBridge = "none"
)

// Config define the configuration of a docker daemon
// These are the configuration settings that you pass
// to the docker daemon when you launch it with say: `docker -d -e lxc`
// FIXME: separate runtime configuration from http api configuration
type Config struct {
	Pidfile                     string              //Docker Daemon所属进程的PID文件
	Root                        string              //Docker 运行时所使用的root路径
	AutoRestart                 bool                //如果为true,则支持docker run时的重启
	Dns                         []string            //Docker 使用DNS Server
	DnsSearch                   []string            //Docker使用指定的DNS查找域名
	EnableIptables              bool                //启用Docker的iptables功能
	EnableIpForward             bool                //启用net.ipv4.ip_forward功能
	DefaultIp                   net.IP              //绑定容器端口时，使用默认的ip
	BridgeIface                 string              //添加容器网络至已有的网桥
	BridgeIP                    string              //创建网桥的IP地址
	InterContainerCommunication bool                //是否允许相同host上容器间的通信
	GraphDriver                 string              //Docker运行时使用的特定存储驱动
	GraphOptions                []string            //可设置的存储驱动选项
	ExecDriver                  string              // Docker运行时使用的特定exec驱动
	Mtu                         int                 //设置容器网络的MTU
	DisableNetwork              bool                //有定义，之后未初始化
	EnableSelinuxSupport        bool                //启用SELinux功能的支持
	Context                     map[string][]string //有定义，之后未初始化
}

// InstallFlags adds command-line options to the top-level flag parser for
// the current process.
// Subsequent calls to `flag.Parse` will populate config with values parsed
// from the command-line.
func (config *Config) InstallFlags() {
	flag.StringVar(&config.Pidfile, []string{"p", "-pidfile"}, "/var/run/docker.pid", "Path to use for daemon PID file")
	flag.StringVar(&config.Root, []string{"g", "-graph"}, "/var/lib/docker", "Path to use as the root of the Docker runtime")
	flag.BoolVar(&config.AutoRestart, []string{"#r", "#-restart"}, true, "--restart on the daemon has been deprecated infavor of --restart policies on docker run")
	flag.BoolVar(&config.EnableIptables, []string{"#iptables", "-iptables"}, true, "Enable Docker's addition of iptables rules")
	flag.BoolVar(&config.EnableIpForward, []string{"#ip-forward", "-ip-forward"}, true, "Enable net.ipv4.ip_forward")
	flag.StringVar(&config.BridgeIP, []string{"#bip", "-bip"}, "", "Use this CIDR notation address for the network bridge's IP, not compatible with -b")
	flag.StringVar(&config.BridgeIface, []string{"b", "-bridge"}, "", "Attach containers to a pre-existing network bridge\nuse 'none' to disable container networking")
	flag.BoolVar(&config.InterContainerCommunication, []string{"#icc", "-icc"}, true, "Enable inter-container communication")
	flag.StringVar(&config.GraphDriver, []string{"s", "-storage-driver"}, "", "Force the Docker runtime to use a specific storage driver")
	flag.StringVar(&config.ExecDriver, []string{"e", "-exec-driver"}, "native", "Force the Docker runtime to use a specific exec driver")
	flag.BoolVar(&config.EnableSelinuxSupport, []string{"-selinux-enabled"}, false, "Enable selinux support. SELinux does not presently support the BTRFS storage driver")
	flag.IntVar(&config.Mtu, []string{"#mtu", "-mtu"}, 0, "Set the containers network MTU\nif no value is provided: default to the default route MTU or 1500 if no default route is available")
	opts.IPVar(&config.DefaultIp, []string{"#ip", "-ip"}, "0.0.0.0", "Default IP address to use when binding container ports")
	opts.ListVar(&config.GraphOptions, []string{"-storage-opt"}, "Set storage driver options")
	// FIXME: why the inconsistency between "hosts" and "sockets"?
	opts.IPListVar(&config.Dns, []string{"#dns", "-dns"}, "Force Docker to use specific DNS servers")
	opts.DnsSearchListVar(&config.DnsSearch, []string{"-dns-search"}, "Force Docker to use specific DNS search domains")
}

func GetDefaultNetworkMtu() int {
	if iface, err := networkdriver.GetDefaultRouteIface(); err == nil {
		return iface.MTU
	}
	return defaultNetworkMtu
}

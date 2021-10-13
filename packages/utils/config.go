package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

func GetServeHost(defaultPort int16) string {
	ip := ""

	if os.Getenv("NOZ_DEV") == "1" {
		ip = "127.0.0.1"
	}

	portString := os.Getenv("PORT")
	if portString == "" {
		portString = ToString(defaultPort)
	}
	return ip + ":" + portString
}

func ResolveAuthServiceHost() *url.URL {
	u, _ := url.Parse("http://0.0.0.0")
	portString := fmt.Sprint(config.portsConfig[config.serviceConfig.Authorization])
	ip, err := ResolveIPByServiceName(config.serviceConfig.Authorization)
	if err != nil {
		panic(err)
	}
	u.Host = ip + ":" + portString

	return u
}
func ResolveRedisServiceHost() *url.URL {
	u, _ := url.Parse("tcp://127.0.0.1")
	name := config.serviceConfig.Redis
	portString := fmt.Sprint(config.portsConfig[name])
	ip, err := ResolveIPByServiceName(name)
	if err != nil {
		panic(err)
	}
	u.Host = ip + ":" + portString

	return u
}
func ResolveMySqlServiceHost() *url.URL {
	u, _ := url.Parse("tcp://127.0.0.1")
	name := config.serviceConfig.MySql
	portString := fmt.Sprint(config.portsConfig[name])
	ip, err := ResolveIPByServiceName(name)
	if err != nil {
		panic(err)
	}
	u.Host = ip + ":" + portString

	return u
}

func ResolveService(r *http.Request) (*url.URL, error) {
	paths := []string{}
	patternServiceMap := map[string]string{}
	for _, routeItem := range config.serviceConfig.Routes {
		name := ToString(routeItem.Key)
		patterns := ToStringSlice(routeItem.Value)
		paths = append(paths, patterns...)

		for _, pattern := range patterns {
			patternServiceMap[pattern] = name
		}
	}

	// matching

	matchedServiceName := ""
	for i := len(paths); i > 0; i-- {
		pattern := paths[i-1]
		segs := strings.Split(pattern, " ")

		var patternPath = ""
		var patternMethod uint16 = Http_method_any
		if len(segs) > 0 {
			patternPath = segs[0]
		}
		if len(segs) > 1 {
			patternMethod = ParseMethodTypeByTypeString(segs[1])
		}

		if patternPath == "" {
			continue
		}

		if patternMethod != Http_method_any && patternMethod != ParseMethodTypeByTypeString(r.Method) {
			continue
		}

		serviceName := patternServiceMap[pattern]
		if ok, _ := regexp.MatchString(patternPath, r.URL.Path); ok {
			matchedServiceName = serviceName
			break
		}
	}

	rewritePath := false

	// default backend service
	cu, _ := url.Parse("http://0.0.0.0")

	if matchedServiceName != "" {
		ip, err := ResolveIPByServiceName(matchedServiceName)
		if err != nil {
			return nil, err
		}
		fmt.Println("Mathched service name: ", matchedServiceName, ip)
		portString := ToString(config.portsConfig[matchedServiceName])
		cu.Host = ip + ":" + portString
		if rewritePath {
			cu.Path = r.URL.Path
		}
		return cu, nil
	}

	return cu, nil
}

func ResolveIPByServiceName(serviceName string) (string, error) {
	matchedIPPool := ResolveIP(serviceName)
	if len(matchedIPPool) == 0 {
		return "", fmt.Errorf("can't resolve service name (%s)", serviceName)
	}
	return matchedIPPool[0], nil
}

func IsURLAuthNeednt(u *url.URL) bool {
	paths := config.serviceConfig.NonAuth

	for i := len(paths); i > 0; i-- {
		if ok, _ := regexp.MatchString(paths[i-1], u.Path); ok {
			return true
		}
	}

	return false
}

func LoadConfig() Config {
	servicesConfigPath := "./service.yml"
	portsConfigPath := "./ports.yml"

	if os.Getenv("NOZ_DEV") == "1" {
		servicesConfigPath = "../../config/service.yml"
		portsConfigPath = "../../config/ports.yml"
	}

	config.serviceConfig = LoadServiceConfig(servicesConfigPath)
	config.portsConfig = LoadPortsConfig(portsConfigPath)

	fmt.Println("Config loaded: ", config)
	return config
}

type Config struct {
	serviceConfig ServiceConfig
	portsConfig   PortsConfig
}

type ServiceConfig struct {
	Authorization string
	Redis         string
	MySql         string
	NonAuth       []string `yaml:"non-auth"`
	Routes        yaml.MapSlice
}

type PortsConfig map[string]int16

var serviceConfig ServiceConfig
var portsConfig PortsConfig
var config = Config{}

var routesConfigLoaded = false
var portsConfigLoaded = false

func LoadServiceConfig(path string) ServiceConfig {
	if routesConfigLoaded {
		return serviceConfig
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(bytes, &serviceConfig)
	if err != nil {
		panic(err)
	}

	routesConfigLoaded = true
	return serviceConfig
}
func LoadPortsConfig(path string) PortsConfig {
	if portsConfigLoaded {
		return portsConfig
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(bytes, &portsConfig)
	if err != nil {
		panic(err)
	}

	portsConfigLoaded = true
	return portsConfig
}

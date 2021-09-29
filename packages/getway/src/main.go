package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"gopkg.in/yaml.v2"
	"noz.zkip.cc/utils"
)

type Config struct {
	routesConfig RoutesConfig
	portsConfig  PortsConfig
}

type RoutesConfig struct {
	Authorization string
	Routes        yaml.MapSlice
}

type PortsConfig map[string]int16

type ValidationAuth struct {
	ok    bool
	which string
}

var routeConfig RoutesConfig
var portsConfig PortsConfig
var config = Config{}

var routesConfigLoaded = false
var portsConfigLoaded = false

func loadRoutesConfig(path string) RoutesConfig {
	if routesConfigLoaded {
		return routeConfig
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(bytes, &routeConfig)
	if err != nil {
		panic(err)
	}

	routesConfigLoaded = true
	return routeConfig
}
func loadPortsConfig(path string) PortsConfig {
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

func needAuth(u *url.URL) bool {
	return true
}

func resolveService(r *http.Request) (*url.URL, error) {
	paths := []string{}
	patternServiceMap := map[string]string{}
	for _, routeItem := range config.routesConfig.Routes {
		name := utils.ToString(routeItem.Key)
		patterns := utils.ToStringSlice(routeItem.Value)
		paths = append(paths, patterns...)

		for _, pattern := range patterns {
			patternServiceMap[pattern] = name
		}
	}

	// matching

	matchedServiceName := ""
	for i := len(paths); i > 0; i-- {
		pattern := paths[i-1]
		serviceName := patternServiceMap[pattern]
		if ok, _ := regexp.MatchString(pattern, r.URL.Path); ok {
			matchedServiceName = serviceName
			break
		}
	}

	rewritePath := false

	// default backend service
	cu, _ := url.Parse("http://0.0.0.0")

	if matchedServiceName != "" {
		ip, err := resolveIPByServiceName(matchedServiceName)
		if err != nil {
			return nil, err
		}
		fmt.Println("Mathched service name: ", matchedServiceName, ip)
		portString := utils.ToString(config.portsConfig[matchedServiceName])
		cu.Host = ip + ":" + portString
		if rewritePath {
			cu.Path = r.URL.Path
		}
		return cu, nil
	}

	return cu, nil
}

func resolveIPByServiceName(serviceName string) (string, error) {
	matchedIPPool := utils.ResolveIP(serviceName)
	if len(matchedIPPool) == 0 {
		return "", fmt.Errorf("can't resolve service name (%s)", serviceName)
	}
	return matchedIPPool[0], nil
}

func getAuthService() *url.URL {
	u, _ := url.Parse("http://0.0.0.0")
	portString := fmt.Sprint(config.portsConfig[config.routesConfig.Authorization])
	ip, err := resolveIPByServiceName(config.routesConfig.Authorization)
	if err != nil {
		panic(err)
	}
	u.Host = ip + ":" + portString

	return u
}

func getValidation(r *http.Request, authService *url.URL) (ValidationAuth, error) {
	va := ValidationAuth{}

	// var err error
	// client := &http.Client{}

	// req, _ := http.NewRequest(r.Method, authService.String(), r.Body)
	// req.Header = r.Header.Clone()

	// resp, err := client.Do(req)
	// if err != nil {
	// 	return va, err
	// }

	// if resp.StatusCode != http.StatusOK {
	// 	return va, nil
	// }

	// result := &model.WhichResult{}
	// err = utils.ParseJsonFromResponse(resp, result)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(result)

	// va.which = result.Which
	// va.ok = true
	return va, nil
}

func getAuthWhich(va ValidationAuth) string {
	return "AX"
}

func loadConfig() Config {
	config.routesConfig = loadRoutesConfig("./config.yml")
	config.portsConfig = loadPortsConfig("./ports.yml")
	return config
}

func main() {

	loadConfig()
	fmt.Println("Config Loaded: ", config)

	fmt.Println("Server actived.")
	err := http.ListenAndServe(utils.GetServeHost(9000), http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		serviceURL, err := resolveService(r)
		if err != nil {
			fmt.Println(err)
			rw.WriteHeader(http.StatusBadGateway)
		}
		fmt.Println("Service URL: ", serviceURL, " Origin: ", r.URL)
		if needAuth(r.URL) {
			va, err := getValidation(r, getAuthService())
			if err != nil {
				fmt.Println(err)
			}

			if va.ok {
				proxy := httputil.NewSingleHostReverseProxy(serviceURL)
				rw.Header().Add("X-Authenticate-Which", getAuthWhich(va))
				proxy.ServeHTTP(rw, r)
			} else {
				rw.WriteHeader(http.StatusUnprocessableEntity)
			}
		} else {
			proxy := httputil.NewSingleHostReverseProxy(serviceURL)
			proxy.ServeHTTP(rw, r)
		}
	}))
	fmt.Println(err)
}

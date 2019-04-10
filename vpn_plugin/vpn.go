package vpn_plugin

import (
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/wssocks/client"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const USTBVpnHost = "n.ustb.edu.cn"
const USTBVpnHttpScheme = "http"
const USTBVpnLoginUrl = USTBVpnHttpScheme + "://" + USTBVpnHost + "/do-login"
const USTBVpnWSScheme = "ws"

type UstbVpn struct {
	enable    bool
	username  string
	password  string
	targetUrl string
}

// create a UstbVpn instance, and add necessary command options to client sub-command.
func NewUstbVpn() *UstbVpn {
	vpn := UstbVpn{}
	// add more command options for client sub-command.
	if ok, clientCmd := cmds.Find(client.CommandNameClient); ok {
		clientCmd.FlagSet.BoolVar(&vpn.enable, "vpn-enable", false, `enable USTB vpn feature.`)
		clientCmd.FlagSet.StringVar(&vpn.username, "vpn-username", "", `username to login vpn.`)
		clientCmd.FlagSet.StringVar(&vpn.password, "vpn-password", "", `password to login vpn.`)
		clientCmd.FlagSet.StringVar(&vpn.targetUrl, "vpn-login-url", USTBVpnLoginUrl, `address to login vpn.`)
	}
	return &vpn
}

func (v *UstbVpn) BeforeRequest(dialer *websocket.Dialer, url *url.URL, header http.Header) error {
	if !v.enable {
		return nil
	}
	// change target url.
	vpnUrl(url)
	// add cookie
	if cookies, err := vpnLogin(v.targetUrl, v.username, v.password); err != nil {
		return err
	} else {
		if jar, err := cookiejar.New(nil); err != nil {
			return err
		} else {
			dialer.Jar = jar
			cookieUrl := *url
			// replace url scheme "wss" to "https" and "ws"to "http"
			cookieUrl.Scheme = strings.Replace(cookieUrl.Scheme, "ws", "http", 1)
			dialer.Jar.SetCookies(&cookieUrl, cookies)
			return nil
		}
	}
}

func vpnUrl(u *url.URL) {
	// replace https://abc.com to "http://n.ustb.edu.cn/https/abc.com"
	// replace https://abc.com:8080 to "http://n.ustb.edu.cn/https-8080/abc.com"
	// ?wrdrecordrvisit=record

	// split host and port if it could
	port := u.Port()
	if strings.ContainsRune(u.Host, ':') {
		if h, p, err := net.SplitHostPort(u.Host); err != nil {
			panic(err)
		} else {
			u.Host = h
			if port != "" {
				port = p
			}
		}
	}

	schemeWithPort := u.Scheme
	if (u.Scheme == "wss" || u.Scheme == "https") && port != "" && port != "443" {
		schemeWithPort = u.Scheme + "-" + port
	}
	if (u.Scheme == "ws" || u.Scheme == "http") && port != "" && port != "80" {
		schemeWithPort = u.Scheme + "-" + port
	}

	u.Path = schemeWithPort + "/" + u.Host + u.Path
	u.Host = USTBVpnHost

	// set scheme
	if u.Scheme == "wss" || u.Scheme == "ws" {
		u.Scheme = USTBVpnWSScheme
	} else if u.Scheme == "https" {
		u.Scheme = USTBVpnHttpScheme
	} else {
		u.Scheme = USTBVpnHttpScheme
	}

	fmt.Println("real url:", u.String())
}

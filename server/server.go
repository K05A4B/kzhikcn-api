package server

import (
	"fmt"
	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/log"
	"net/http"
)

func Serve(addr string) error {
	conf := config.Conf()

	cert := conf.CertFile
	key := conf.KeyFile

	r := NewRouter()

	fmt.Println("Version:", appinfo.CurrentInfo.Version)
	fmt.Println("Author:", appinfo.CurrentInfo.Author)
	fmt.Println("Copyright:", appinfo.CurrentInfo.Copyright)

	if cert != "" && key != "" {
		log.Info("Starting HTTPS server on ", addr)
		return http.ListenAndServeTLS(addr, cert, key, r)
	}

	log.Info("Starting non-HTTPS server on ", addr)

	return http.ListenAndServe(addr, r)
}

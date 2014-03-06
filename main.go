package grunway

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

func Start(routerPtr *Router, host string) error {
	log.Println("Starting")
	log.Println("go v:", runtime.Version())
	log.Println("grunway v:", Version())

	if routerPtr.AllRoutesCount() == 0 {
		return fmt.Errorf("3621140790 Router has no valid routes defined")
	}

	log.Println("All Routes:")
	fmt.Print(routerPtr.AllRoutesSummary())
	log.Println("End Routes")

	log.Println("Listening from ", host)
	err := http.ListenAndServe(host, routerPtr)
	if err != nil {
		return fmt.Errorf("3621140792 http.ListenAndServe returned error:", err)
	}

	return nil
}

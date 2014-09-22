package grunway

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"
)

func Start(routerPtr *Router, host string) error {
	log.Printf("Starting Grunway Server (%v, grunway v%v(%v))", runtime.Version(), Version(), BuildNumber())

	if routerPtr.AllRoutesCount() == 0 {
		return fmt.Errorf("3621140790 Router has no valid routes defined")
	}

	log.Println("All Routes:")
	routerPtr.LogAllRoutes("↳ Route:")

	log.Println("Listening from ", host)
	time.Sleep(100 * time.Millisecond) // give the log statements time to print...

	err := http.ListenAndServe(host, routerPtr)
	if err != nil {
		return fmt.Errorf("3621140792 http.ListenAndServe returned error:", err)
	}

	return nil
}

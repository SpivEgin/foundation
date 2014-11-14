// Package "foundation" represents default Ottemo e-commerce product build.
//
// This package contains main() function which is the start point of assembled
// application, as well as this file declares an application components to use
// by void usage import of them. So, these void import packages are self-init
// packages are replaceable modules/extension/plugins.
//
// Example:
//   go build github.com/ottemo/foundation
//   go build -tags mongo github.com/ottemo/foundation
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ottemo/foundation/app"

	_ "github.com/ottemo/foundation/env/config"
	_ "github.com/ottemo/foundation/env/errors"
	_ "github.com/ottemo/foundation/env/events"
	_ "github.com/ottemo/foundation/env/ini"
	_ "github.com/ottemo/foundation/env/logger"

	_ "github.com/ottemo/foundation/media/fsmedia"

	_ "github.com/ottemo/foundation/api/rest"

	_ "github.com/ottemo/foundation/app/actors/cart"
	_ "github.com/ottemo/foundation/app/actors/category"
	_ "github.com/ottemo/foundation/app/actors/product"
	_ "github.com/ottemo/foundation/app/actors/visitor"
	_ "github.com/ottemo/foundation/app/actors/visitor/address"

	_ "github.com/ottemo/foundation/app/actors/checkout"
	_ "github.com/ottemo/foundation/app/actors/order"

	_ "github.com/ottemo/foundation/app/actors/payment/checkmo"
	_ "github.com/ottemo/foundation/app/actors/payment/paypal"

	_ "github.com/ottemo/foundation/app/actors/shipping/flat"
	_ "github.com/ottemo/foundation/app/actors/shipping/usps"

	_ "github.com/ottemo/foundation/app/actors/discount"
	_ "github.com/ottemo/foundation/app/actors/tax"

	_ "github.com/ottemo/foundation/app/actors/product/review"

	_ "github.com/ottemo/foundation/app/actors/cms"
	_ "github.com/ottemo/foundation/app/actors/rts"
	_ "github.com/ottemo/foundation/app/actors/seo"

	_ "github.com/ottemo/foundation/app/actors/payment/authorize"
	_ "github.com/ottemo/foundation/app/actors/shipping/fedex"

	_ "github.com/ottemo/foundation/impex"
)

// executable file start point
func main() {
	defer app.End() // application close event

	// we should intercept os signals to application as we should call app.End() before
	signalChain := make(chan os.Signal, 1)
	signal.Notify(signalChain, os.Interrupt, syscall.SIGTERM)
	go func() {
		for _ = range signalChain {
			err := app.End()
			if err != nil {
				fmt.Println(err.Error())
			}

			os.Exit(0)
		}
	}()

	// application start event
	if err := app.Start(); err != nil {
		fmt.Println(err.Error())
	}

	// starting HTTP server
	app.Serve()
}

package app

import (
	"runtime"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {
	var err error

	err = api.GetRestService().RegisterAPI("app", "GET", "login", restLogin)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("app", "POST", "login", restLogin)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("app", "GET", "logout", restLogout)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("app", "GET", "rights", restRightsInfo)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("app", "GET", "status", restStatusInfo)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// WEB REST API function login application with root rights
func restLogin(params *api.StructAPIHandlerParams) (interface{}, error) {

	var requestLogin string
	var requestPassword string

	if params.Request.Method == "GET" && utils.KeysInMapAndNotBlank(params.RequestGETParams, "login", "password") {
		requestLogin = params.RequestGETParams["login"]
		requestPassword = params.RequestGETParams["password"]

	} else {

		reqData, err := api.GetRequestContentAsMap(params)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		if !utils.KeysInMapAndNotBlank(reqData, "login", "password") {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "fee28a56-adb1-44b9-a0e2-1c9be6bd6fdb", "login and password should be specified")
		}

		requestLogin = utils.InterfaceToString(reqData["login"])
		requestPassword = utils.InterfaceToString(reqData["password"])
	}

	rootLogin := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathStoreRootLogin))
	rootPassword := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathStoreRootPassword))

	if requestLogin == rootLogin && requestPassword == rootPassword {
		params.Session.Set(api.ConstSessionKeyAdminRights, true)

		return "ok", nil
	}

	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "68546aa8-a6be-4c31-ac44-ea4278dfbdb0", "wrong login or password")
}

// WEB REST API function logout application - session data clear
func restLogout(params *api.StructAPIHandlerParams) (interface{}, error) {
	err := params.Session.Close()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	return "ok", nil
}

// WEB REST API function to get info about current rights
func restRightsInfo(params *api.StructAPIHandlerParams) (interface{}, error) {
	result := make(map[string]interface{})

	result["is_admin"] = utils.InterfaceToBool(params.Session.Get(api.ConstSessionKeyAdminRights))

	return result, nil
}

// WEB REST API function to get info about current rights
func restStatusInfo(params *api.StructAPIHandlerParams) (interface{}, error) {
	result := make(map[string]interface{})

	result["Version"] = runtime.Version()
	result["NumGoroutine"] = runtime.NumGoroutine()
	result["NumCPU"] = runtime.NumCPU()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// General statistics.
	result["memStats.Alloc"] = memStats.Alloc           // bytes allocated and still in use
	result["memStats.TotalAlloc"] = memStats.TotalAlloc // bytes allocated (even if freed)
	result["memStats.Sys"] = memStats.Sys               // bytes obtained from system (sum of XxxSys below)
	result["memStats.Lookups"] = memStats.Lookups       // number of pointer lookups
	result["memStats.Mallocs"] = memStats.Mallocs       // number of mallocs
	result["memStats.Frees"] = memStats.Frees           // number of frees

	// Main allocation heap statistics.
	result["memStats.HeapAlloc"] = memStats.HeapAlloc       // bytes allocated and still in use
	result["memStats.HeapSys"] = memStats.HeapSys           // bytes obtained from system
	result["memStats.HeapIdle"] = memStats.HeapIdle         // bytes in idle spans
	result["memStats.HeapInuse"] = memStats.HeapInuse       // bytes in non-idle span
	result["memStats.HeapReleased"] = memStats.HeapReleased // bytes released to the OS
	result["memStats.HeapObjects"] = memStats.HeapObjects   // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	// (Inuse is bytes used now.; Sys is bytes obtained from system.)
	result["memStats.StackInuse"] = memStats.StackInuse // bytes used by stack allocator
	result["memStats.StackSys"] = memStats.StackSys
	result["memStats.MSpanInuse"] = memStats.MSpanInuse // mspan structures
	result["memStats.MSpanSys"] = memStats.MSpanSys
	result["memStats.MCacheInuse"] = memStats.MCacheInuse // mcache structures
	result["memStats.MCacheSys"] = memStats.MCacheSys
	result["memStats.BuckHashSys"] = memStats.BuckHashSys // profiling bucket hash table
	result["memStats.GCSys"] = memStats.GCSys             // GC metadata
	result["memStats.OtherSys"] = memStats.OtherSys       // other system allocations

	// Garbage collector statistics.
	result["memStats.NextGC"] = memStats.NextGC // next collection will happen when HeapAlloc ≥ this amount
	result["memStats.LastGC"] = memStats.LastGC // end time of last collection (nanoseconds since 1970)
	result["memStats.PauseTotalNs"] = memStats.PauseTotalNs
	result["memStats.NumGC"] = memStats.NumGC
	result["memStats.EnableGC"] = memStats.EnableGC
	result["memStats.DebugGC"] = memStats.DebugGC

	return result, nil
}

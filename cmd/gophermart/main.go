package main

import "github.com/zasuchilas/gophermart/internal/gophermart"

func main() {
	//config.ParseFlags()
	//log.Println("RUN_ADDRESS", config.RunAddress)
	//
	//logger.Init()
	//logger.ServiceInfo("GOPHERMART (... service)", "TEST VERSION")
	//
	//store := pgstorage.New()
	//log.Println("STORE", store.InstanceName())
	//
	////chisrv.InitJWT()
	////log.Println("INIT JWT OK!")
	//
	//log.Println("HELLO GOPHERMART!")
	//helloHandler := func(w http.ResponseWriter, req *http.Request) {
	//	log.Println("GET: HELLO GOPHERMART!")
	//	io.WriteString(w, "HELLO GOPHERMART!\n")
	//}
	//http.HandleFunc("/", helloHandler)
	//log.Fatal(http.ListenAndServe(config.RunAddress, nil))
	a := gophermart.New()
	a.Run()
}

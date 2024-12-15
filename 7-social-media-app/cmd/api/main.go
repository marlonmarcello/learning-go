package main

import "log"

func main() {
	cfg := apiConfig{
		addr: ":8080",
	}

	app := &api{
		config: cfg,
	}

	mux := app.mount()

	log.Fatal(app.serve(mux))
}

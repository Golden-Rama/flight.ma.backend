package main

import (
    "fmt"
    "gr-flight-ma-new/config"
    "gr-flight-ma-new/src/factory"
)

func main() {
    cfg := config.Get()
    res, err := factory.NewResolver(cfg)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    providers, err := res.ProviderRepo.FindAll(true)
    if err != nil {
        fmt.Println("Provider error:", err)
        return
    }

    fmt.Printf("Active providers count: %d\n", len(providers))
    for _, p := range providers {
        fmt.Printf("Provider: %s (IsActive: %v, Domestic: %v) - URL: %s\n", p.Code, p.IsActive, p.AvailableDomestic, p.BaseUrl)
		if p.FlightQuest != nil {
			fmt.Printf("  Quest Endpoint: %s\n", p.FlightQuest.Endpoint)
		} else {
			fmt.Println("  Quest is nil")
		}
    }
}

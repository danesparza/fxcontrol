// Package model defines the FX services returned by discovery.
package model

// Service describes an FX API advertised on the local network.
type Service struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Service   string   `json:"service"`
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	Addresses []string `json:"addresses"`
	API       string   `json:"api"`
	Scheme    string   `json:"scheme"`
	Path      string   `json:"path"`
}

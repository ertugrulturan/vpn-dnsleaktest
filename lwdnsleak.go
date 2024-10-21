package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const API_URL = "bash.ws"

type DnsData struct {
	IP        string `json:"ip"`
	Country   string `json:"country"`
	ASN       string `json:"asn"`
	TypeField string `json:"type"`
}

func testDnsLeak() ([]DnsData, error) {
	var data []DnsData

	resp, err := http.Get(fmt.Sprintf("https://%s/id", API_URL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	idBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	id := string(idBody)

	for i := 0; i < 10; i++ {
		_, _ = http.Get(fmt.Sprintf("https://%d.%s.%s", i, id, API_URL))
	}

	dnsResp, err := http.Get(fmt.Sprintf("https://%s/dnsleak/test/%s?json", API_URL, id))
	if err != nil {
		return nil, err
	}
	defer dnsResp.Body.Close()

	dnsBody, err := ioutil.ReadAll(dnsResp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(dnsBody, &data)
	if err != nil {
		return nil, err
	}
	var cleanedData []DnsData
	for _, dns := range data {
		if dns.TypeField == "dns" && dns.Country != "" {
			cleanedData = append(cleanedData, dns)
		}
	}

	return cleanedData, nil
}

type TraceData struct {
	Summary string
	Hops    []Hop
}

type Hop struct {
	TTL     string
	Host    string
	Address string
	Samples string
}

func traceroute(hostname string) (TraceData, error) {
	var traceData TraceData

	cmd := exec.Command("traceroute", hostname)
	output, err := cmd.Output()
	if err != nil {
		return traceData, err
	}

	lines := strings.Split(string(output), "\n")
	traceData.Summary = fmt.Sprintf("Traceroute to %s", hostname)

	for _, line := range lines {
		if strings.Contains(line, "ms") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				hop := Hop{
					TTL:     parts[0],
					Host:    parts[1],
					Address: parts[2],
					Samples: parts[3],
				}
				traceData.Hops = append(traceData.Hops, hop)
			}
		}
	}

	return traceData, nil
}

func logOutput(dnsData []DnsData, traceData TraceData) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(currentTime, "DNS Leak Test Results:")
	for _, dns := range dnsData {
		fmt.Printf("IP: %s, Country: %s, ASN: %s\n", dns.IP, dns.Country, dns.ASN)
	}

	fmt.Println()
	fmt.Println(currentTime, "Traceroute Results:")
	fmt.Println(traceData.Summary)
	for _, hop := range traceData.Hops {
		fmt.Printf("TTL: %s, Host: %s, Address: %s, Samples: %s\n", hop.TTL, hop.Host, hop.Address, hop.Samples)
	}
}
func main() {
	fmt.Println("Collecting DNS leak test data...")
	dnsData, err := testDnsLeak()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Running traceroute...")
	traceData, err := traceroute("discord.com")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	logOutput(dnsData, traceData)
}

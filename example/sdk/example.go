package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	eagleeye "github.com/EscapeBearSecond/falcon/pkg/sdk"
	"github.com/EscapeBearSecond/falcon/pkg/types"
)

func main() {
	stdlog := log.New(os.Stderr, "", log.LstdFlags)

	options := &types.Options{
		Targets: []string{
			"192.168.1.0-192.168.1.255",
		},
		ExcludeTargets: []string{
			"192.168.1.108",
		},
		PortScanning: types.PortScanningOptions{
			Use:         true,
			Timeout:     "5s",
			Count:       1,
			Format:      "csv",
			Ports:       "http",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		HostDiscovery: types.HostDiscoveryOptions{
			Use:         true,
			Timeout:     "5s",
			Count:       1,
			Format:      "csv",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		Jobs: []types.JobOptions{
			{
				Name:        "漏洞扫描",
				Kind:        "vul-scan",
				Template:    "./templates/漏洞扫描",
				Format:      "csv",
				Count:       1,
				Timeout:     "5s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
			{
				Name: "资产扫描",
				Kind: "asset-scan",
				GetTemplates: func() []*types.RawTemplate {
					var templates []*types.RawTemplate
					templates = append(templates, &types.RawTemplate{
						ID: "pgsql-detect",
						Original: `id: pgsql-detect

info:
  name: PostgreSQL Authentication - Detect
  author: nybble04,geeknik
  severity: info
  description: |
    PostgreSQL authentication error messages which could reveal information useful in formulating further attacks were detected.
  reference:
    - https://www.postgresql.org/docs/current/errcodes-appendix.html
    - https://www.postgresql.org/docs/current/client-authentication-problems.html
  classification:
    cvss-metrics: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N
    cwe-id: CWE-200
  metadata:
    max-request: 1
    shodan-query: port:5432 product:"PostgreSQL"
    verified: true
  tags: network,postgresql,db,detect

tcp:
  - inputs:
      - data: "000000500003000075736572006e75636c6569006461746162617365006e75636c6569006170706c69636174696f6e5f6e616d65007073716c00636c69656e745f656e636f64696e6700555446380000"
        type: hex
      - data: "7000000036534352414d2d5348412d32353600000000206e2c2c6e3d2c723d000000000000000000000000000000000000000000000000"
        type: hex

    host:
      - "{{Hostname}}"
    port: 5432
    read-size: 2048

    matchers-condition: and
    matchers:
      - type: word
        part: body
        words:
          - "C0A000"                  # Error code for unsupported frontend protocol
          - "C08P01"                  # Error code for invalide startup packet layout
          - "28000"                   # Error code for invalid_authorization_specification
          - "28P01"                   # Error code for invalid_password
          - "SCRAM-SHA-256"           # Authentication prompt
          - "pg_hba.conf"             # Client authentication config file
          - "user \"nuclei\""         # The user nuclei (sent in request) doesn't exist
          - "database \"nuclei\""     # The db nuclei (sent in request) doesn't exist"
        condition: or

      - type: word
        words:
          - "HTTP/1.1"
        negative: true
# digest: 4a0a004730450220190550562f0223183090e8ca4117ace44d725bdece7b84c58edaed8d93935aa7022100872d6d635b69589e7e99749cae0639a48551bbdee2d3d7038aa4699257a00383:922c64590222798bb761d5b6d8e72950`,
					})
					return templates
				},
				Format:      "csv",
				Count:       1,
				Timeout:     "5s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
		},
	}

	engine, err := eagleeye.NewEngine(eagleeye.WithDirectory("./results"))
	if err != nil {
		stdlog.Fatalln(err)
	}
	defer engine.Close()

	{
		entry, err := engine.NewEntry(options)
		if err != nil {
			stdlog.Fatalln("error:", err)
		}

		stage := entry.Stage()
		stdlog.Printf("stage: %#v\n", stage)

		stop := make(chan struct{})
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					stdlog.Printf("stage: %#v\n", entry.Stage())
				case <-stop:
					return
				}
			}
		}()

		err = entry.Run(context.Background())
		if err != nil {
			stdlog.Fatalln("error:", err)
		}

		close(stop)

		stage = entry.Stage()
		stdlog.Printf("stage: %#v\n", stage)

		ret := entry.Result()
		fmt.Printf("result: %#v\n", ret)
	}

	{
		entry, err := engine.NewEntry(options)
		if err != nil {
			stdlog.Fatalln("error:", err)
		}

		go func() {
			<-time.After(10 * time.Second)
			entry.Stop()
		}()

		err = entry.Run(context.Background())
		if err != nil {
			stdlog.Fatalln("error:", err)
		}
	}
}

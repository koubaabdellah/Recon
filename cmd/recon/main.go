package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var defaultSources = []string{
	"crt.sh",
	"anubis",
}

func main() {
	domain := flag.String("d", "", "single domain")
	file := flag.String("f", "", "file with domains")
	output := flag.String("o", "subdomains.txt", "output file")
	threads := flag.Int("t", 10, "number of workers")
	flag.Parse()

	domains := make([]string, 0)
	if *domain != "" {
		domains = append(domains, strings.TrimSpace(*domain))
	}
	if *file != "" {
		list, err := readLines(*file)
		if err != nil {
			fatal(err)
		}
		domains = append(domains, list...)
	}
	if len(domains) == 0 {
		fatal(fmt.Errorf("use -d or -f"))
	}

	jobs := make(chan string)
	results := make(chan string, 1024)
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 12 * time.Second}
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for d := range jobs {
				for _, s := range collect(client, d) {
					results <- s
				}
			}
		}()
	}

	go func() {
		for _, d := range domains {
			jobs <- d
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	set := map[string]struct{}{}
	for r := range results {
		set[r] = struct{}{}
	}

	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)

	if err := writeLines(*output, out); err != nil {
		fatal(err)
	}

	fmt.Printf("saved %d subdomains to %s\n", len(out), *output)
}

func collect(client *http.Client, domain string) []string {
	results := map[string]struct{}{}
	for _, source := range defaultSources {
		var found []string
		switch source {
		case "crt.sh":
			found = fromCRTSh(client, domain)
		case "anubis":
			found = fromAnubis(client, domain)
		}
		for _, s := range found {
			results[s] = struct{}{}
		}
	}
	out := make([]string, 0, len(results))
	for s := range results {
		out = append(out, s)
	}
	return out
}

func fromCRTSh(client *http.Client, domain string) []string {
	url := "https://crt.sh/?q=%25." + domain + "&output=json"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "recon/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	return extractDomains(resp, domain)
}

func fromAnubis(client *http.Client, domain string) []string {
	url := "https://tls.bufferover.run/dns?q=." + domain
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "recon/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	return extractDomains(resp, domain)
}

func extractDomains(resp *http.Response, domain string) []string {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	set := map[string]struct{}{}
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		parts := strings.FieldsFunc(line, func(r rune) bool {
			return !(r == '.' || r == '-' || r == '_' || r == '*' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
		})
		for _, p := range parts {
			p = strings.TrimPrefix(p, "*.")
			if strings.HasSuffix(p, domain) && net.ParseIP(p) == nil {
				set[p] = struct{}{}
			}
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	return out
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" {
			out = append(out, line)
		}
	}
	return out, s.Err()
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, line := range lines {
		_, err := w.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return w.Flush()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

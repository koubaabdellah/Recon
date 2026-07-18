# Recon

Go tool to collect subdomains for domains you are authorized to test.

## Build
```bash
go build -o recon ./cmd/recon
```

## Run
Single domain:
```bash
./recon -d example.com -o subdomains.txt
```

From file:
```bash
./recon -f domains.txt -o subdomains.txt
```

## Output
The tool writes a deduplicated, sorted list to the output file.

## Linux
```bash
chmod +x recon
./recon -d example.com -o out.txt
```

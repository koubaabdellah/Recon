# Recon

Recon is a small Go-based CLI for collecting authorized target hosts, resolving them, and checking which ones are alive over HTTP.

## Features
- Collect domains from one or more input files
- Resolve DNS names
- Check HTTP responsiveness
- Write normalized output files
- Easy Linux install script

## Build
```bash
make build
```

## Run
Single domain:
```bash
./bin/recon -d example.com
```

From file:
```bash
./bin/recon -f domains.txt
```

## Install on Linux
```bash
chmod +x scripts/install.sh
./scripts/install.sh
```

Then run:
```bash
recon -d example.com
```

## Output
Results are written to the `output/` directory:
- `all.txt`
- `resolved.txt`
- `alive.txt`
- `report.json`

## Notes
Use this only on domains and assets you are authorized to test.

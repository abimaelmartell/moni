TAILWIND_VERSION := 3.3.4

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
  ifeq ($(UNAME_M),x86_64)
    PLATFORM := linux-x64
  else
    PLATFORM := linux-arm64
  endif
else ifeq ($(UNAME_S),Darwin)
  ifeq ($(UNAME_M),arm64)
    PLATFORM := macos-arm64
  else
    PLATFORM := macos-x64
  endif
endif

TAILWIND_CLI := bin/tailwindcss
TAILWIND_URL := https://github.com/tailwindlabs/tailwindcss/releases/download/v$(TAILWIND_VERSION)/tailwindcss-$(PLATFORM)

OUTPUT_CSS   := static/assets/tailwind.min.css

CONTENT := "./static/index.html,./static/assets/main.js"

OUTPUT_BIN := bin/moni

.PHONY: all build cli clean

all: build

cli: $(TAILWIND_CLI)

$(TAILWIND_CLI):
	@echo "👉  Downloading Tailwind CLI v$(TAILWIND_VERSION) for $(PLATFORM)…"
	@mkdir -p $(dir $@)
	@curl -fsSL $(TAILWIND_URL) -o $@
	@chmod +x $@

$(OUTPUT_CSS): cli
	@echo "👉  Building Tailwind CSS → $@"
	@mkdir -p $(dir $@)
	@printf "@tailwind base;\n@tailwind components;\n@tailwind utilities;\n" | \
	  $(TAILWIND_CLI) \
	    -i - \
	    --content $(CONTENT) \
	    --output $@ \
	    --minify

build: $(OUTPUT_CSS) $(OUTPUT_BIN)

$(OUTPUT_BIN):
	@echo "👉  Building moni → $@"
	@mkdir -p $(dir $@)
	@go build -o $@ main.go

# linux build

OUTPUT_BIN_LINUX := $(OUTPUT_BIN)-linux

$(OUTPUT_BIN_LINUX):
	@echo "👉  Building moni for Linux → $(OUTPUT_BIN_LINUX)"
	@mkdir -p $(dir $@)
	@GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_BIN_LINUX) main.go

# darwin build

OUTPUT_BIN_DARWIN := $(OUTPUT_BIN)-darwin

$(OUTPUT_BIN_DARWIN):
	@echo "👉  Building moni for Darwin → $(OUTPUT_BIN_DARWIN)"
	@mkdir -p $(dir $@)
	@GOOS=darwin GOARCH=amd64 go build -o $(OUTPUT_BIN_DARWIN) main.go

# windows build

OUTPUT_BIN_WINDOWS := $(OUTPUT_BIN)-windows.exe

$(OUTPUT_BIN_WINDOWS):
	@echo "👉  Building moni for Windows → $(OUTPUT_BIN_WINDOWS)"
	@mkdir -p $(dir $@)
	@GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_BIN_WINDOWS) main.go

release: $(OUTPUT_CSS) $(OUTPUT_BIN_LINUX) $(OUTPUT_BIN_DARWIN) $(OUTPUT_BIN_WINDOWS)

clean:
	@rm -f bin/*

# Makefile – Download Tailwind CLI and build minimal CSS via stdin

# Tailwind CLI version
TAILWIND_VERSION := 3.3.4

# Detect OS and ARCH for the correct binary
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

# adjust this to wherever you want your CSS
OUTPUT_CSS   := static/assets/tailwind.min.css

# Files to scan for class usage
CONTENT := "./static/index.html,./static/assets/main.js"

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

build: $(OUTPUT_CSS)

clean:
	@rm -f $(TAILWIND_CLI) $(OUTPUT_CSS)

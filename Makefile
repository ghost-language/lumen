.PHONY: build release run clean examples fmt vet check

# Build for the host platform. Lumen links against SDL2 through cgo, so a build
# always targets the machine it runs on unless a cross-compiler is set up.
build: clean
	go build -o dist/lumen ./cmd

# A stripped release build.
release: clean
	go build -ldflags "-s -w" -o dist/lumen ./cmd

# make run EXAMPLE=60_rpg
EXAMPLE ?= 60_rpg

run: build
	./dist/lumen examples/$(EXAMPLE)

# Start every example in turn for a few seconds, reporting any that fail. Useful
# after changing the module layer.
examples: build
	@for dir in examples/*/; do \
		name=$$(basename $$dir); \
		[ -f "$$dir/main.ghost" ] || continue; \
		printf '%-24s' "$$name"; \
		output=$$(SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy timeout 3 ./dist/lumen "$$dir" 2>&1 | grep -Ei 'error|syntax' | head -3); \
		if [ -n "$$output" ]; then echo "FAIL"; echo "$$output"; else echo "ok"; fi; \
	done

# Mirrors Ghost's Makefile so CI and a local check run the same thing.
fmt:
	@test -z "$$(gofmt -l . | tee /dev/stderr)" || (echo "run gofmt -w ." && exit 1)

vet:
	go vet ./...

check: fmt vet examples

clean:
	@rm -rf dist/lumen dist/lumen.exe

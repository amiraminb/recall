BINARY := recall
CMD := ./cmd/recall
BIN_DIR := ./bin
BIN_PATH := $(BIN_DIR)/$(BINARY)

LOCAL_BIN ?= $(HOME)/.local/bin
ZSH_COMPLETION_DIR ?= $(HOME)/.config/zsh/completions

.PHONY: build install completion clean

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) $(CMD)

install: build
	@mkdir -p $(LOCAL_BIN)
	cp $(BIN_PATH) $(LOCAL_BIN)/$(BINARY)

completion: build
	@mkdir -p $(ZSH_COMPLETION_DIR)
	$(BIN_PATH) completion zsh > $(ZSH_COMPLETION_DIR)/_$(BINARY)

clean:
	rm -rf $(BIN_DIR)

APPS := fd rg

ARCHS := x86_64 aarch64

# Default target: build all apps for all architectures
all: $(foreach app,$(APPS),$(foreach arch,$(ARCHS),$(app)-$(arch)))

define BUILD_APP_RULE
TARGET_EXE_PATH_FOR_RULE := bin/$(2)/$(1)
SOURCE_FILE_FOR_RULE := $(1).zig

$(1)-$(2): $$(TARGET_EXE_PATH_FOR_RULE)

$$(TARGET_EXE_PATH_FOR_RULE): $$(SOURCE_FILE_FOR_RULE)
	@echo "Building $(1) for $(2) (source: $$<, target: $$@)..."
	@mkdir -p $$(dir $$@)
	zig build-exe $$< \
		-target $(2)-linux \
		-O ReleaseSmall \
		-fstrip \
		--name $(1) \
		-static \
		-femit-bin=$$@
	@rm $$@.o
	@echo "Finished building $(1) for $(2). Output: $$@"
endef

$(foreach app,$(APPS),$(foreach arch,$(ARCHS),$(eval $(call BUILD_APP_RULE,$(app),$(arch)))))

clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf bin
	@echo "Clean finished."

.PHONY: all clean $(foreach app,$(APPS),$(foreach arch,$(ARCHS),$(app)-$(arch)))

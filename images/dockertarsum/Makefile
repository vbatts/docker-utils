
DOCKER			:= docker
APP_NAME		:= dockertarsum
RUNTIME_NAME		:= $(USER)/$(APP_NAME):latest
BUILDER_DOCKERFILE	:= ./Dockerfile.build
BUILDER_NAME		:= $(USER)/$(APP_NAME):build

# Pass DEFAULT_KEY to override this
DEFAULT_KEY		:= 733F1362


default: runtime

$(APP_NAME): $(BUILDER_DOCKERFILE)
	cat $(BUILDER_DOCKERFILE) | $(DOCKER) build -t $(BUILDER_NAME) - && \
		$(DOCKER) run -v $(shell pwd)/:/output/ $(BUILDER_NAME) cp /gopath/bin/$(APP_NAME) /output/$(APP_NAME) && \
		echo "built $(BUILDER_NAME)"

$(APP_NAME).sum: $(APP_NAME)
	sha256sum $< > $@

$(APP_NAME).asc: $(APP_NAME)
	rm -f $@ && gpg --default-key $(DEFAULT_KEY) --armor --detach-sign $<

runtime: $(APP_NAME).sum $(APP_NAME).asc
	$(DOCKER) build -t $(RUNTIME_NAME) .

check: runtime
	$(DOCKER) save $(RUNTIME_NAME) | $(DOCKER) run -i $(RUNTIME_NAME)

clean:
	rm -rf $(APP_NAME) *~ $(wildcard *.sum) $(wildcard *.asc)


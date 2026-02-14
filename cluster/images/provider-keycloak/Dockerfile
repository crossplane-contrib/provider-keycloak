FROM gcr.io/distroless/static@sha256:d90359c7a3ad67b3c11ca44fd5f3f5208cbef546f2e692b0dc3410a869de46bf

ARG TARGETOS
ARG TARGETARCH

ADD bin/$TARGETOS\_$TARGETARCH/provider /usr/local/bin/provider

USER 65532
ENTRYPOINT ["provider"]
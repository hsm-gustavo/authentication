# referência: https://github.com/dreamsofcode-io/guestbook/blob/d00a11e5353974807524bbc7e5a567ffa382460b/Dockerfile

# variável para ser usada em build time
ARG GO_VERSION=1.26.3
# isso só funciona com buildx
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder
LABEL org.opencontainers.image.source="https://github.com/hsm-gustavo/authentication"
WORKDIR /src

# cache dos módulos do Go para acelerar builds subsequentes
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# tambám só funciona com buildx, a variável TARGETARCH é injetada automaticamente para indicar a arquitetura de destino (amd64, arm64, etc)
ARG TARGETARCH

# build do binário, usando cache para os módulos e montando o código fonte para dentro do container. CGO_ENABLED=0 garante que o binário seja estático, o que facilita a execução em ambientes variados sem depender de bibliotecas nativas.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=${TARGETARCH} go build -o /bin/server ./cmd/server

FROM alpine:3.23.4 AS final

# como é multistage o label deve ser adicionado em cada estágio
LABEL org.opencontainers.image.source="https://github.com/hsm-gustavo/authentication"

# instalação de dependências necessárias para rodar o binário, como certificados SSL e timezone data. O cache do apk é usado para acelerar instalações subsequentes, e a limpeza do cache é feita automaticamente pelo apk.
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

# cria um usuário não-root para rodar o servidor
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

# aqui copiamos quaisquer arquivos necessários para rodar o binário (incluindo o próprio binário)
COPY --from=builder /bin/server /bin/

EXPOSE 8080

ENTRYPOINT [ "/bin/server" ]
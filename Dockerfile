FROM ubuntu:20.04 AS base
RUN apt-get update && apt-get install -y curl

FROM base AS build
RUN curl -O https://example.com/app

FROM build AS final
CMD ["./app"]


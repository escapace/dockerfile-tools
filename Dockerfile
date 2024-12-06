FROM asd
RUN --mount=type=cache,target=./target \
  --mount=type=cache,target=/usr/local/cargo/git \
  --mount=type=cache,target=/usr/local/cargo/registry \
  cargo zigbuild --bin puffin --target $(cat rust_target.txt) --release

# Copy binary into normal layer
RUN --mount=type=cache,target=./target \
  cp ./target/$(cat rust_target.txt)/release/puffin /puffin

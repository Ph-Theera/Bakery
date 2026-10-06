# Packaging only, matching shipment-service: make build creates the binary.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --chmod=0555 build/bakery /app/bakery
USER nonroot
ENV HTTP_ADDR=0.0.0.0:8080
EXPOSE 8080
ENTRYPOINT ["/app/bakery"]
